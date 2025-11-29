package service

import (
	"strings"
	"hello-fiber/app/repository"
	"hello-fiber/app/model"
	"github.com/gofiber/fiber/v2"
	"database/sql"
)

var roleRepo repository.RoleRepository

func InitRepoService(db *sql.DB) {
    roleRepo = repository.NewRoleRepositoryPostgres(db)
}

// GetAllRolesService godoc
// @Summary Dapatkan semua role (Admin only)
// @Description Mengambil daftar semua role dengan pagination
// @Tags Roles
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Success 200 {object} model.RoleListResponse "Role list berhasil diambil"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /roles [get]
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
// @Summary Dapatkan detail role (Admin only)
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
// @Router /roles/{id} [get]
// @Security BearerAuth
func GetRoleByIDService(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
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
// @Summary Dapatkan detail role by name (Admin only)
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
// @Router /roles/byname [get]
func GetRoleByNameService(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Query 'name' harus diisi"})
	}

	role, err := roleRepo.GetRoleByName(name)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{"success": false, "message": "Role tidak ditemukan"})
		}
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal mengambil data role", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Data role berhasil diambil", "data": role})
}