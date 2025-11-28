package route

import (
	"database/sql"
	"hello-fiber/app/service"
	"github.com/gofiber/fiber/v2"
	"hello-fiber/middleware"
)

func SetupRoutes(app *fiber.App, db *sql.DB) {
	// init repository/service (pakai db dari main)
	service.InitUserService(db)

	api := app.Group("/api")

	// ===== PUBLIC =====
	api.Post("/register", func(c *fiber.Ctx) error {
		return service.Register(c, db)
	})

	api.Post("/login", func(c *fiber.Ctx) error {
		return service.Login(c, db)
	})

	// ===== PROTECTED (JWT) =====
	protected := api.Group("/", middleware.JWTAuthMiddleware())

	// ===== USERS (Admin Only) =====
	users := protected.Group("/users", middleware.AdminOnlyMiddleware(db))

	// CRUD Users
	users.Get("/", service.GetAllUsersService)      // READ list
	users.Get("/:id", service.GetUserByIDService) // READ detail
	users.Post("/", service.CreateUserAdmin)        // CREATE
	users.Put("/:id", service.UpdateUserService)    // UPDATE
	users.Delete("/:id", service.DeleteUserService) // DELETE
}
