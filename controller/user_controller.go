package controller

import (
	"net/http"
	"ticketing-system/entity"
	"ticketing-system/middleware"
	"ticketing-system/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body entity.RegisterRequest true "Registration data"
// @Success 201 {object} entity.Response{data=entity.User}
// @Failure 400 {object} entity.Response
// @Failure 409 {object} entity.Response
// @Router /register [post]
func (uc *UserController) Register(c *gin.Context) {
	var req entity.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	user, err := uc.userService.Register(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "email already registered" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Registration failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, entity.Response{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}

// Login godoc
// @Summary Login user
// @Description Login with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body entity.LoginRequest true "Login credentials"
// @Success 200 {object} entity.Response{data=entity.LoginResponse}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Router /login [post]
func (uc *UserController) Login(c *gin.Context) {
	var req entity.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	response, err := uc.userService.Login(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid email or password" || err.Error() == "account is deactivated" {
			statusCode = http.StatusUnauthorized
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Login failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user profile
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} entity.Response{data=entity.User}
// @Failure 401 {object} entity.Response
// @Failure 404 {object} entity.Response
// @Router /profile [get]
func (uc *UserController) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, entity.Response{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	user, err := uc.userService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, entity.Response{
			Success: false,
			Message: "User not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    user,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update current user profile
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entity.User true "User update data"
// @Success 200 {object} entity.Response{data=entity.User}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Failure 409 {object} entity.Response
// @Router /profile [put]
func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, entity.Response{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	var updateData entity.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	user, err := uc.userService.UpdateProfile(userID, &updateData)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "email already taken" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Profile update failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Profile updated successfully",
		Data:    user,
	})
}

// GetAllUsers godoc
// @Summary Get all users (Admin only)
// @Description Get list of all users with pagination and search
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param q query string false "Search query"
// @Success 200 {object} entity.PaginatedResponse{data=[]entity.User}
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Router /users [get]
func (uc *UserController) GetAllUsers(c *gin.Context) {
	var pagination entity.Pagination
	var search entity.Search

	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid pagination parameters",
			Error:   err.Error(),
		})
		return
	}

	if err := c.ShouldBindQuery(&search); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid search parameters",
			Error:   err.Error(),
		})
		return
	}

	users, meta, err := uc.userService.GetAllUsers(&pagination, &search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{
			Success: false,
			Message: "Failed to retrieve users",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.PaginatedResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
		Meta:    *meta,
	})
}

// DeleteUser godoc
// @Summary Delete user (Admin only)
// @Description Delete a user by ID
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {object} entity.Response
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Failure 404 {object} entity.Response
// @Router /users/{id} [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "User ID is required",
		})
		return
	}

	err := uc.userService.DeleteUser(userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "cannot delete admin user" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "User deleted successfully",
	})
} 