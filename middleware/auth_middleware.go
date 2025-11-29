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
func JWTAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
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
				"error":  "Invalid or expired token",
				"detail": err.Error(),
			})
		}

		claims, ok := token.Claims.(*utils.Claims)
		if !ok || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Simpan ke context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role_id", claims.RoleID)

		return c.Next()
	}
}

// AdminOnlyMiddleware
// Mengambil role_id dari Locals (hasil JWTAuthMiddleware)
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
