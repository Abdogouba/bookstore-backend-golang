package repositories

import (
	"bookstore-backend/internal/models"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

// Constructor
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

// Check if user has ANY active orders
// Active = orders that are NOT cancelable safely
func (r *OrderRepository) HasActiveOrders(userID uint) (bool, error) {

	var count int64

	err := r.db.Model(&models.Order{}).
		Where("user_id = ? AND status IN ?", userID,
			[]string{"pending", "confirmed", "out_for_delivery"}).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}