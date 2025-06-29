package middleware

import (
	"net/http"
	"reflect"
	"strings"
	"ticketing-system/entity"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom tag name function to use json tag names in error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateJSON validates the JSON request body against the provided struct
func ValidateJSON(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(obj); err != nil {
			c.JSON(http.StatusBadRequest, entity.Response{
				Success: false,
				Message: "Invalid JSON format",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		if err := validate.Struct(obj); err != nil {
			var errors []string
			for _, err := range err.(validator.ValidationErrors) {
				errors = append(errors, formatValidationError(err))
			}

			c.JSON(http.StatusBadRequest, entity.Response{
				Success: false,
				Message: "Validation failed",
				Error:   errors,
			})
			c.Abort()
			return
		}

		// Store validated object in context
		c.Set("validated_data", obj)
		c.Next()
	}
}

// ValidateQuery validates query parameters
func ValidateQuery(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindQuery(obj); err != nil {
			c.JSON(http.StatusBadRequest, entity.Response{
				Success: false,
				Message: "Invalid query parameters",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		if err := validate.Struct(obj); err != nil {
			var errors []string
			for _, err := range err.(validator.ValidationErrors) {
				errors = append(errors, formatValidationError(err))
			}

			c.JSON(http.StatusBadRequest, entity.Response{
				Success: false,
				Message: "Query validation failed",
				Error:   errors,
			})
			c.Abort()
			return
		}

		// Store validated object in context
		c.Set("validated_query", obj)
		c.Next()
	}
}

func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + err.Param() + " characters long"
	case "max":
		return field + " must be at most " + err.Param() + " characters long"
	case "oneof":
		return field + " must be one of: " + err.Param()
	default:
		return field + " validation failed: " + tag
	}
}

// GetValidatedData retrieves validated data from context
func GetValidatedData(c *gin.Context) (interface{}, bool) {
	data, exists := c.Get("validated_data")
	return data, exists
}

// GetValidatedQuery retrieves validated query from context
func GetValidatedQuery(c *gin.Context) (interface{}, bool) {
	data, exists := c.Get("validated_query")
	return data, exists
} 