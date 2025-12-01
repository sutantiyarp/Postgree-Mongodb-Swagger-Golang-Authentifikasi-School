package service

import (
	"strings"
	"hello-fiber/app/repository"
	"hello-fiber/app/model"
	"github.com/gofiber/fiber/v2"
	"database/sql"
	"net/url"
)

var roleRepo repository.RoleRepository

func InitRepoService(db *sql.DB) {
    roleRepo = repository.NewRoleRepositoryPostgres(db)
}

// GetAllRolesService godoc
// @Summary Dapatkan semua role (Admin)
// @Description Mengambil daftar semua role dengan pagination
// @Tags Roles
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Success 200 {object} model.RoleListResponse "Role list berhasil diambil"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/roles [get]
// @Security BearerAuth
func GetAllRolesService(c *fiber.Ctx) error {
	page := int64(1)
	limit := int64(10)

	if p := c.Query("page"); p != "" {
		page = int64(c.QueryInt("page", 1))
	}
	if l := c.Query("limit"); l != "" {
		limit = int64(c.QueryInt("limit", 10))
	}

	roles, total, err := roleRepo.GetAllRoles(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data role",
			"error":   err.Error(),
		})
	}

	resp := make([]model.Role, 0, len(roles))
	for _, r := range roles {
		resp = append(resp, model.Role{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			CreatedAt:   r.CreatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data role berhasil diambil",
		"data":    resp,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetRoleByIDService godoc
// @Summary Dapatkan detail role (Admin)
// @Description Mengambil detail role berdasarkan Role ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Success 200 {object} model.RoleDetailResponse "Data role berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Role tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/roles/{id} [get]
// @Security BearerAuth
func GetRoleByIDService(c *fiber.Ctx) error {
    rawID := c.Params("id")

    id, _ := url.PathUnescape(rawID)
    id = strings.TrimSpace(id)

    if id == "" {
        return c.Status(400).JSON(fiber.Map{
            "success": false,
            "message": "Role ID harus diisi",
        })
    }

    role, err := roleRepo.GetRoleByID(id)
    if err != nil {
        if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
            return c.Status(404).JSON(fiber.Map{
                "success": false,
                "message": "Role tidak ditemukan",
            })
        }
        return c.Status(500).JSON(fiber.Map{
            "success": false,
            "message": "Gagal mengambil data role",
            "error":   err.Error(),
        })
    }

    if role == nil {
        return c.Status(404).JSON(fiber.Map{
            "success": false,
            "message": "Role tidak ditemukan",
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "Data role berhasil diambil",
        "data": model.Role{
            ID:          role.ID,
            Name:        role.Name,
            Description: role.Description,
            CreatedAt:   role.CreatedAt,
        },
    })
}

// GetRoleByNameService godoc
// @Summary Dapatkan detail role by name (Admin)
// @Description Contoh: /roles/byname?name=Admin
// @Tags Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name query string true "Role name (misal: Admin)"
// @Success 200 {object} model.RoleDetailResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/roles/byname [get]
// func GetRoleByNameService(c *fiber.Ctx) error {
// 	name := strings.TrimSpace(c.Query("name"))
// 	if name == "" {
// 		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Query 'name' harus diisi"})
// 	}

// 	role, err := roleRepo.GetRoleByName(name)
// 	if err != nil {
// 		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
// 			return c.Status(404).JSON(fiber.Map{"success": false, "message": "Role tidak ditemukan"})
// 		}
// 		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal mengambil data role", "error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{"success": true, "message": "Data role berhasil diambil", "data": role})
// }

// CreateRoleService godoc
// @Summary Buat role baru (Admin)
// @Description Admin membuat role baru
// @Tags Roles
// @Accept json
// @Produce json
// @Param body body model.CreateRoleRequest true "Data role baru"
// @Success 201 {object} model.SuccessResponse "Role berhasil dibuat"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/roles [post]
// @Security BearerAuth
func CreateRoleService(c *fiber.Ctx) error {
	var req model.CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Nama role harus diisi",
		})
	}

	if existing, err := roleRepo.GetRoleByName(req.Name); err == nil && existing != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Role dengan nama tersebut sudah ada",
		})
	}

	id, err := roleRepo.CreateRole(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat role",
			"error":   err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Role berhasil dibuat",
		"id":      id,
	})
}

// UpdateRoleService godoc
// @Summary Update role (Admin)
// @Description Admin mengupdate data role berdasarkan ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Param body body model.UpdateRoleRequest true "Data role yang akan diupdate"
// @Success 200 {object} model.SuccessResponse "Role berhasil diupdate"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Role tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/roles/{id} [put]
// @Security BearerAuth
func UpdateRoleService(c *fiber.Ctx) error {
	roleID := c.Params("id")
	if roleID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Role ID harus diisi",
		})
	}

	var req model.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	hasUpdate := strings.TrimSpace(req.Name) != "" || req.Description != ""
	if !hasUpdate {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Minimal ada satu field yang harus diupdate",
		})
	}

	if err := roleRepo.UpdateRole(roleID, req); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Role tidak ditemukan",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal update role",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Role berhasil diupdate",
	})
}

// DeleteRoleService godoc
// @Summary Hapus role (Admin)
// @Description Admin dapat menghapus role berdasarkan ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Success 200 {object} model.SuccessResponse "Role berhasil dihapus"
// @Failure 400 {object} model.ErrorResponse "Role ID tidak valid"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Role tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/roles/{id} [delete]
// @Security BearerAuth
func DeleteRoleService(c *fiber.Ctx) error {
	roleID := c.Params("id")
	if roleID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Role ID harus diisi",
		})
	}

	if err := roleRepo.DeleteRole(roleID); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Role tidak ditemukan",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal delete role",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Role berhasil dihapus",
	})
}
