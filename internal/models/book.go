package models

import (
	"time"

	"gorm.io/gorm"
)

// Book represents books in the store.
type Book struct {
	ID uint `gorm:"primaryKey"`

	Title     string `gorm:"size:255;not null;index"`
	Author    string `gorm:"size:255;not null"`
	Publisher string `gorm:"size:255"`
	Category  string `gorm:"size:100;index;not null"`

	Price float64 `gorm:"type:decimal(10,2);not null"`
	Stock int     `gorm:"not null"`

	ImagePath string `gorm:"size:500"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
