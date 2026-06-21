package handlers

import (
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/services"
	"bookstore-backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(
	db *gorm.DB,
) *UserHandler {

	return &UserHandler{
		userService: services.NewUserService(db),
	}
}

// GetProfile godoc
//
// @Summary View Profile
// @Description Returns the authenticated user's profile
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(
	c *gin.Context,
) {

	// Get authenticated user ID
	userID :=
		utils.GetUserID(c)

	response, err :=
		h.userService.GetProfile(
			userID,
		)

	if err != nil {

		if err.Error() ==
			"user not found" {

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

// UpdateProfile godoc
//
// @Summary Edit My Profile
// @Description Update authenticated user's profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "Profile data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(
	c *gin.Context,
) {

	var request dto.UpdateProfileRequest

	if err :=
		c.ShouldBindJSON(&request); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	userID :=
		utils.GetUserID(c)

	response, err :=
		h.userService.UpdateProfile(
			userID,
			request,
		)

	if err != nil {

		switch err.Error() {

		case "user not found":

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				},
			)

			return

		case "email already exists":

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

// ChangePassword godoc
//
// @Summary Change Password
// @Description Change authenticated user's password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "Passwords"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/change-password [put]
func (h *UserHandler) ChangePassword(
	c *gin.Context,
) {

	var request dto.ChangePasswordRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	userID := utils.GetUserID(c)

	err := h.userService.ChangePassword(
		userID,
		request,
	)

	if err != nil {

		switch err.Error() {

		case "user not found":

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				},
			)

			return

		case "current password is incorrect":

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	response := dto.MessageResponse{Message: "password changed successfully"}

	c.JSON(
		http.StatusOK,
		response,
	)
}

// DeleteProfile godoc
// @Summary Delete My Profile
// @Description Soft deletes the authenticated user account after verifying password. User must not have active orders (pending, confirmed, out_for_delivery).
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.DeleteProfileRequest true "Password confirmation"
// @Success 200 {object} dto.MessageResponse "Profile deleted successfully"
// @Failure 400 {object} map[string]string "Invalid request / wrong password / active orders exist"
// @Failure 401 {object} map[string]string "Unauthorized (missing or invalid token)"
// @Failure 403 {object} map[string]string "Forbidden (role not allowed)"
// @Failure 404 {object} map[string]string "User not found"
// @Router /users/profile [delete]
func (h *UserHandler) DeleteProfile(c *gin.Context) {

	var request dto.DeleteProfileRequest

	// -------------------------
	// Validate request
	// -------------------------
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// -------------------------
	// Get user from JWT
	// -------------------------
	userID := utils.GetUserID(c)

	// -------------------------
	// Call service
	// -------------------------
	err := h.userService.DeleteProfile(userID, request)

	if err != nil {

		switch err.Error() {

		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return

		case "password is incorrect":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		case "cannot delete account with active orders":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// -------------------------
	// Success response
	// -------------------------
	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "profile deleted successfully",
	})
}