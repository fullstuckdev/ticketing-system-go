package service

import (
	"errors"
	"ticketing-system/entity"
	"ticketing-system/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	Register(req *entity.RegisterRequest) (*entity.User, error)
	Login(req *entity.LoginRequest) (*entity.LoginResponse, error)
	GetProfile(userID string) (*entity.User, error)
	UpdateProfile(userID string, user *entity.User) (*entity.User, error)
	GetAllUsers(pagination *entity.Pagination, search *entity.Search) ([]entity.User, *entity.PaginationMeta, error)
	DeleteUser(userID string) error
	GenerateJWT(user *entity.User) (string, error)
	ValidateJWT(tokenString string) (*entity.User, error)
}

type userService struct {
	userRepo  repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) UserService {
	return &userService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *userService) Register(req *entity.RegisterRequest) (*entity.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &entity.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		Role:     entity.RoleUser,
		IsActive: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(req *entity.LoginRequest) (*entity.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.GenerateJWT(user)
	if err != nil {
		return nil, err
	}

	return &entity.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *userService) GetProfile(userID string) (*entity.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *userService) UpdateProfile(userID string, updateData *entity.User) (*entity.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Update only allowed fields
	if updateData.Name != "" {
		user.Name = updateData.Name
	}
	if updateData.Email != "" && updateData.Email != user.Email {
		// Check if new email is already taken
		existingUser, err := s.userRepo.GetByEmail(updateData.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existingUser != nil {
			return nil, errors.New("email already taken")
		}
		user.Email = updateData.Email
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetAllUsers(pagination *entity.Pagination, search *entity.Search) ([]entity.User, *entity.PaginationMeta, error) {
	users, total, err := s.userRepo.GetAll(pagination, search)
	if err != nil {
		return nil, nil, err
	}

	meta := &entity.PaginationMeta{
		CurrentPage: pagination.Page,
		TotalItems:  total,
		Limit:       pagination.GetLimit(),
		TotalPages:  int((total + int64(pagination.GetLimit()) - 1) / int64(pagination.GetLimit())),
	}

	return users, meta, nil
}

func (s *userService) DeleteUser(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Prevent admin from deleting themselves
	if user.Role == entity.RoleAdmin {
		return errors.New("cannot delete admin user")
	}

	return s.userRepo.Delete(userID)
}

func (s *userService) GenerateJWT(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(s.jwtExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *userService) ValidateJWT(tokenString string) (*entity.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	return user, nil
} 