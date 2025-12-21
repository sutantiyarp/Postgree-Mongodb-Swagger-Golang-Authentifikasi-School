package route

import (
	"database/sql"
	"hello-fiber/app/service"
	"hello-fiber/database"
	"hello-fiber/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, db *sql.DB) {
	service.InitUserService(db)
	service.InitRepoService(db)
	service.InitPermissionService(db)
	service.InitRolePermissionService(db)
	service.InitLecturerService(db)
	service.InitStudentService(db)
	service.InitAchievementService(db, database.MongoDB)
	api := app.Group("/api")

	api.Post("/v1/auth/register", func(c *fiber.Ctx) error {
		return service.Register(c, db)
	})
	api.Post("/v1/auth/login", func(c *fiber.Ctx) error {
		return service.Login(c, db)
	})
	api.Post("/v1/auth/refresh", func(c *fiber.Ctx) error {
		return service.Refresh(c, db)
	})
	api.Post("/v1/auth/logout", func(c *fiber.Ctx) error {
		return service.Logout(c, db)
	})
	api.Get("/v1/auth/profile", middleware.JWTAuthMiddleware(db), func(c *fiber.Ctx) error {
		return service.GetProfileService(c)
	})

	protected := api.Group("/", middleware.JWTAuthMiddleware(db))

	user := protected.Group("/v1/users", middleware.RequirePermission(db, "user:manage"))
	user.Get("/", service.GetAllUsersService)
	// user.Get("/byrole", service.GetUsersByRoleNameService)
	// user.Get("/byemail", service.GetUserByEmailService)
	// user.Get("/byusername", service.GetUserByUsernameService)
	user.Get("/:id", service.GetUserByIDService)
	user.Post("/", service.CreateUserAdmin)
	user.Put("/:id", service.UpdateUserService)
	user.Put("/:id/role", service.UpdateUserRoleByNameService)
	user.Delete("/:id", service.DeleteUserService)

	role := protected.Group("/v1/roles", middleware.RequirePermission(db, "user:manage"))
	role.Get("/", service.GetAllRolesService)
	// role.Get("/byname", service.GetRoleByNameService)
	role.Get("/:id", service.GetRoleByIDService)
	role.Post("/", service.CreateRoleService)
	role.Put("/:id", service.UpdateRoleService)
	role.Delete("/:id", service.DeleteRoleService)

	permission := protected.Group("/v1/permissions", middleware.RequirePermission(db, "user:manage"))
	permission.Get("/", service.GetAllPermissionsService)
	permission.Get("/:id", service.GetPermissionByIDService)
	permission.Post("/", service.CreatePermissionService)
	permission.Put("/:id", service.UpdatePermissionService)
	permission.Delete("/:id", service.DeletePermissionService)

	rolePermission := protected.Group("/v1/role-permissions", middleware.RequirePermission(db, "user:manage"))
	rolePermission.Get("/", service.GetAllRolePermissionsService)
	rolePermission.Get("/byrole/:role_id", service.GetPermissionsByRoleIDService)
	// rolePermission.Get("/:role_id/:permission_id", service.GetRolePermissionDetailService)
	rolePermission.Post("/", service.CreateRolePermissionService)
	rolePermission.Put("/:role_id/:permission_id", service.UpdateRolePermissionService)
	rolePermission.Delete("/:role_id/:permission_id", service.DeleteRolePermissionService)

	lecturer := protected.Group("/v1/lecturers", middleware.RequirePermission(db, "user:manage"))
	lecturer.Get("/", service.GetAllLecturersService)
	lecturer.Get("/:id", service.GetLecturerByIDService)
	lecturer.Post("/", service.CreateLecturerService)
	lecturer.Put("/:id", service.UpdateLecturerService)
	lecturer.Delete("/:id", service.DeleteLecturerService)

	student := protected.Group("/v1/students", middleware.RequirePermission(db, "user:manage"))
	student.Get("/", service.GetAllStudentsService)
	student.Get("/:id", service.GetStudentByIDService)
	student.Post("/", service.CreateStudentService)
	student.Put("/:id", service.UpdateStudentService)
	student.Delete("/:id", service.DeleteStudentService)

	achievements := protected.Group("/v1/achievements")
	achievements.Post("/", middleware.RequirePermission(db, "achievement:create"), service.CreateAchievementService)
	achievements.Put("/:id/submit", middleware.RequirePermission(db, "achievement:update"), service.SubmitAchievementService)
	achievements.Put("/:id/soft-delete", middleware.RequirePermission(db, "achievement:delete"), service.SoftDeleteAchievementService)
	achievements.Put("/:id/review", middleware.RequirePermission(db, "achievement:verify"), service.ReviewAchievementService)
	achievements.Delete("/:id/delete", middleware.RequirePermission(db, "user:manage"), service.HardDeleteAchievementService)
	achievements.Get("/", middleware.RequirePermission(db, "achievement:read"), service.GetAchievementsService)

	achievementRefs := protected.Group("/v1/achievement-references")
	achievementRefs.Get("/", middleware.RequirePermission(db, "achievement:read"), service.GetAchievementReferencesService)
}
