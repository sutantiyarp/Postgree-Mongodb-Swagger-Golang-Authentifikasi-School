// package config

// import (
// 	// "database/sql"
// 	"hello-fiber/route"
// 	"hello-fiber/middleware"
// 	"github.com/gofiber/fiber/v2"
// 	"hello-fiber/database"  // Mengimpor package database
// )

// func NewApp() *fiber.App {
// 	// Connect to the database
// 	db := database.ConnectDB()

// 	// Initialize the Fiber application
// 	app := fiber.New()

// 	// Middleware
// 	app.Use(middleware.LoggerMiddleware)

// 	// Set up routes, passing db as a dependency to the route handler
// 	route.SetupRoutes(app, db)

// 	return app
// }

package config

import (
	// "database/sql"
	"hello-fiber/route"
	"hello-fiber/middleware"
	"github.com/gofiber/fiber/v2"
	"hello-fiber/database"
)

func NewApp() *fiber.App {
	// Connect ke database
	db := database.ConnectDB()

	// Initialize the Fiber application
	app := fiber.New()

	// Middleware
	app.Use(middleware.LoggerMiddleware)

	// Set up routes, passing db as a dependency to the route handler
	route.SetupRoutes(app, db)

	return app
}