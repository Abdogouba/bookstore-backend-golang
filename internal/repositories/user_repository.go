package repositories

import (
	"bookstore-backend/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

// Inject database dependency.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create new user.
func (r *UserRepository) Create(
	user *models.User,
) error {

	return r.db.Create(user).Error
}

// Find user by email.
func (r *UserRepository) GetByEmail(
	email string,
) (*models.User, error) {

	var user models.User

	err := r.db.
		Where("email = ?", email).
		First(&user).
		Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByID(
	id uint,
) (*models.User, error) {

	var user models.User

	err := r.db.
		First(&user, id).
		Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Save updates an existing user.
func (r *UserRepository) Save(
	user *models.User,
) error {

	return r.db.Save(user).Error
}

// Delete performs a soft delete on the user.
// Because User model includes gorm.DeletedAt,
// GORM will automatically set deleted_at instead of removing the row.
func (r *UserRepository) Delete(user *models.User) error {
	return r.db.Delete(user).Error
}