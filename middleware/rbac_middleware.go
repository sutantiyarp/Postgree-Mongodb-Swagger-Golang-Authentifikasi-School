package middleware

// import (
// 	"database/sql"
// 	"strings"

// 	"hello-fiber/app/repository"

// 	"github.com/gofiber/fiber/v2"
// )

// RequirePermission memeriksa apakah role user memiliki permission tertentu.
// Gunakan setelah JWTAuthMiddleware agar role_id tersedia di Locals.
// func RequirePermission(db *sql.DB, permName string) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		roleIDVal := c.Locals("role_id")
// 		roleID, ok := roleIDVal.(string)
// 		if roleIDVal == nil || !ok || strings.TrimSpace(roleID) == "" {
// 			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 				"error": "Access denied. Role not found",
// 			})
// 		}

// 		rpRepo := repository.NewRolePermissionRepositoryPostgres(db)
// 		perms, err := rpRepo.GetPermissionsByRoleID(roleID)
// 		if err != nil {
// 			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 				"error": "Failed to load permissions",
// 			})
// 		}

// 		for _, p := range perms {
// 			// super permission
// 			if strings.EqualFold(p.Name, "user:manage") {
// 				return c.Next()
// 			}
// 			if strings.EqualFold(p.Name, permName) {
// 				return c.Next()
// 			}
// 		}

// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error": "Access denied. Permission required: " + permName,
// 		})
// 	}
// }