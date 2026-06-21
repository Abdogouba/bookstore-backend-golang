package handlers

import (
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(
	db *gorm.DB,
) *AuthHandler {

	return &AuthHandler{
		authService: services.NewAuthService(db),
	}
}

// Register godoc
//
//	@Summary		Register user
//	@Description	Create a new user account
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.RegisterRequest	true	"Register Request"
//	@Success		201	{object}	dto.UserResponse
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(
	c *gin.Context,
) {

	var request dto.RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)
		return
	}

	user, err := h.authService.Register(request)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		user,
	)
}

// Login godoc
//
// @Summary Login
// @Description Login with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Request"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(
	c *gin.Context,
) {

	var req dto.LoginRequest

	if err := c.ShouldBindJSON(
		&req,
	); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	response, err :=
		h.authService.Login(req)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

// RefreshToken godoc
//
// @Summary Refresh Access Token
// @Description Generate a new access token using a refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} dto.RefreshTokenResponse
// @Failure 400 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(
	c *gin.Context,
) {

	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(
		&req,
	); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	response, err :=
		h.authService.RefreshToken(req)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}

// Logout godoc
//
// @Summary Logout
// @Description Logout user and invalidate refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Logout Request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(
	c *gin.Context,
) {

	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(
		&req,
	); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	response, err :=
		h.authService.Logout(req)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		response,
	)
}