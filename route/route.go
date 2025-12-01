package route

import (
	"database/sql"
	"hello-fiber/app/service"
	"github.com/gofiber/fiber/v2"
	"hello-fiber/middleware"
)

func SetupRoutes(app *fiber.App, db *sql.DB) {
	service.InitUserService(db)
	service.InitRepoService(db)
	service.InitPermissionService(db)
	service.InitRolePermissionService(db)
	service.InitLecturerService(db)
	service.InitStudentService(db)
	api := app.Group("/api")

	api.Post("/v1/auth/register", func(c *fiber.Ctx) error {
		return service.Register(c, db)
	})
	api.Post("/v1/auth/login", func(c *fiber.Ctx) error {
		return service.Login(c, db)
	})

	protected := api.Group("/", middleware.JWTAuthMiddleware())

	user := protected.Group("/v1/users", middleware.AdminOnlyMiddleware(db))
	user.Get("/", service.GetAllUsersService)
	// user.Get("/byrole", service.GetUsersByRoleNameService)
	// user.Get("/byemail", service.GetUserByEmailService)
	// user.Get("/byusername", service.GetUserByUsernameService)
	user.Get("/:id", service.GetUserByIDService)
	user.Post("/", service.CreateUserAdmin)
	user.Put("/:id", service.UpdateUserService)
	user.Delete("/:id", service.DeleteUserService)

	role := protected.Group("/v1/roles", middleware.AdminOnlyMiddleware(db))
	role.Get("/", service.GetAllRolesService)
	// role.Get("/byname", service.GetRoleByNameService)
	role.Get("/:id", service.GetRoleByIDService)
	role.Post("/", service.CreateRoleService)
	role.Put("/:id", service.UpdateRoleService)
	role.Delete("/:id", service.DeleteRoleService)

	permission := protected.Group("/v1/permissions", middleware.AdminOnlyMiddleware(db))
	permission.Get("/", service.GetAllPermissionsService)
	permission.Get("/:id", service.GetPermissionByIDService)
	permission.Post("/", service.CreatePermissionService)
	permission.Put("/:id", service.UpdatePermissionService)
	permission.Delete("/:id", service.DeletePermissionService)

	rolePermission := protected.Group("/v1/role-permissions", middleware.AdminOnlyMiddleware(db))
	rolePermission.Get("/", service.GetAllRolePermissionsService)
	rolePermission.Get("/byrole/:role_id", service.GetPermissionsByRoleIDService)
	// rolePermission.Get("/:role_id/:permission_id", service.GetRolePermissionDetailService)
	rolePermission.Post("/", service.CreateRolePermissionService)
	rolePermission.Put("/:role_id/:permission_id", service.UpdateRolePermissionService)
	rolePermission.Delete("/:role_id/:permission_id", service.DeleteRolePermissionService)

	lecturer := protected.Group("/v1/lecturers", middleware.AdminOnlyMiddleware(db))
	lecturer.Get("/", service.GetAllLecturersService)
	lecturer.Get("/:id", service.GetLecturerByIDService)
	lecturer.Post("/", service.CreateLecturerService)
	lecturer.Put("/:id", service.UpdateLecturerService)
	lecturer.Delete("/:id", service.DeleteLecturerService)

	student := protected.Group("/v1/students", middleware.AdminOnlyMiddleware(db))
	student.Get("/", service.GetAllStudentsService)
	student.Get("/:id", service.GetStudentByIDService)
	student.Post("/", service.CreateStudentService)
	student.Put("/:id", service.UpdateStudentService)
	student.Delete("/:id", service.DeleteStudentService)
}
