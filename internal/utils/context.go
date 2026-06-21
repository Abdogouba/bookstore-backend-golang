package utils

import "github.com/gin-gonic/gin"

// GetUserID retrieves the authenticated user's ID.
func GetUserID(
	c *gin.Context,
) uint {

	userID, _ :=
		c.Get("userID")

	return userID.(uint)
}

// GetUserRole retrieves the authenticated user's role.
func GetUserRole(
	c *gin.Context,
) string {

	role, _ :=
		c.Get("role")

	return role.(string)
}