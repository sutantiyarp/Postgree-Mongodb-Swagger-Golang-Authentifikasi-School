package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"hello-fiber/app/model"
	"hello-fiber/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var achievementMongoRepo repository.AchievementMongoRepository
var achievementRefRepo repository.AchievementReferenceRepository
var achievementRoleRepo repository.RoleRepository
var achievementStudentRepo repository.StudentRepository
var achievementLecturerRepo repository.LecturerRepository

func InitAchievementService(db *sql.DB, mongoDB *mongo.Database) {
	achievementMongoRepo = repository.NewAchievementMongoRepository(mongoDB)
	achievementRefRepo = repository.NewAchievementReferenceRepository(db)
	achievementRoleRepo = repository.NewRoleRepositoryPostgres(db)
	achievementStudentRepo = repository.NewStudentRepositoryPostgres(db)
	achievementLecturerRepo = repository.NewLecturerRepositoryPostgres(db)
}

// parse multipart payload for achievement create, including attachments.
func parseMultipartCreateAchievement(c *fiber.Ctx) (*model.CreateAchievementRequest, error) {
	req := model.CreateAchievementRequest{}

	req.AchievementType = c.FormValue("achievement_type")
	req.Title = c.FormValue("title")
	req.Description = c.FormValue("description")

	if detailsStr := c.FormValue("details"); detailsStr != "" {
		var det map[string]interface{}
		if err := json.Unmarshal([]byte(detailsStr), &det); err != nil {
			return nil, fmt.Errorf("details harus JSON: %w", err)
		}
		req.Details = det
	}

	if tagsStr := c.FormValue("tags"); tagsStr != "" {
		tags := []string{}
		for _, t := range strings.Split(tagsStr, ",") {
			if v := strings.TrimSpace(t); v != "" {
				tags = append(tags, v)
			}
		}
		req.Tags = tags
	}

	if pointsStr := c.FormValue("points"); pointsStr != "" {
		p, err := strconv.ParseFloat(pointsStr, 64)
		if err != nil {
			return nil, fmt.Errorf("points harus numerik: %w", err)
		}
		req.Points = &p
	}

	form, err := c.MultipartForm()
	if err == nil && form != nil && form.File != nil {
		files := form.File["attachments"]
		if len(files) > 0 {
			if err := os.MkdirAll("uploads", 0o755); err != nil {
				return nil, fmt.Errorf("gagal buat folder uploads: %w", err)
			}
			for _, fh := range files {
				if fh.Size > 7*1024*1024 {
					return nil, fmt.Errorf("ukuran file maksimal 7MB")
				}
				ext := strings.ToLower(filepath.Ext(fh.Filename))
				ctype := fh.Header.Get("Content-Type")
				if ext != ".pdf" && !strings.EqualFold(strings.ToLower(ctype), "application/pdf") {
					return nil, fmt.Errorf("hanya file PDF yang diperbolehkan")
				}
				storedName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(fh.Filename))
				savePath := filepath.Join("uploads", storedName)
				if err := c.SaveFile(fh, savePath); err != nil {
					return nil, fmt.Errorf("gagal simpan file %s: %w", fh.Filename, err)
				}
				fileType := ctype
				if fileType == "" {
					fileType = strings.TrimPrefix(ext, ".")
				}
				req.Attachments = append(req.Attachments, model.Attachment{
					FileName:   fh.Filename,
					FileURL:    "/" + filepath.ToSlash(savePath),
					FileType:   fileType,
					UploadedAt: time.Now(),
				})
			}
		}
	}

	return &req, nil
}

// normalizeDetails memastikan tipe data sesuai schema Mongo (misal rank harus int).
func normalizeDetails(achType string, details map[string]interface{}) (map[string]interface{}, error) {
	if details == nil {
		return map[string]interface{}{}, nil
	}
	achType = strings.ToLower(strings.TrimSpace(achType))

	// Clone map untuk menghindari mutasi referensi caller
	out := make(map[string]interface{}, len(details))
	for k, v := range details {
		out[k] = v
	}

	if achType == "competition" {
		if rankVal, ok := out["rank"]; ok {
			switch rv := rankVal.(type) {
			case float64:
				out["rank"] = int(rv)
			case int32, int64, int:
				// sudah OK
			default:
				return nil, fmt.Errorf("rank harus numerik (int)")
			}
		}
		if level, ok := out["competitionLevel"].(string); ok {
			out["competitionLevel"] = strings.ToLower(strings.TrimSpace(level))
		}
	}

	return out, nil
}

func resolveRoleName(c *fiber.Ctx) (string, error) {
	roleIDVal := c.Locals("role_id")
	roleID, ok := roleIDVal.(string)
	if roleIDVal == nil || !ok || strings.TrimSpace(roleID) == "" {
		return "", fmt.Errorf("role tidak ditemukan")
	}
	role, err := achievementRoleRepo.GetRoleByID(roleID)
	if err != nil || role == nil {
		return "", fmt.Errorf("role tidak ditemukan")
	}
	return strings.ToLower(strings.TrimSpace(role.Name)), nil
}

// allowedStatusesByRole menentukan status apa saja yang boleh diakses.
// jika forAchievements=true dan role mahasiswa, filter juga ke student_id miliknya.
// untuk dosen wali, filter ke advisor_id (lecturer) yang sesuai.
func allowedStatusesByRole(c *fiber.Ctx, roleName string, forAchievements bool) ([]string, *uuid.UUID, *uuid.UUID, error) {
	roleName = strings.ToLower(strings.TrimSpace(roleName))

	switch roleName {
	case "admin":
		return []string{
			model.AchievementStatusDraft,
			model.AchievementStatusSubmitted,
			model.AchievementStatusVerified,
			model.AchievementStatusRejected,
			model.AchievementStatusDeleted,
		}, nil, nil, nil
	case "mahasiswa":
		studentUUID, ok := c.Locals("student_uuid").(uuid.UUID)
		if !ok {
			userIDVal := c.Locals("user_id")
			userIDStr, okUser := userIDVal.(string)
			if !okUser || strings.TrimSpace(userIDStr) == "" {
				return nil, nil, nil, fmt.Errorf("mahasiswa tidak memiliki student_id")
			}
			st, err := achievementStudentRepo.GetStudentByUserID(userIDStr)
			if err != nil || st == nil {
				return nil, nil, nil, fmt.Errorf("mahasiswa tidak memiliki student_id")
			}
			studentUUID = st.ID
			// cache ke context untuk request selanjutnya
			c.Locals("student_uuid", studentUUID)
		}
		return []string{
			model.AchievementStatusDraft,
			model.AchievementStatusSubmitted,
			model.AchievementStatusVerified,
			model.AchievementStatusRejected,
		}, &studentUUID, nil, nil
	case "dosen wali":
		userIDVal := c.Locals("user_id")
		userIDStr, ok := userIDVal.(string)
		if userIDVal == nil || !ok || strings.TrimSpace(userIDStr) == "" {
			return nil, nil, nil, fmt.Errorf("user tidak valid")
		}
		if cachedLect, ok := c.Locals("lecturer_uuid").(uuid.UUID); ok {
			return []string{model.AchievementStatusSubmitted}, nil, &cachedLect, nil
		}
		lect, err := achievementLecturerRepo.GetLecturerByUserID(userIDStr)
		if err != nil || lect == nil {
			return nil, nil, nil, fmt.Errorf("dosen wali tidak ditemukan")
		}
		c.Locals("lecturer_uuid", lect.ID)
		return []string{model.AchievementStatusSubmitted}, nil, &lect.ID, nil
	case "staff":
		return []string{model.AchievementStatusVerified, model.AchievementStatusRejected}, nil, nil, nil
	default:
		return nil, nil, nil, fmt.Errorf("role tidak diperbolehkan")
	}
}

// CreateAchievementService godoc
// @Summary Mahasiswa membuat achievement (Mongo) + reference draft (Postgres)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param body body model.CreateAchievementRequest true "Data achievement (details mengikuti achievement_type). oneOf examples: competition {competitionName, competitionLevel, rank}, publication {publicationType, publicationTitle, authors, publisher, issn}, organization {organizationName, position, periodStart, periodEnd}, certification {certificationName, issuedBy, certificationNumber, validUntil}, academic {description, score}."
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/achievements [post]
// @Security BearerAuth
func CreateAchievementService(c *fiber.Ctx) error {
	studentUUID, ok := c.Locals("student_uuid").(uuid.UUID)
	if !ok {
		userIDVal := c.Locals("user_id")
		userID, ok := userIDVal.(string)
		if userIDVal == nil || !ok || strings.TrimSpace(userID) == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "User tidak valid",
			})
		}
		st, err := achievementStudentRepo.GetStudentByUserID(userID)
		if err != nil || st == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "mahasiswa tidak memiliki student_id",
			})
		}
		studentUUID = st.ID
		c.Locals("student_uuid", studentUUID)
	}

	var req model.CreateAchievementRequest
	ct := strings.ToLower(c.Get("Content-Type"))
	if strings.HasPrefix(ct, "multipart/") {
		parsed, err := parseMultipartCreateAchievement(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		req = *parsed
	} else {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Request body tidak valid",
				"error":   err.Error(),
			})
		}
	}

	req.AchievementType = strings.ToLower(strings.TrimSpace(req.AchievementType))
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.AchievementType == "" || req.Title == "" || req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "achievement_type, title, dan description wajib diisi",
		})
	}

	normalizedDetails, err := normalizeDetails(req.AchievementType, req.Details)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}
	req.Details = normalizedDetails

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoID, err := achievementMongoRepo.Create(ctx, studentUUID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan achievement",
			"error":   err.Error(),
		})
	}

	refID, err := achievementRefRepo.CreateDraft(ctx, studentUUID, mongoID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat reference",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Achievement berhasil dibuat",
		"data": fiber.Map{
			"reference_id":         refID,
			"mongo_achievement_id": mongoID,
			"status":               model.AchievementStatusDraft,
		},
	})
}

// SubmitAchievementService godoc
// @Summary Mahasiswa submit achievement (draft -> submitted)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement reference ID (UUID)"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/achievements/{id}/submit [put]
// @Security BearerAuth
func SubmitAchievementService(c *fiber.Ctx) error {
	refID := strings.TrimSpace(c.Params("id"))
	if refID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID reference harus diisi",
		})
	}

	studentUUID, ok := c.Locals("student_uuid").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya mahasiswa yang dapat mengakses",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := achievementRefRepo.SubmitDraft(ctx, refID, studentUUID); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "tidak ditemukan") || strings.Contains(msg, "bukan milik") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Success: true,
		Message: "Status achievement berubah ke submitted",
	})
}

// ReviewAchievementService godoc
// @Summary Dosen review achievement (submitted -> verified/rejected)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement reference ID (UUID)"
// @Param body body model.UpdateAchievementStatusRequest true "Status baru"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/achievements/{id}/review [put]
// @Security BearerAuth
func ReviewAchievementService(c *fiber.Ctx) error {
	refID := strings.TrimSpace(c.Params("id"))
	if refID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID reference harus diisi",
		})
	}

	var req model.UpdateAchievementStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Request body tidak valid",
			"error":   err.Error(),
		})
	}

	req.Status = strings.ToLower(strings.TrimSpace(req.Status))

	roleName, err := resolveRoleName(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	userIDVal := c.Locals("user_id")
	userIDStr, ok := userIDVal.(string)
	if userIDVal == nil || !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}
	actorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch roleName {
	case "admin":
		if req.Status == model.AchievementStatusRejected {
			if req.RejectionNote == nil || strings.TrimSpace(*req.RejectionNote) == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "rejection_note wajib diisi jika status rejected",
				})
			}
		}
		if req.Status != model.AchievementStatusVerified &&
			req.Status != model.AchievementStatusRejected {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Status harus verified/rejected",
			})
		}
		if err := achievementRefRepo.Review(ctx, refID, req.Status, actorID, req.RejectionNote); err != nil {
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "tidak ditemukan") {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
	case "dosen wali":
		if req.Status == model.AchievementStatusRejected {
			if req.RejectionNote == nil || strings.TrimSpace(*req.RejectionNote) == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "rejection_note wajib diisi jika status rejected",
				})
			}
		}
		if req.Status != model.AchievementStatusVerified &&
			req.Status != model.AchievementStatusRejected {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Status harus verified/rejected",
			})
		}
		ref, err := achievementRefRepo.GetByID(ctx, refID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		if ref.Status != model.AchievementStatusSubmitted {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Hanya boleh memproses status submitted",
			})
		}
		lect, err := achievementLecturerRepo.GetLecturerByUserID(userIDStr)
		if err != nil || lect == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Dosen wali tidak ditemukan",
			})
		}
		st, err := achievementStudentRepo.GetStudentByID(ref.StudentID.String())
		if err != nil || st == nil || st.AdvisorID == nil || st.AdvisorID.String() != lect.ID.String() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Tidak berhak memproses mahasiswa ini",
			})
		}
		if err := achievementRefRepo.Review(ctx, refID, req.Status, actorID, req.RejectionNote); err != nil {
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "tidak ditemukan") {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
	default:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Role tidak diperbolehkan untuk aksi ini",
		})
	}

	return c.JSON(model.SuccessResponse{
		Success: true,
		Message: "Status achievement berhasil diupdate",
	})
}

// SoftDeleteAchievementService godoc
// @Summary Mahasiswa menghapus (soft delete) draft achievement reference (draft -> deleted)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement reference ID (UUID)"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/achievements/{id}/soft-delete [put]
// @Security BearerAuth
func SoftDeleteAchievementService(c *fiber.Ctx) error {
	refID := strings.TrimSpace(c.Params("id"))
	if refID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID reference harus diisi",
		})
	}

	userIDVal := c.Locals("user_id")
	userIDStr, ok := userIDVal.(string)
	if userIDVal == nil || !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	studentUUID, ok := c.Locals("student_uuid").(uuid.UUID)
	if !ok {
		st, err := achievementStudentRepo.GetStudentByUserID(userIDStr)
		if err != nil || st == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "mahasiswa tidak memiliki student_id",
			})
		}
		studentUUID = st.ID
	}
	if err := achievementRefRepo.DeleteByStudent(ctx, refID, studentUUID); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "tidak ditemukan") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Success: true,
		Message: "Status achievement berubah ke deleted (soft delete)",
	})
}

// HardDeleteAchievementService godoc
// @Summary Hard delete achievement (hapus permanen Mongo + reference) untuk status deleted
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement reference ID (UUID)"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/achievements/{id}/delete [delete]
// @Security BearerAuth
func HardDeleteAchievementService(c *fiber.Ctx) error {
	refID := strings.TrimSpace(c.Params("id"))
	if refID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID reference harus diisi",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ref, err := achievementRefRepo.GetByID(ctx, refID)
	if err != nil || ref == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "achievement reference tidak ditemukan",
		})
	}
	if ref.Status != model.AchievementStatusDeleted {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Hard delete hanya boleh untuk status deleted",
		})
	}

	if err := achievementMongoRepo.Delete(ctx, ref.MongoAchievementID); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "tidak ditemukan") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	if err := achievementRefRepo.HardDelete(ctx, refID); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "tidak ditemukan") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Success: true,
		Message: "Achievement dihapus permanen",
	})
}

// GetAchievementsService godoc
// @Summary Daftar semua achievements (Mongo)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default 1)"
// @Param limit query int false "Jumlah per halaman (default 10)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/achievements [get]
// @Security BearerAuth
func GetAchievementsService(c *fiber.Ctx) error {
	page := int64(c.QueryInt("page", 1))
	limit := int64(c.QueryInt("limit", 10))

	roleName, err := resolveRoleName(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	statuses, studentFilter, advisorFilter, err := allowedStatusesByRole(c, roleName, true)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	refs, total, err := achievementRefRepo.ListByStatuses(ctx, statuses, studentFilter, advisorFilter, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil achievement references",
			"error":   err.Error(),
		})
	}

	var ids []string
	for _, r := range refs {
		ids = append(ids, r.MongoAchievementID)
	}
	achievements, err := achievementMongoRepo.GetByIDs(ctx, ids)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data achievements",
			"error":   err.Error(),
		})
	}

	achMap := make(map[string]model.Achievement)
	for _, a := range achievements {
		achMap[a.ID.Hex()] = a
	}

	var combined []model.AchievementWithReference
	for _, r := range refs {
		if a, ok := achMap[r.MongoAchievementID]; ok {
			combined = append(combined, model.AchievementWithReference{
				Achievement: a,
				Reference:   r,
			})
		} else {
			combined = append(combined, model.AchievementWithReference{
				Reference: r,
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data achievements berhasil diambil",
		"data":    combined,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetAchievementReferencesService godoc
// @Summary Daftar semua achievement references (Postgres)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default 1)"
// @Param limit query int false "Jumlah per halaman (default 10)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/achievement-references [get]
// @Security BearerAuth
func GetAchievementReferencesService(c *fiber.Ctx) error {
	page := int64(c.QueryInt("page", 1))
	limit := int64(c.QueryInt("limit", 10))

	roleName, err := resolveRoleName(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	statuses, studentFilter, advisorFilter, err := allowedStatusesByRole(c, roleName, false)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, total, err := achievementRefRepo.ListByStatuses(ctx, statuses, studentFilter, advisorFilter, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil achievement references",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data achievement references berhasil diambil",
		"data":    data,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}
