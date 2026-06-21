package models

import (
	"time"

	"gorm.io/gorm"
)

// Order represents a user's purchase order.
type Order struct {
	ID uint `gorm:"primaryKey"`

	UserID uint `gorm:"not null"`

	// Belongs to one user.
	User User

	Status string `gorm:"size:30;default:pending;not null"`

	ShippingAddress string `gorm:"type:text;not null"`

	TotalPrice float64 `gorm:"type:decimal(10,2);not null"`

	// One order can contain many order items.
	OrderItems []OrderItem

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
