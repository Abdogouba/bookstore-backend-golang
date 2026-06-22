package repositories

import (
	"bookstore-backend/internal/models"

	"gorm.io/gorm"
)

type OrderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(
	db *gorm.DB,
) *OrderItemRepository {

	return &OrderItemRepository{
		db: db,
	}
}

func (r *OrderItemRepository) CreateMany(
	tx *gorm.DB,
	items []models.OrderItem,
) error {

	return tx.
		Create(&items).
		Error
}