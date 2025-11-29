package route

import (
	"database/sql"
	"hello-fiber/app/service"
	"github.com/gofiber/fiber/v2"
	"hello-fiber/middleware"
)

func SetupRoutes(app *fiber.App, db *sql.DB) {
	service.InitUserService(db)
	api := app.Group("/api")

	api.Post("/register", func(c *fiber.Ctx) error {
		return service.Register(c, db)
	})
	api.Post("/login", func(c *fiber.Ctx) error {
		return service.Login(c, db)
	})

	protected := api.Group("/", middleware.JWTAuthMiddleware())

	users := protected.Group("/users", middleware.AdminOnlyMiddleware(db))
	users.Get("/", service.GetAllUsersService)
	users.Get("/:id", service.GetUserByIDService)
	users.Post("/", service.CreateUserAdmin)
	users.Put("/:id", service.UpdateUserService)
	users.Delete("/:id", service.DeleteUserService)

	roles := protected.Group("/roles", middleware.AdminOnlyMiddleware(db))
	roles.Get("/", service.GetAllRolesService)
	roles.Get("/byname", service.GetRoleByNameService)
	roles.Get("/:id", service.GetRoleByIDService)
}
