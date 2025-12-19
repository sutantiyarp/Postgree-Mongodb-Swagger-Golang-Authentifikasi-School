package middleware

import (
	"database/sql"
	"fmt"
	"strings"

	"hello-fiber/app/repository"
	"hello-fiber/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware
// Membaca header Authorization: Bearer <token>
// Verifikasi JWT menggunakan utils.Claims dan utils.GetJWTSecret()
// Simpan user_id, email, role_id ke Locals untuk dipakai handler / middleware lain
func JWTAuthMiddleware(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header dibutuhkan",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Parse dan verifikasi token JWT
		token, err := jwt.ParseWithClaims(tokenString, &utils.Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return utils.GetJWTSecret(), nil
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":  "Invalid atau expired token",
				"detail": err.Error(),
			})
		}

		claims, ok := token.Claims.(*utils.Claims)
		if !ok || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		userRepo := repository.NewUserRepositoryPostgres(db)
		user, err := userRepo.GetUserByID(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user tidak ditemukan",
			})
		}

		if !user.IsActive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token di matikan, akun tidak aktif",
			})
		}

		// Simpan ke context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role_id", claims.RoleID)
		if len(claims.Permissions) > 0 {
			c.Locals("permissions", claims.Permissions)
		}

		return c.Next()
	}
}

// AdminOnlyMiddleware
func AdminOnlyMiddleware(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleIDVal := c.Locals("role_id")
		if roleIDVal == nil {
			fmt.Printf("[DEBUG] role_id not found in locals\n")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Admin role required",
			})
		}

		roleIDStr, ok := roleIDVal.(string)
		if !ok || roleIDStr == "" {
			fmt.Printf("[DEBUG] role_id type assertion failed or empty\n")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Invalid role information",
			})
		}

		roleRepo := repository.NewRoleRepositoryPostgres(db)
		adminRole, err := roleRepo.GetRoleByName("Admin")
		if err != nil {
			fmt.Printf("[DEBUG] Admin role not found: %v\n", err)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Admin role not found in database",
			})
		}

		fmt.Printf("[DEBUG] User roleID: %s, Admin roleID: %s\n", roleIDStr, adminRole.ID)

		if roleIDStr != adminRole.ID {
			fmt.Printf("[DEBUG] Role mismatch - User has: %s, Admin is: %s\n", roleIDStr, adminRole.ID)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Admin role required",
			})
		}

		fmt.Printf("[DEBUG] Admin verification passed!\n")
		return c.Next()
	}
}

// StudentOnlyMiddleware menyimpan student_id ke context untuk handler berikutnya
func StudentOnlyMiddleware(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDVal := c.Locals("user_id")
		if userIDVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		userIDStr, ok := userIDVal.(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		studentRepo := repository.NewStudentRepositoryPostgres(db)
		st, err := studentRepo.GetStudentByUserID(userIDStr)
		if err != nil || st == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Hanya mahasiswa yang dapat mengakses",
			})
		}

		// Simpan ke context
		c.Locals("student_id", st.ID.String())
		c.Locals("student_uuid", st.ID)
		c.Locals("student_user_id", st.UserID)

		return c.Next()
	}
}

func RequirePermission(db *sql.DB, permName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Cek permissions dari JWT (cache)
		if permsVal := c.Locals("permissions"); permsVal != nil {
			if permSlice, ok := permsVal.([]string); ok && len(permSlice) > 0 {
				for _, p := range permSlice {
					if strings.EqualFold(p, "user:manage") || strings.EqualFold(p, permName) {
						return c.Next()
					}
				}
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Access denied. Permission required: " + permName,
				})
			}
		}

		roleIDVal := c.Locals("role_id")
		roleID, ok := roleIDVal.(string)
		if roleIDVal == nil || !ok || strings.TrimSpace(roleID) == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Role not found",
			})
		}

		rpRepo := repository.NewRolePermissionRepositoryPostgres(db)
		perms, err := rpRepo.GetPermissionsByRoleID(roleID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to load permissions",
			})
		}

		for _, p := range perms {
			// super permission
			if strings.EqualFold(p.Name, "user:manage") {
				return c.Next()
			}
			if strings.EqualFold(p.Name, permName) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied. Permission required: " + permName,
		})
	}
}
