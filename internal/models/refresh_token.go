package models

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken stores refresh tokens for users.
type RefreshToken struct {
	ID uint `gorm:"primaryKey"`

	UserID uint `gorm:"not null"`

	// Belongs to one user.
	User User

	Token string `gorm:"type:text;not null"`

	ExpiresAt time.Time `gorm:"not null"`

	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
