package middleware

import (
	"net/http"
	"strings"

	"bookstore-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the access token
// and stores user data in Gin context.
func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		// -------------------------
		// Get Authorization Header
		// -------------------------

		authHeader :=
			c.GetHeader("Authorization")

		if authHeader == "" {

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "authorization header is required",
				},
			)

			c.Abort()
			return
		}

		// -------------------------
		// Validate Bearer Format
		// -------------------------

		parts := strings.Split(
			authHeader,
			" ",
		)

		if len(parts) != 2 ||
			parts[0] != "Bearer" {

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid authorization header",
				},
			)

			c.Abort()
			return
		}

		// -------------------------
		// Extract Token
		// -------------------------

		tokenString := parts[1]

		// -------------------------
		// Validate Token
		// -------------------------

		claims, err :=
			utils.ValidateAccessToken(
				tokenString,
			)

		if err != nil {

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid access token",
				},
			)

			c.Abort()
			return
		}

		// -------------------------
		// Save User Info
		// -------------------------

		c.Set(
			"userID",
			claims.UserID,
		)

		c.Set(
			"role",
			claims.Role,
		)

		// Continue to handler.
		c.Next()
	}
}

// RoleMiddleware checks whether the authenticated
// user has the required role.
func RoleMiddleware(
	requiredRole string,
) gin.HandlerFunc {

	return func(c *gin.Context) {

		// -------------------------
		// Get role from context
		// -------------------------

		role, exists := c.Get("role")

		if !exists {

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "user role not found",
				},
			)

			c.Abort()
			return
		}

		// -------------------------
		// Type assertion
		// -------------------------

		userRole := role.(string)

		// -------------------------
		// Compare roles
		// -------------------------

		if userRole != requiredRole {

			c.JSON(
				http.StatusForbidden,
				gin.H{
					"error": "forbidden",
				},
			)

			c.Abort()
			return
		}

		// Continue.
		c.Next()
	}
}