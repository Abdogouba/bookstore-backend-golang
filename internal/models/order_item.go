package models

// OrderItem represents books inside an order.
type OrderItem struct {
	ID uint `gorm:"primaryKey"`

	OrderID uint `gorm:"not null"`

	// Belongs to one order.
	Order Order

	BookID uint `gorm:"not null"`

	// References one book.
	Book Book

	Quantity int `gorm:"not null"`

	// Store price at purchase time.
	Price float64 `gorm:"type:decimal(10,2);not null"`
}
