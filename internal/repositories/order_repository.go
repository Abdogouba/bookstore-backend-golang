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

func (r *OrderRepository) Create(
	tx *gorm.DB,
	order *models.Order,
) error {

	return tx.
		Create(order).
		Error
}

// GetUserOrders returns paginated orders
// belonging to a specific user.
func (r *OrderRepository) GetUserOrders(
	userID uint,
	page int,
	pageSize int,
) (
	[]models.Order,
	int64,
	error,
) {

	var orders []models.Order

	var total int64

	// -------------------------
	// Count total
	// -------------------------

	err := r.db.
		Model(&models.Order{}).
		Where("user_id = ?", userID).
		Count(&total).
		Error

	if err != nil {
		return nil, 0, err
	}

	// -------------------------
	// Pagination
	// -------------------------

	offset :=
		(page - 1) * pageSize

	// -------------------------
	// Query orders
	// -------------------------

	err = r.db.
		Preload("OrderItems").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).
		Error

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}