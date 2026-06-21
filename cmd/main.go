package main

import (
	"log"

	"bookstore-backend/config"
	"bookstore-backend/database"
	"bookstore-backend/internal/routes"
	"bookstore-backend/internal/seeder"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "bookstore-backend/docs"
)

// @title Bookstore API
// @version 1.0
// @description Simple bookstore backend built with Go, Gin, GORM and PostgreSQL.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	// Load environment variables.
	config.LoadConfig()

	// Connect database.
	database.Connect()

	// Seed admin
	err := seeder.SeedAdmin(database.DB)
	if err != nil {
		log.Fatal("Failed to seed admin: ", err)
	}

	// Create Gin router.
	router := gin.Default()

	router.Static("/uploads", "./uploads")

	// Swagger endpoint.
	router.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	// Register application routes.
	routes.SetupRoutes(router, database.DB)

	log.Printf(
		"Server running on port %s",
		config.AppConfig.Port,
	)

	if err := router.Run(
		":" + config.AppConfig.Port,
	); err != nil {
		log.Fatal(err)
	}
}
