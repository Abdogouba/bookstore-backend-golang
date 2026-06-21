package services

import (
	"errors"

	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/mappers"
	"bookstore-backend/internal/repositories"
	"bookstore-backend/internal/utils"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo         *repositories.UserRepository
	refreshTokenRepo *repositories.RefreshTokenRepository
	orderRepo        *repositories.OrderRepository
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		userRepo:         repositories.NewUserRepository(db),
		refreshTokenRepo: repositories.NewRefreshTokenRepository(db),
		orderRepo:        repositories.NewOrderRepository(db),
	}
}

// GetProfile returns the authenticated user's profile.
func (s *UserService) GetProfile(
	userID uint,
) (*dto.UserResponse, error) {

	user, err :=
		s.userRepo.GetByID(userID)

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {
			return nil,
				errors.New("user not found")
		}

		return nil, err
	}

	response := mappers.ToUserResponse(user)

	return &response, nil
}

// UpdateProfile updates the authenticated user's profile.
func (s *UserService) UpdateProfile(
	userID uint,
	request dto.UpdateProfileRequest,
) (*dto.UserResponse, error) {

	// -------------------------
	// Get current user
	// -------------------------

	user, err :=
		s.userRepo.GetByID(userID)

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {
			return nil,
				errors.New("user not found")
		}

		return nil, err
	}

	// -------------------------
	// Email uniqueness check
	// -------------------------

	existingUser, err :=
		s.userRepo.GetByEmail(
			request.Email,
		)

	if err == nil &&
		existingUser.ID != user.ID {

		return nil,
			errors.New("email already exists")
	}

	// -------------------------
	// Update fields
	// -------------------------

	user.Name = request.Name

	user.Email = request.Email

	user.PhoneNumber =
		request.PhoneNumber

	// -------------------------
	// Save changes
	// -------------------------

	err = s.userRepo.Save(user)

	if err != nil {
		return nil, err
	}

	// -------------------------
	// Build response DTO
	// -------------------------

	response := mappers.ToUserResponse(user)

	return &response, nil
}

func (s *UserService) ChangePassword(
	userID uint,
	request dto.ChangePasswordRequest,
) error {

	// -------------------------
	// Get user
	// -------------------------

	user, err := s.userRepo.GetByID(userID)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}

		return err
	}

	// -------------------------
	// Verify current password
	// -------------------------

	err = utils.CheckPassword(
		request.CurrentPassword,
		user.PasswordHash,
	)

	if err != nil {
		return errors.New("current password is incorrect")
	}

	// -------------------------
	// Hash new password
	// -------------------------

	hashedPassword, err :=
		utils.HashPassword(request.NewPassword)

	if err != nil {
		return err
	}

	// -------------------------
	// Update password
	// -------------------------

	user.PasswordHash = hashedPassword

	// -------------------------
	// Save
	// -------------------------

	err = s.userRepo.Save(user)

	if err != nil {
		return err
	}

	// -------------------------
	// Revoke all refresh tokens
	// -------------------------

	err = s.refreshTokenRepo.DeleteByUserID(user.ID)
	
	if err != nil {
		return err
	}

	return nil
}

// DeleteProfile deletes a user account after validation.
func (s *UserService) DeleteProfile(
	userID uint,
	request dto.DeleteProfileRequest,
) error {

	// -------------------------
	// Get user
	// -------------------------
	user, err := s.userRepo.GetByID(userID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// -------------------------
	// Verify password
	// -------------------------
	err = utils.CheckPassword(request.Password, user.PasswordHash)

	if err != nil {
		return errors.New("password is incorrect")
	}

	// -------------------------
	// Check active orders
	// -------------------------
	hasActiveOrders, err := s.orderRepo.HasActiveOrders(userID)

	if err != nil {
		return err
	}

	if hasActiveOrders {
		return errors.New("cannot delete account with active orders")
	}

	// -------------------------
	// Delete refresh tokens
	// -------------------------
	err = s.refreshTokenRepo.DeleteByUserID(userID)

	if err != nil {
		return err
	}

	// -------------------------
	// Soft delete user
	// -------------------------
	userErr := s.userRepo.Delete(user)

	if userErr != nil {
		return userErr
	}

	return nil
}