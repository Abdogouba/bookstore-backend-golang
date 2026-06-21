package services

import (
	"errors"

	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/mappers"
	"bookstore-backend/internal/models"
	"bookstore-backend/internal/repositories"
	"bookstore-backend/internal/utils"

	"gorm.io/gorm"
)

type AuthService struct {
	userRepo         *repositories.UserRepository
	refreshTokenRepo *repositories.RefreshTokenRepository
}

func NewAuthService(
	db *gorm.DB,
) *AuthService {

	return &AuthService{
		userRepo: repositories.NewUserRepository(db),

		refreshTokenRepo:
			repositories.NewRefreshTokenRepository(db),
	}
}

func (s *AuthService) Register(
	request dto.RegisterRequest,
) (*dto.UserResponse, error) {

	// Check if email already exists.
	_, err := s.userRepo.GetByEmail(request.Email)

	if err == nil {
		return nil, errors.New("email already exists")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hash password.
	hashedPassword, err := utils.HashPassword(
		request.Password,
	)

	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:         request.Name,
		Email:        request.Email,
		PasswordHash: hashedPassword,
		PhoneNumber:  request.PhoneNumber,
		Role:         "user",
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, err
	}

	response := mappers.ToUserResponse(&user)

	return &response, nil
}

func (s *AuthService) Login(
	req dto.LoginRequest,
) (*dto.LoginResponse, error) {

	// Find user
	user, err := s.userRepo.GetByEmail(
		req.Email,
	)

	if err != nil {
		return nil, errors.New(
			"invalid credentials",
		)
	}

	// Verify password
	err = utils.CheckPassword(
		req.Password,
		user.PasswordHash,
	)

	if err != nil {
		return nil, errors.New(
			"invalid credentials",
		)
	}

	// Remove old refresh tokens
	err = s.refreshTokenRepo.DeleteByUserID(
		user.ID,
	)

	if err != nil {
		return nil, err
	}

	// Generate access token
	accessToken, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken,
		expiresAt,
		err :=
		utils.GenerateRefreshToken(
			user.ID,
		)

	if err != nil {
		return nil, err
	}

	// Save refresh token
	err = s.refreshTokenRepo.Create(
		&models.RefreshToken{
			UserID:    user.ID,
			Token:     refreshToken,
			ExpiresAt: expiresAt,
		},
	)

	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		User: mappers.ToUserResponse(user),

		AccessToken: accessToken,

		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(
	req dto.RefreshTokenRequest,
) (*dto.RefreshTokenResponse, error) {

	// --------------------------------
	// Validate JWT
	// --------------------------------

	claims, err :=
		utils.ValidateRefreshToken(
			req.RefreshToken,
		)

	if err != nil {
		return nil,
			errors.New("invalid refresh token")
	}

	// --------------------------------
	// Verify token exists in DB
	// --------------------------------

	refreshToken, err :=
		s.refreshTokenRepo.GetByToken(
			req.RefreshToken,
		)

	if err != nil {
		return nil,
			errors.New("invalid refresh token")
	}

	// --------------------------------
	// Verify token belongs to user
	// --------------------------------

	user, err :=
		s.userRepo.GetByID(
			claims.UserID,
		)

	if err != nil {
		return nil,
			errors.New("user not found")
	}

	// Optional safety check
	if refreshToken.UserID != user.ID {
		return nil,
			errors.New("invalid refresh token")
	}

	// --------------------------------
	// Generate new access token
	// --------------------------------

	accessToken, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *AuthService) Logout(
	req dto.RefreshTokenRequest,
) (*dto.MessageResponse, error) {

	err := s.refreshTokenRepo.DeleteByToken(
		req.RefreshToken,
	)

	if err != nil {
		return nil, err
	}

	return &dto.MessageResponse{
		Message: "logged out successfully",
	}, nil
}