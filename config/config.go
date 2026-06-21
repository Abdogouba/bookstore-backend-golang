package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values.
// We store everything in one struct so the app can access configs cleanly.
type Config struct {
	AppEnv string
	Port   string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	TestDBHost     string
	TestDBPort     string
	TestDBUser     string
	TestDBPassword string
	TestDBName     string
	TestDBSSLMode  string

	JWTSecret            string
	AccessTokenDuration  string
	RefreshTokenDuration string

	AdminEmail    string
	AdminPassword string
}

// AppConfig will hold the loaded configuration globally.
var AppConfig Config

// LoadConfig loads environment variables from .env file
// and stores them inside AppConfig.
func LoadConfig() {
	// Load .env file
	err := godotenv.Load()

	if err != nil {
		// Try parent directory (useful for tests)
		err = godotenv.Load("../.env")
	}

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	AppConfig = Config{
		AppEnv: os.Getenv("APP_ENV"),
		Port:   os.Getenv("PORT"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),

		TestDBHost:     os.Getenv("TEST_DB_HOST"),
		TestDBPort:     os.Getenv("TEST_DB_PORT"),
		TestDBUser:     os.Getenv("TEST_DB_USER"),
		TestDBPassword: os.Getenv("TEST_DB_PASSWORD"),
		TestDBName:     os.Getenv("TEST_DB_NAME"),
		TestDBSSLMode:  os.Getenv("TEST_DB_SSLMODE"),

		JWTSecret:            os.Getenv("JWT_SECRET"),
		AccessTokenDuration:  os.Getenv("ACCESS_TOKEN_DURATION"),
		RefreshTokenDuration: os.Getenv("REFRESH_TOKEN_DURATION"),

		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}
}
