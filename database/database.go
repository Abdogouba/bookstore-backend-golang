package database

import (
	"fmt"
	"log"

	"bookstore-backend/config"
	"bookstore-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the global database connection instance.
var DB *gorm.DB

// Connect connects to the main development database.
func Connect() {
	cfg := config.AppConfig

	db := connect(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	DB = db

	log.Println("Connected to PostgreSQL database")
}

// ConnectTestDB connects to the test database.
// Integration tests should use this database only.
func ConnectTestDB() *gorm.DB {
	cfg := config.AppConfig

	db := connect(
		cfg.TestDBHost,
		cfg.TestDBPort,
		cfg.TestDBUser,
		cfg.TestDBPassword,
		cfg.TestDBName,
		cfg.TestDBSSLMode,
	)

	log.Println("Connected to TEST PostgreSQL database")

	return db
}

// connect creates a PostgreSQL connection and runs migrations.
func connect(
	host,
	port,
	user,
	password,
	dbName,
	sslMode string,
) *gorm.DB {

	// Build PostgreSQL DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host,
		user,
		password,
		dbName,
		port,
		sslMode,
	)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Run database migrations
	if err := migrate(db); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	return db
}

// migrate creates/updates database tables.
func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Book{},
		&models.Order{},
		&models.OrderItem{},
	)
}