package service

import (
	"database/sql"
	"net/url"
	"strings"
	"hello-fiber/app/model"
	"hello-fiber/app/repository"
	"github.com/gofiber/fiber/v2"
)

var permissionRepo repository.PermissionRepository

func InitPermissionService(db *sql.DB) {
    permissionRepo = repository.NewPermissionRepositoryPostgres(db)
}

func normalizePathParam(raw string) string {
    if raw == "" {
        return ""
    }
    decoded, err := url.PathUnescape(raw)
    if err != nil {
        decoded = raw
    }
    return strings.TrimSpace(decoded)
}

// GetAllPermissionsService godoc
// @Summary Dapatkan semua permission (Permission: user:manage)
// @Description Mengambil daftar semua permission dengan pagination
// @Tags Permissions
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Success 200 {object} map[string]interface{} "Data permission berhasil diambil"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/permissions [get]
// @Security BearerAuth
func GetAllPermissionsService(c *fiber.Ctx) error {
	page := int64(1)
	limit := int64(10)

	if p := c.Query("page"); p != "" {
		page = int64(c.QueryInt("page", 1))
	}
	if l := c.Query("limit"); l != "" {
		limit = int64(c.QueryInt("limit", 10))
	}

	permissions, total, err := permissionRepo.GetAllPermissions(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data permission",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data permission berhasil diambil",
		"data":    permissions,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetPermissionByIDService godoc
// @Summary Dapatkan detail permission (Permission: user:manage)
// @Description Mengambil detail permission berdasarkan ID
// @Tags Permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID (UUID)"
// @Success 200 {object} map[string]interface{} "Data permission berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Permission tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/permissions/{id} [get]
// @Security BearerAuth
func GetPermissionByIDService(c *fiber.Ctx) error {
	id := normalizePathParam(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Permission ID harus diisi",
		})
	}

	perm, err := permissionRepo.GetPermissionByID(id)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Permission tidak ditemukan",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data permission",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data permission berhasil diambil",
		"data":    perm,
	})
}

// CreatePermissionService godoc
// @Summary Buat permission baru (Permission: user:manage)
// @Description Memerlukan permission user:manage untuk membuat permission baru
// @Tags Permissions
// @Accept json
// @Produce json
// @Param body body model.CreatePermissionRequest true "Data permission yang akan dibuat"
// @Success 201 {object} model.SuccessResponse "Permission berhasil dibuat"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/permissions [post]
// @Security BearerAuth
func CreatePermissionService(c *fiber.Ctx) error {
	var req model.CreatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Resource = strings.TrimSpace(req.Resource)
	req.Action = strings.TrimSpace(req.Action)
	req.Description = strings.TrimSpace(req.Description)

	if req.Name == "" || req.Resource == "" || req.Action == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Name, resource, dan action harus diisi",
		})
	}

	id, err := permissionRepo.CreatePermission(req)
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "sudah ada") || strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat permission",
			"error":   err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Permission berhasil dibuat",
		"id":      id,
	})
}

// UpdatePermissionService godoc
// @Summary Update permission (Permission: user:manage)
// @Description Memerlukan permission user:manage untuk mengupdate data permission berdasarkan ID
// @Tags Permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID (UUID)"
// @Param body body model.UpdatePermissionRequest true "Data permission yang akan diupdate"
// @Success 200 {object} model.SuccessResponse "Permission berhasil diupdate"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Permission tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/permissions/{id} [put]
// @Security BearerAuth
func UpdatePermissionService(c *fiber.Ctx) error {
	id := normalizePathParam(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Permission ID harus diisi",
		})
	}

	var req model.UpdatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	if strings.TrimSpace(req.Name) == "" &&
		strings.TrimSpace(req.Resource) == "" &&
		strings.TrimSpace(req.Action) == "" &&
		strings.TrimSpace(req.Description) == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Minimal satu field harus diisi untuk update",
		})
	}

	if err := permissionRepo.UpdatePermission(id, req); err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Permission tidak ditemukan",
			})
		}
		if strings.Contains(lower, "sudah ada") || strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate permission",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Permission berhasil diupdate",
	})
}

// DeletePermissionService godoc
// @Summary Hapus permission (Permission: user:manage)
// @Description Memerlukan permission user:manage untuk menghapus permission berdasarkan ID
// @Tags Permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID (UUID)"
// @Success 200 {object} model.SuccessResponse "Permission berhasil dihapus"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Permission tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/permissions/{id} [delete]
// @Security BearerAuth
func DeletePermissionService(c *fiber.Ctx) error {
	id := normalizePathParam(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Permission ID harus diisi",
		})
	}

	if err := permissionRepo.DeletePermission(id); err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Permission tidak ditemukan",
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus permission",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Permission berhasil dihapus",
	})
}
