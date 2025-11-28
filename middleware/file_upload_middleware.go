package middleware

import (

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// FileUploadOwnerMiddleware memastikan user hanya bisa upload untuk alumni miliknya atau admin bisa upload untuk siapa saja
func FileUploadOwnerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Identitas dari middleware auth (user_id & role_id sudah ada di Locals)
		userID, _ := c.Locals("user_id").(bson.ObjectID)
		roleID, _ := c.Locals("role_id").(bson.ObjectID)

		// Jika tidak ada user_id atau role_id (user tidak terautentikasi), return Unauthorized
		if userID == bson.NilObjectID || roleID == bson.NilObjectID {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// Opsional: Kirimkan user_id dan role_id ke handler berikutnya jika dibutuhkan
		c.Locals("uploader_user_id", userID)
		c.Locals("uploader_role_id", roleID)

		// Tidak perlu lagi cek admin atau validasi alumni_id karena semua role boleh upload
		return c.Next()
	}
}
