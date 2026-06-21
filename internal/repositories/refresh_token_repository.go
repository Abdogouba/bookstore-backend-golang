package repositories

import (
	"bookstore-backend/internal/models"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(
	db *gorm.DB,
) *RefreshTokenRepository {

	return &RefreshTokenRepository{
		db: db,
	}
}

// Create saves a refresh token.
func (r *RefreshTokenRepository) Create(
	token *models.RefreshToken,
) error {

	return r.db.Create(token).Error
}

// Delete all refresh tokens belonging to a user.
func (r *RefreshTokenRepository) DeleteByUserID(
	userID uint,
) error {

	return r.db.
		Where("user_id = ?", userID).
		Delete(&models.RefreshToken{}).
		Error
}

func (r *RefreshTokenRepository) GetByToken(
	token string,
) (*models.RefreshToken, error) {

	var refreshToken models.RefreshToken

	err := r.db.
		Where("token = ?", token).
		First(&refreshToken).
		Error

	if err != nil {
		return nil, err
	}

	return &refreshToken, nil
}

// DeleteByToken removes a refresh token.
func (r *RefreshTokenRepository) DeleteByToken(
	token string,
) error {

	return r.db.
		Where("token = ?", token).
		Delete(&models.RefreshToken{}).
		Error
}