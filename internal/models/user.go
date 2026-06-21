package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents application users.
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"size:100;not null"`
	Email        string `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	PhoneNumber  string `gorm:"size:20"`
	Role         string `gorm:"size:20;default:user;not null"`

	// One user can have many orders.
	Orders []Order

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
