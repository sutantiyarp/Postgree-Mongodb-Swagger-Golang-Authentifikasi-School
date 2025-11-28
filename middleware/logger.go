package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// LoggerMiddleware logs requests
func LoggerMiddleware(c *fiber.Ctx) error {
    // Log request details (example)
    return c.Next()
}