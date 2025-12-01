package service

import (
	"database/sql"
	"strings"

	"hello-fiber/app/model"
	"hello-fiber/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var studentRepo repository.StudentRepository

func InitStudentService(db *sql.DB) {
	studentRepo = repository.NewStudentRepositoryPostgres(db)
}

func toStudentResponse(s *model.Student) *model.StudentResponse {
	if s == nil {
		return nil
	}
	return &model.StudentResponse{
		ID:           s.ID,
		UserID:       s.UserID,
		StudentID:    s.StudentID,
		ProgramStudy: s.ProgramStudy,
		AcademicYear: s.AcademicYear,
		AdvisorID:    s.AdvisorID,
		CreatedAt:    s.CreatedAt,
	}
}

// GetAllStudentsService godoc
// @Summary Dapatkan semua students (Admin)
// @Description Mengambil daftar semua students dengan pagination
// @Tags Students
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Success 200 {object} map[string]interface{} "Data student berhasil diambil"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/students [get]
// @Security BearerAuth
func GetAllStudentsService(c *fiber.Ctx) error {
	page := int64(c.QueryInt("page", 1))
	limit := int64(c.QueryInt("limit", 10))

	data, total, err := studentRepo.GetAllStudents(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data student",
			"error":   err.Error(),
		})
	}

	var resp []model.StudentResponse
	for i := range data {
		resp = append(resp, *toStudentResponse(&data[i]))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data student berhasil diambil",
		"data":    resp,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetStudentByIDService godoc
// @Summary Dapatkan students by ID (Admin)
// @Description Mengambil detail students berdasarkan id (UUID)
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Success 200 {object} map[string]interface{} "Data student berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Student tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/students/{id} [get]
// @Security BearerAuth
func GetStudentByIDService(c *fiber.Ctx) error {
	id := normParam(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Student ID harus diisi",
		})
	}
	if _, err := uuid.Parse(id); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format Student ID tidak valid",
		})
	}

	st, err := studentRepo.GetStudentByID(id)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Student tidak ditemukan",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data student",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data student berhasil diambil",
		"data":    toStudentResponse(st),
	})
}

// CreateStudentService godoc
// @Summary Buat students (Admin)
// @Description Membuat data students baru
// @Tags Students
// @Accept json
// @Produce json
// @Param body body model.CreateStudentRequest true "Data student"
// @Success 201 {object} model.SuccessResponse "Student berhasil dibuat"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/students [post]
// @Security BearerAuth
func CreateStudentService(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	req.StudentID = strings.TrimSpace(req.StudentID)
	req.ProgramStudy = strings.TrimSpace(req.ProgramStudy)
	req.AcademicYear = strings.TrimSpace(req.AcademicYear)

	if req.UserID == uuid.Nil || req.StudentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "user_id dan student_id harus diisi",
		})
	}

	id, err := studentRepo.CreateStudent(req)
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
			"message": "Gagal membuat student",
			"error":   err.Error(),
		})
	}

	return c.Status(201).JSON(model.SuccessResponse{
		Success: true,
		Message: "Student berhasil dibuat",
		ID:      id,
	})
}

// UpdateStudentService godoc
// @Summary Update students (Admin)
// @Description Update students by id (partial update). Untuk hapus advisor_id, kirim advisor_id = "00000000-0000-0000-0000-000000000000"
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Param body body model.UpdateStudentRequest true "Field yang ingin diupdate"
// @Success 200 {object} model.SuccessResponse "Student berhasil diupdate"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Student tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/students/{id} [put]
// @Security BearerAuth
func UpdateStudentService(c *fiber.Ctx) error {
	id := normParam(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Student ID harus diisi",
		})
	}
	if _, err := uuid.Parse(id); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format Student ID tidak valid",
		})
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	if req.StudentID == nil && req.ProgramStudy == nil && req.AcademicYear == nil && req.AdvisorID == nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Minimal satu field harus diisi untuk update",
		})
	}

	if err := studentRepo.UpdateStudent(id, req); err != nil {
		l := strings.ToLower(err.Error())
		if strings.Contains(l, "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Student tidak ditemukan",
			})
		}
		if strings.Contains(l, "tidak ada field") ||
			strings.Contains(l, "tidak boleh kosong") ||
			strings.Contains(l, "sudah digunakan") ||
			strings.Contains(l, "tidak valid") {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate student",
			"error":   err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Success: true,
		Message: "Student berhasil diupdate",
	})
}

// DeleteStudentService godoc
// @Summary Hapus students (Admin)
// @Description Menghapus students by id
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Success 200 {object} model.SuccessResponse "Student berhasil dihapus"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Student tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/students/{id} [delete]
// @Security BearerAuth
func DeleteStudentService(c *fiber.Ctx) error {
	id := normParam(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Student ID harus diisi",
		})
	}
	if _, err := uuid.Parse(id); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format Student ID tidak valid",
		})
	}

	if err := studentRepo.DeleteStudent(id); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Student tidak ditemukan",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus student",
			"error":   err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Success: true,
		Message: "Student berhasil dihapus",
	})
}
