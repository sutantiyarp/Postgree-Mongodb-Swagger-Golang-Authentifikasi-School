package service

import (
	"database/sql"
	"strings"

	"hello-fiber/app/model"
	"hello-fiber/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var lecturerRepo repository.LecturerRepository

func InitLecturerService(db *sql.DB) {
	lecturerRepo = repository.NewLecturerRepositoryPostgres(db)
}

func toLecturerResponse(l *model.Lecturer) *model.LecturerResponse {
	if l == nil {
		return nil
	}
	return &model.LecturerResponse{
		ID:         l.ID,
		UserID:     l.UserID,
		LecturerID: l.LecturerID,
		Department: l.Department,
		CreatedAt:  l.CreatedAt,
	}
}

// GetAllLecturersService godoc
// @Summary Dapatkan semua lecturer (Permission: user:manage)
// @Description Mengambil daftar semua lecturer dengan pagination
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Success 200 {object} map[string]interface{} "Data lecturer berhasil diambil"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/lecturers [get]
// @Security BearerAuth
func GetAllLecturersService(c *fiber.Ctx) error {
	page := int64(c.QueryInt("page", 1))
	limit := int64(c.QueryInt("limit", 10))

	data, total, err := lecturerRepo.GetAllLecturers(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data lecturer",
			"error":   err.Error(),
		})
	}

	var resp []model.LecturerResponse
	for i := range data {
		resp = append(resp, *toLecturerResponse(&data[i]))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data lecturer berhasil diambil",
		"data":    resp,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetLecturerByIDService godoc
// @Summary Dapatkan lecturer by ID (Permission: user:manage)
// @Description Mengambil detail lecturer berdasarkan id (UUID)
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param id path string true "Lecturer ID (UUID)"
// @Success 200 {object} map[string]interface{} "Data lecturer berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Lecturer tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/lecturers/{id} [get]
// @Security BearerAuth
func GetLecturerByIDService(c *fiber.Ctx) error {
	id := normalizePathParam(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Lecturer ID harus diisi",
		})
	}
	if _, err := uuid.Parse(id); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format Lecturer ID tidak valid",
		})
	}

	lec, err := lecturerRepo.GetLecturerByID(id)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Lecturer tidak ditemukan",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data lecturer",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data lecturer berhasil diambil",
		"data":    toLecturerResponse(lec),
	})
}

// CreateLecturerService godoc
// @Summary Buat lecturer (Permission: user:manage)
// @Description Membuat data lecturer baru
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param body body model.CreateLecturerRequest true "Data lecturer"
// @Success 201 {object} model.SuccessResponse "Lecturer berhasil dibuat"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/lecturers [post]
// @Security BearerAuth
func CreateLecturerService(c *fiber.Ctx) error {
	var req model.CreateLecturerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	req.LecturerID = strings.TrimSpace(req.LecturerID)
	req.Department = strings.TrimSpace(req.Department)

	if req.UserID == uuid.Nil || req.LecturerID == "" || req.Department == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "user_id, lecturer_id, dan department harus diisi",
		})
	}

	id, err := lecturerRepo.CreateLecturer(req)
	if err != nil {
		l := strings.ToLower(err.Error())
		if strings.Contains(l, "sudah digunakan") || strings.Contains(l, "tidak valid") {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat lecturer",
			"error":   err.Error(),
		})
	}

	return c.Status(201).JSON(model.SuccessResponse{
		Success: true,
		Message: "Lecturer berhasil dibuat",
		ID:      id,
	})
}

// UpdateLecturerService godoc
// @Summary Update lecturer (Permission: user:manage)
// @Description Update lecturer by id (partial update)
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param id path string true "Lecturer ID (UUID)"
// @Param body body model.UpdateLecturerRequest true "Field yang ingin diupdate"
// @Success 200 {object} model.SuccessResponse "Lecturer berhasil diupdate"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Lecturer tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/lecturers/{id} [put]
// @Security BearerAuth
func UpdateLecturerService(c *fiber.Ctx) error {
	id := normalizePathParam(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Lecturer ID harus diisi",
		})
	}
	if _, err := uuid.Parse(id); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format Lecturer ID tidak valid",
		})
	}

	var req model.UpdateLecturerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	if req.LecturerID == nil && req.Department == nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Minimal satu field harus diisi untuk update",
		})
	}

	if err := lecturerRepo.UpdateLecturer(id, req); err != nil {
		l := strings.ToLower(err.Error())
		if strings.Contains(l, "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Lecturer tidak ditemukan",
			})
		}
		if strings.Contains(l, "tidak ada field") ||
			strings.Contains(l, "tidak boleh kosong") ||
			strings.Contains(l, "sudah digunakan") {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate lecturer",
			"error":   err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Success: true,
		Message: "Lecturer berhasil diupdate",
	})
}

// DeleteLecturerService godoc
// @Summary Hapus lecturer (Permission: user:manage)
// @Description Menghapus lecturer by id
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param id path string true "Lecturer ID (UUID)"
// @Success 200 {object} model.SuccessResponse "Lecturer berhasil dihapus"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Lecturer tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/lecturers/{id} [delete]
// @Security BearerAuth
func DeleteLecturerService(c *fiber.Ctx) error {
	id := normalizePathParam(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Lecturer ID harus diisi",
		})
	}
	if _, err := uuid.Parse(id); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format Lecturer ID tidak valid",
		})
	}

	if err := lecturerRepo.DeleteLecturer(id); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Lecturer tidak ditemukan",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus lecturer",
			"error":   err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Success: true,
		Message: "Lecturer berhasil dihapus",
	})
}
