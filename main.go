package main

import (
	"log"

	"github.com/joho/godotenv"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"hello-fiber/config"
	"hello-fiber/database"
	_ "hello-fiber/docs" // Import generated docs package
)

// @title Alumni Management API
// @version 1.0.0
// @description API untuk mengelola data alumni, user, pekerjaan alumni, dan file dengan MongoDB menggunakan Clean Architecture
// @contact.name API Support
// @contact.url http://localhost:3000
// @license.name MIT
// @host localhost:3000
// @BasePath /api
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env not loaded:", err)
	}

	// NewApp will call ConnectMongoDB internally
	app := config.NewApp()

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// disconnect saat program keluar (DisconnectMongoDB harus aman dipanggil jika belum terhubung)
	defer func() {
		if err := database.DisconnectMongoDB(); err != nil {
			log.Println("Error disconnecting from MongoDB:", err)
		}
	}()

	log.Fatal(app.Listen(":3000"))
}
