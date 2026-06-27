package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"bookstore-backend/internal/handlers"
	"bookstore-backend/internal/middleware"
)

func SetupRoutes(
	router *gin.Engine,
	db *gorm.DB,
) {
	authHandler := handlers.NewAuthHandler(db)
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
	}

	userHandler := handlers.NewUserHandler(db)
	users := router.Group("/users")
	{
		users.GET(
			"/profile",
			middleware.AuthMiddleware(),
			userHandler.GetProfile,
		)

		users.PUT(
			"/profile",
			middleware.AuthMiddleware(),
			userHandler.UpdateProfile,
		)

		users.PUT(
			"/change-password",
			middleware.AuthMiddleware(),
			userHandler.ChangePassword,
		)

		users.DELETE(
			"/profile",
			middleware.AuthMiddleware(),
			middleware.RoleMiddleware("user"),
			userHandler.DeleteProfile,
		)
	}

	bookHandler := handlers.NewBookHandler(db)
	books := router.Group("/books")
	{
		books.POST(
			"",
			middleware.AuthMiddleware(),
			middleware.RoleMiddleware("admin"),
			bookHandler.CreateBook,
		)

		books.GET(
			"",
			bookHandler.GetBooks,
		)

		books.GET(
			"/:id",
			bookHandler.GetBook,
		)

		books.PUT(
			"/:id",
			middleware.AuthMiddleware(),
			middleware.RoleMiddleware("admin"),
			bookHandler.UpdateBook,
		)

		books.DELETE(
			"/:id",
			middleware.AuthMiddleware(),
			middleware.RoleMiddleware("admin"),
			bookHandler.DeleteBook,
		)
	}

	orderHandler := handlers.NewOrderHandler(db)
	orders := router.Group("/orders")
	{
		orders.POST(
			"",
			middleware.AuthMiddleware(),
			middleware.RoleMiddleware("user"),
			orderHandler.CreateOrder,
		)

		orders.GET(
			"",
			middleware.AuthMiddleware(),
			middleware.RoleMiddleware("user"),
			orderHandler.GetMyOrders,
		)

		orders.GET(
			"/:id",
			middleware.AuthMiddleware(),
			middleware.RoleMiddleware("user"),
			orderHandler.GetMyOrder,
		)
	}

	admin := router.Group("/admin")
	{
		admin.GET(
			"/orders",
			middleware.AuthMiddleware(),
			middleware.RoleMiddleware("admin"),
			orderHandler.GetAllOrders,
		)

		admin.GET(
			"/orders/:id",
			middleware.AuthMiddleware(),
			middleware.RoleMiddleware("admin"),
			orderHandler.GetOrder,
		)
	}
}


