package seeder

import (
	"log"

	"bookstore-backend/config"
	"bookstore-backend/internal/models"
	"bookstore-backend/internal/utils"

	"gorm.io/gorm"
)

// SeedAdmin ensures an admin user exists in the system.
func SeedAdmin(db *gorm.DB) error {

	// Check if admin already exists
	var count int64
	err := db.Model(&models.User{}).
		Where("role = ?", "admin").
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	// Hash admin password
	hashedPassword, err := utils.HashPassword(config.AppConfig.AdminPassword)
	if err != nil {
		return err
	}

	// Create admin user
	admin := models.User{
		Name:         "Admin",
		Email:        config.AppConfig.AdminEmail,
		PasswordHash: hashedPassword,
		Role:         "admin",
	}

	err = db.Create(&admin).Error
	if err != nil {
		return err
	}

	log.Println("Admin user created successfully")

	return nil
}
