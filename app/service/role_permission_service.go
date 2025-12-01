package service

import (
	"database/sql"
	"net/url"
	"strings"

	"hello-fiber/app/model"
	"hello-fiber/app/repository"

	"github.com/gofiber/fiber/v2"
)

var rolePermissionRepo repository.RolePermissionRepository

func InitRolePermissionService(db *sql.DB) {
	rolePermissionRepo = repository.NewRolePermissionRepositoryPostgres(db)
}

func normParam(raw string) string {
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		decoded = raw
	}
	return strings.TrimSpace(decoded)
}

// GetAllRolePermissionsService godoc
// @Summary Dapatkan semua role_permission (Admin)
// @Description Mengambil daftar mapping role_id dan permission_id dengan pagination dan filter opsional
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Param role_id query string false "Filter role_id (UUID)"
// @Param permission_id query string false "Filter permission_id (UUID)"
// @Success 200 {object} map[string]interface{} "Data role_permission berhasil diambil"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/role-permissions [get]
// @Security BearerAuth
func GetAllRolePermissionsService(c *fiber.Ctx) error {
	page := int64(c.QueryInt("page", 1))
	limit := int64(c.QueryInt("limit", 10))
	roleID := strings.TrimSpace(c.Query("role_id"))
	permissionID := strings.TrimSpace(c.Query("permission_id"))

	data, total, err := rolePermissionRepo.GetAllRolePermissions(page, limit, roleID, permissionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data role_permission",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data role_permission berhasil diambil",
		"data":    data,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetRolePermissionDetailService godoc
// @Summary Dapatkan detail role_permission (Admin)
// @Description Mengambil 1 mapping role_id + permission_id (composite key)
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID (UUID)"
// @Param permission_id path string true "Permission ID (UUID)"
// @Success 200 {object} map[string]interface{} "Detail role_permission berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "role_permission tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/role-permissions/{role_id}/{permission_id} [get]
// @Security BearerAuth
// func GetRolePermissionDetailService(c *fiber.Ctx) error {
// 	roleID := normParam(c.Params("role_id"))
// 	permissionID := normParam(c.Params("permission_id"))

// 	if roleID == "" || permissionID == "" {
// 		return c.Status(400).JSON(fiber.Map{
// 			"success": false,
// 			"message": "role_id dan permission_id harus diisi",
// 		})
// 	}

// 	rp, err := rolePermissionRepo.GetRolePermission(roleID, permissionID)
// 	if err != nil {
// 		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
// 			return c.Status(404).JSON(fiber.Map{
// 				"success": false,
// 				"message": "role_permission tidak ditemukan",
// 			})
// 		}
// 		return c.Status(500).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil detail role_permission",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Detail role_permission berhasil diambil",
// 		"data":    rp,
// 	})
// }

// GetPermissionsByRoleIDService godoc
// @Summary Dapatkan daftar permissions milik role (Admin)
// @Description Mengambil list permission yang dimiliki oleh role tertentu
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID (UUID)"
// @Success 200 {object} map[string]interface{} "Data permissions milik role berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/role-permissions/byrole/{role_id} [get]
// @Security BearerAuth
func GetPermissionsByRoleIDService(c *fiber.Ctx) error {
	roleID := normParam(c.Params("role_id"))
	if roleID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "role_id harus diisi",
		})
	}

	perms, err := rolePermissionRepo.GetPermissionsByRoleID(roleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil permissions milik role",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data permissions milik role berhasil diambil",
		"data":    perms,
	})
}

// CreateRolePermissionService godoc
// @Summary Tambah role_permission (Admin)
// @Description Membuat mapping role_id dan permission_id
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param body body model.CreateRolePermissionRequest true "Data role_permission"
// @Success 201 {object} model.SuccessResponse "role_permission berhasil dibuat"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal / data sudah ada / foreign key tidak valid"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/role-permissions [post]
// @Security BearerAuth
func CreateRolePermissionService(c *fiber.Ctx) error {
	var req model.CreateRolePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	req.RoleID = strings.TrimSpace(req.RoleID)
	req.PermissionID = strings.TrimSpace(req.PermissionID)
	if req.RoleID == "" || req.PermissionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "role_id dan permission_id harus diisi",
		})
	}

	if err := rolePermissionRepo.CreateRolePermission(req.RoleID, req.PermissionID); err != nil {
		l := strings.ToLower(err.Error())
		code := 500
		if strings.Contains(l, "sudah ada") || strings.Contains(l, "tidak valid") {
			code = 400
		}
		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "role_permission berhasil dibuat",
	})
}

// UpdateRolePermissionService godoc
// @Summary Update role_permission (Admin)
// @Description Update composite key mapping (role_id, permission_id) menjadi (new_role_id, new_permission_id)
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID lama (UUID)"
// @Param permission_id path string true "Permission ID lama (UUID)"
// @Param body body model.UpdateRolePermissionRequest true "Data role_permission baru"
// @Success 200 {object} model.SuccessResponse "role_permission berhasil diupdate"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal / data sudah ada / foreign key tidak valid"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "role_permission tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/role-permissions/{role_id}/{permission_id} [put]
// @Security BearerAuth
func UpdateRolePermissionService(c *fiber.Ctx) error {
	oldRoleID := normParam(c.Params("role_id"))
	oldPermissionID := normParam(c.Params("permission_id"))

	if oldRoleID == "" || oldPermissionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "role_id dan permission_id harus diisi",
		})
	}

	var req model.UpdateRolePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	newRoleID := strings.TrimSpace(req.NewRoleID)
	newPermissionID := strings.TrimSpace(req.NewPermissionID)

	if newRoleID == "" && newPermissionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Minimal salah satu dari new_role_id atau new_permission_id harus diisi",
		})
	}
	if newRoleID == "" {
		newRoleID = oldRoleID
	}
	if newPermissionID == "" {
		newPermissionID = oldPermissionID
	}

	if err := rolePermissionRepo.UpdateRolePermission(oldRoleID, oldPermissionID, newRoleID, newPermissionID); err != nil {
		l := strings.ToLower(err.Error())
		if strings.Contains(l, "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "role_permission tidak ditemukan",
			})
		}
		if strings.Contains(l, "sudah ada") || strings.Contains(l, "tidak valid") {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal update role_permission",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "role_permission berhasil diupdate",
	})
}

// DeleteRolePermissionService godoc
// @Summary Hapus role_permission (Admin)
// @Description Menghapus mapping role_id dan permission_id
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID (UUID)"
// @Param permission_id path string true "Permission ID (UUID)"
// @Success 200 {object} model.SuccessResponse "role_permission berhasil dihapus"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "role_permission tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/role-permissions/{role_id}/{permission_id} [delete]
// @Security BearerAuth
func DeleteRolePermissionService(c *fiber.Ctx) error {
	roleID := normParam(c.Params("role_id"))
	permissionID := normParam(c.Params("permission_id"))

	if roleID == "" || permissionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "role_id dan permission_id harus diisi",
		})
	}

	if err := rolePermissionRepo.DeleteRolePermission(roleID, permissionID); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "role_permission tidak ditemukan",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus role_permission",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "role_permission berhasil dihapus",
	})
}
