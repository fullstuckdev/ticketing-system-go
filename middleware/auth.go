package middleware

import (
	"net/http"
	"strings"
	"ticketing-system/entity"
	"ticketing-system/service"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	userService service.UserService
}

func NewAuthMiddleware(userService service.UserService) *AuthMiddleware {
	return &AuthMiddleware{userService: userService}
}

// AuthRequired ensures the request has a valid JWT token
func (a *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, entity.Response{
				Success: false,
				Message: "Authorization header required",
				Error:   "missing_authorization_header",
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, entity.Response{
				Success: false,
				Message: "Invalid authorization header format",
				Error:   "invalid_authorization_format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		user, err := a.userService.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, entity.Response{
				Success: false,
				Message: "Invalid or expired token",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		// Store user in context for use in handlers
		c.Set("current_user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Next()
	}
}

// AdminRequired ensures the user has admin role
func (a *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user, exists := c.Get("current_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, entity.Response{
				Success: false,
				Message: "Authentication required",
				Error:   "missing_user_context",
			})
			c.Abort()
			return
		}

		currentUser, ok := user.(*entity.User)
		if !ok || !currentUser.IsAdmin() {
			c.JSON(http.StatusForbidden, entity.Response{
				Success: false,
				Message: "Admin access required",
				Error:   "insufficient_permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// OptionalAuth middleware that validates JWT if present but doesn't require it
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		token := tokenParts[1]
		user, err := a.userService.ValidateJWT(token)
		if err == nil {
			c.Set("current_user", user)
			c.Set("user_id", user.ID)
			c.Set("user_role", user.Role)
		}

		c.Next()
	}
}

// GetCurrentUser helper function to get current user from context
func GetCurrentUser(c *gin.Context) (*entity.User, bool) {
	user, exists := c.Get("current_user")
	if !exists {
		return nil, false
	}

	currentUser, ok := user.(*entity.User)
	return currentUser, ok
}

// GetCurrentUserID helper function to get current user ID from context
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

// IsAdmin helper function to check if current user is admin
func IsAdmin(c *gin.Context) bool {
	role, exists := c.Get("user_role")
	if !exists {
		return false
	}

	userRole, ok := role.(entity.UserRole)
	return ok && userRole == entity.RoleAdmin
} 