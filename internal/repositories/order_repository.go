package repositories

import (
	"bookstore-backend/internal/dto"
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

// GetUserOrderByID returns a single order
// that belongs to the specified user.
func (r *OrderRepository) GetUserOrderByID(
	orderID uint,
	userID uint,
) (
	*models.Order,
	error,
) {

	var order models.Order

	err := r.db.
		Preload("OrderItems").
		Where(
			"id = ? AND user_id = ?",
			orderID,
			userID,
		).
		First(&order).
		Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// GetAllOrders returns paginated orders
// with optional user name search
// and status filter.
func (r *OrderRepository) GetAllOrders(
	query dto.AdminGetOrdersQuery,
) (
	[]models.Order,
	int64,
	error,
) {

	var orders []models.Order

	var total int64

	// -------------------------
	// Start query
	// -------------------------

	db := r.db.
		    Model(&models.Order{}).
			Joins("JOIN users ON users.id = orders.user_id")

	// -------------------------
	// Search user name
	// -------------------------

	if query.UserName != "" {

		db =
			db.Where(
				"users.name ILIKE ?",
				"%"+query.UserName+"%",
			)
	}

	// -------------------------
	// Filter status
	// -------------------------

	if query.Status != "" {

		db =
			db.Where(
				"orders.status = ?",
				query.Status,
			)
	}

	// -------------------------
	// Count total
	// -------------------------

	err :=
		db.
			Count(&total).
			Error

	if err != nil {

		return nil,
			0,
			err
	}

	// -------------------------
	// Pagination
	// -------------------------

	offset :=
		(query.Page - 1) *
			query.PageSize

	// -------------------------
	// Query Orders
	// -------------------------

	err =
		db.
			Preload("User").
			Preload("OrderItems").
			Order("orders.created_at DESC").
			Offset(offset).
			Limit(query.PageSize).
			Find(&orders).
			Error

	if err != nil {

		return nil,
			0,
			err
	}

	return orders,
		total,
		nil
}
