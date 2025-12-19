package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"hello-fiber/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockAchievementMongoRepo struct {
	CreateFn   func(ctx context.Context, studentID uuid.UUID, req model.CreateAchievementRequest) (string, error)
	GetByIDsFn func(ctx context.Context, ids []string) ([]model.Achievement, error)
	ListFn     func(ctx context.Context, page, limit int64) ([]model.Achievement, int64, error)
	DeleteFn   func(ctx context.Context, id string) error
}

func (m *mockAchievementMongoRepo) Create(ctx context.Context, studentID uuid.UUID, req model.CreateAchievementRequest) (string, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, studentID, req)
	}
	return "", nil
}

func (m *mockAchievementMongoRepo) GetByIDs(ctx context.Context, ids []string) ([]model.Achievement, error) {
	if m.GetByIDsFn != nil {
		return m.GetByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *mockAchievementMongoRepo) List(ctx context.Context, page, limit int64) ([]model.Achievement, int64, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, page, limit)
	}
	return nil, 0, nil
}

func (m *mockAchievementMongoRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

type mockAchievementRefRepo struct {
	CreateDraftFn     func(ctx context.Context, studentID uuid.UUID, mongoID string) (string, error)
	SubmitDraftFn     func(ctx context.Context, refID string, studentID uuid.UUID) error
	ReviewFn          func(ctx context.Context, refID string, status string, adminID uuid.UUID, note *string) error
	DeleteFn          func(ctx context.Context, refID string, adminID uuid.UUID) error
	DeleteByStudentFn func(ctx context.Context, refID string, studentID uuid.UUID) error
	HardDeleteFn      func(ctx context.Context, refID string) error
	GetByIDFn         func(ctx context.Context, id string) (*model.AchievementReference, error)
	ListFn            func(ctx context.Context, page, limit int64) ([]model.AchievementReference, int64, error)
	ListByStatusesFn  func(ctx context.Context, statuses []string, studentID *uuid.UUID, advisorID *uuid.UUID, page, limit int64) ([]model.AchievementReference, int64, error)
}

func (m *mockAchievementRefRepo) CreateDraft(ctx context.Context, studentID uuid.UUID, mongoID string) (string, error) {
	if m.CreateDraftFn != nil {
		return m.CreateDraftFn(ctx, studentID, mongoID)
	}
	return "", nil
}

func (m *mockAchievementRefRepo) SubmitDraft(ctx context.Context, refID string, studentID uuid.UUID) error {
	if m.SubmitDraftFn != nil {
		return m.SubmitDraftFn(ctx, refID, studentID)
	}
	return nil
}

func (m *mockAchievementRefRepo) Review(ctx context.Context, refID string, status string, adminID uuid.UUID, note *string) error {
	if m.ReviewFn != nil {
		return m.ReviewFn(ctx, refID, status, adminID, note)
	}
	return nil
}

func (m *mockAchievementRefRepo) Delete(ctx context.Context, refID string, adminID uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, refID, adminID)
	}
	return nil
}

func (m *mockAchievementRefRepo) DeleteByStudent(ctx context.Context, refID string, studentID uuid.UUID) error {
	if m.DeleteByStudentFn != nil {
		return m.DeleteByStudentFn(ctx, refID, studentID)
	}
	return nil
}

func (m *mockAchievementRefRepo) HardDelete(ctx context.Context, refID string) error {
	if m.HardDeleteFn != nil {
		return m.HardDeleteFn(ctx, refID)
	}
	return nil
}

func (m *mockAchievementRefRepo) GetByID(ctx context.Context, id string) (*model.AchievementReference, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockAchievementRefRepo) List(ctx context.Context, page, limit int64) ([]model.AchievementReference, int64, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, page, limit)
	}
	return nil, 0, nil
}

func (m *mockAchievementRefRepo) ListByStatuses(ctx context.Context, statuses []string, studentID *uuid.UUID, advisorID *uuid.UUID, page, limit int64) ([]model.AchievementReference, int64, error) {
	if m.ListByStatusesFn != nil {
		return m.ListByStatusesFn(ctx, statuses, studentID, advisorID, page, limit)
	}
	return nil, 0, nil
}

type mockStudentRepo struct {
	GetAllStudentsFn     func(page, limit int64) ([]model.Student, int64, error)
	GetStudentByIDFn     func(id string) (*model.Student, error)
	GetStudentByUserIDFn func(userID string) (*model.Student, error)
	CreateStudentFn      func(req model.CreateStudentRequest) (string, error)
	UpdateStudentFn      func(id string, req model.UpdateStudentRequest) error
	DeleteStudentFn      func(id string) error
}

func (m *mockStudentRepo) GetAllStudents(page, limit int64) ([]model.Student, int64, error) {
	if m.GetAllStudentsFn != nil {
		return m.GetAllStudentsFn(page, limit)
	}
	return nil, 0, nil
}

func (m *mockStudentRepo) GetStudentByID(id string) (*model.Student, error) {
	if m.GetStudentByIDFn != nil {
		return m.GetStudentByIDFn(id)
	}
	return nil, nil
}

func (m *mockStudentRepo) GetStudentByUserID(userID string) (*model.Student, error) {
	if m.GetStudentByUserIDFn != nil {
		return m.GetStudentByUserIDFn(userID)
	}
	return nil, nil
}

func (m *mockStudentRepo) CreateStudent(req model.CreateStudentRequest) (string, error) {
	if m.CreateStudentFn != nil {
		return m.CreateStudentFn(req)
	}
	return "", nil
}

func (m *mockStudentRepo) UpdateStudent(id string, req model.UpdateStudentRequest) error {
	if m.UpdateStudentFn != nil {
		return m.UpdateStudentFn(id, req)
	}
	return nil
}

func (m *mockStudentRepo) DeleteStudent(id string) error {
	if m.DeleteStudentFn != nil {
		return m.DeleteStudentFn(id)
	}
	return nil
}

type mockLectRepo struct {
	GetAllLecturersFn     func(page, limit int64) ([]model.Lecturer, int64, error)
	GetLecturerByIDFn     func(id string) (*model.Lecturer, error)
	GetLecturerByUserIDFn func(userID string) (*model.Lecturer, error)
	CreateLecturerFn      func(req model.CreateLecturerRequest) (string, error)
	UpdateLecturerFn      func(id string, req model.UpdateLecturerRequest) error
	DeleteLecturerFn      func(id string) error
}

func (m *mockLectRepo) GetAllLecturers(page, limit int64) ([]model.Lecturer, int64, error) {
	if m.GetAllLecturersFn != nil {
		return m.GetAllLecturersFn(page, limit)
	}
	return nil, 0, nil
}

func (m *mockLectRepo) GetLecturerByID(id string) (*model.Lecturer, error) {
	if m.GetLecturerByIDFn != nil {
		return m.GetLecturerByIDFn(id)
	}
	return nil, nil
}

func (m *mockLectRepo) GetLecturerByUserID(userID string) (*model.Lecturer, error) {
	if m.GetLecturerByUserIDFn != nil {
		return m.GetLecturerByUserIDFn(userID)
	}
	return nil, nil
}

func (m *mockLectRepo) CreateLecturer(req model.CreateLecturerRequest) (string, error) {
	if m.CreateLecturerFn != nil {
		return m.CreateLecturerFn(req)
	}
	return "", nil
}

func (m *mockLectRepo) UpdateLecturer(id string, req model.UpdateLecturerRequest) error {
	if m.UpdateLecturerFn != nil {
		return m.UpdateLecturerFn(id, req)
	}
	return nil
}

func (m *mockLectRepo) DeleteLecturer(id string) error {
	if m.DeleteLecturerFn != nil {
		return m.DeleteLecturerFn(id)
	}
	return nil
}

func toJSONReaderAchievement(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewReader(b)
}

func decodeMapAchievement(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return out
}

func TestCreateAchievementService_MultipartPDF_Success(t *testing.T) {
	os.RemoveAll("uploads")
	defer os.RemoveAll("uploads")

	studentID := uuid.New()
	achievementMongoRepo = &mockAchievementMongoRepo{
		CreateFn: func(ctx context.Context, sID uuid.UUID, req model.CreateAchievementRequest) (string, error) {
			if len(req.Attachments) != 1 {
				t.Fatalf("expected 1 attachment, got %d", len(req.Attachments))
			}
			att := req.Attachments[0]
			if att.FileName != "file.pdf" {
				t.Fatalf("unexpected filename: %s", att.FileName)
			}
			if !strings.HasPrefix(att.FileURL, "/uploads/") {
				t.Fatalf("unexpected file url: %s", att.FileURL)
			}
			return "mongo123", nil
		},
	}
	achievementRefRepo = &mockAchievementRefRepo{
		CreateDraftFn: func(ctx context.Context, sID uuid.UUID, mongoID string) (string, error) {
			return "ref123", nil
		},
	}

	app := fiber.New()
	app.Post("/achievements", func(c *fiber.Ctx) error {
		c.Locals("student_uuid", studentID)
		return CreateAchievementService(c)
	})

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("achievement_type", "academic")
	_ = w.WriteField("title", "Hasil Turnitin")
	_ = w.WriteField("description", "Cek turnitin")
	_ = w.WriteField("details", `{"score":8}`)
	fw, _ := w.CreateFormFile("attachments", "file.pdf")
	fw.Write([]byte("dummy"))
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/achievements", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestCreateAchievementService_MultipartRejectNonPDF(t *testing.T) {
	os.RemoveAll("uploads")
	defer os.RemoveAll("uploads")

	studentID := uuid.New()
	achievementMongoRepo = &mockAchievementMongoRepo{}
	achievementRefRepo = &mockAchievementRefRepo{}

	app := fiber.New()
	app.Post("/achievements", func(c *fiber.Ctx) error {
		c.Locals("student_uuid", studentID)
		return CreateAchievementService(c)
	})

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("achievement_type", "academic")
	_ = w.WriteField("title", "Hasil Turnitin")
	_ = w.WriteField("description", "Cek turnitin")
	_ = w.WriteField("details", `{"score":8}`)
	fw, _ := w.CreateFormFile("attachments", "note.txt")
	fw.Write([]byte("dummy"))
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/achievements", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}
func TestCreateAchievementService_Success(t *testing.T) {
	studentID := uuid.New()
	achievementMongoRepo = &mockAchievementMongoRepo{
		CreateFn: func(ctx context.Context, sID uuid.UUID, req model.CreateAchievementRequest) (string, error) {
			if sID != studentID {
				t.Fatalf("studentID mismatch: %v", sID)
			}
			if rank, ok := req.Details["rank"].(int); !ok || rank != 1 {
				t.Fatalf("rank not normalized to int: %#v", req.Details["rank"])
			}
			return "mongo123", nil
		},
	}
	achievementRefRepo = &mockAchievementRefRepo{
		CreateDraftFn: func(ctx context.Context, sID uuid.UUID, mongoID string) (string, error) {
			if sID != studentID {
				t.Fatalf("studentID mismatch: %v", sID)
			}
			if mongoID != "mongo123" {
				t.Fatalf("unexpected mongoID: %s", mongoID)
			}
			return "ref123", nil
		},
	}

	app := fiber.New()
	app.Post("/achievements", func(c *fiber.Ctx) error {
		c.Locals("student_uuid", studentID)
		return CreateAchievementService(c)
	})

	payload := map[string]any{
		"achievement_type": "competition",
		"title":            "Juara 1",
		"description":      "Menang lomba",
		"details": map[string]any{
			"rank":             1.0,
			"competitionLevel": "National",
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/achievements", toJSONReaderAchievement(t, payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	body := decodeMapAchievement(t, resp)
	if body["message"] != "Achievement berhasil dibuat" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestCreateAchievementService_NoStudent(t *testing.T) {
	app := fiber.New()
	app.Post("/achievements", func(c *fiber.Ctx) error {
		return CreateAchievementService(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/achievements", bytes.NewBufferString("{}"))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
	body := decodeMapAchievement(t, resp)
	if body["message"] != "Hanya mahasiswa yang dapat mengakses" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestSubmitAchievementService_Success(t *testing.T) {
	studentID := uuid.New()
	called := false
	achievementRefRepo = &mockAchievementRefRepo{
		SubmitDraftFn: func(ctx context.Context, refID string, sID uuid.UUID) error {
			called = true
			if refID != "ref-1" {
				t.Fatalf("unexpected refID: %s", refID)
			}
			if sID != studentID {
				t.Fatalf("studentID mismatch: %v", sID)
			}
			return nil
		},
	}

	app := fiber.New()
	app.Put("/achievements/:id/submit", func(c *fiber.Ctx) error {
		c.Locals("student_uuid", studentID)
		return SubmitAchievementService(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/achievements/ref-1/submit", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !called {
		t.Fatalf("SubmitDraft was not called")
	}
}

func TestReviewAchievementService_AdminVerified(t *testing.T) {
	userID := uuid.New()
	roleID := "role-admin"
	achievementRoleRepo = &mockRoleRepo{
		GetRoleByIDFn: func(id string) (*model.Role, error) {
			if id != roleID {
				t.Fatalf("unexpected roleID: %s", id)
			}
			return &model.Role{ID: id, Name: "Admin"}, nil
		},
	}
	called := false
	achievementRefRepo = &mockAchievementRefRepo{
		ReviewFn: func(ctx context.Context, refID string, status string, adminID uuid.UUID, note *string) error {
			called = true
			if status != model.AchievementStatusVerified {
				t.Fatalf("unexpected status: %s", status)
			}
			if refID != "ref-2" {
				t.Fatalf("unexpected refID: %s", refID)
			}
			if adminID.String() != userID.String() {
				t.Fatalf("unexpected adminID: %s", adminID)
			}
			return nil
		},
	}

	app := fiber.New()
	app.Put("/achievements/:id/review", func(c *fiber.Ctx) error {
		c.Locals("role_id", roleID)
		c.Locals("user_id", userID.String())
		return ReviewAchievementService(c)
	})

	payload := map[string]any{"status": "verified"}
	req := httptest.NewRequest(http.MethodPut, "/achievements/ref-2/review", toJSONReaderAchievement(t, payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !called {
		t.Fatalf("Review was not called")
	}
}

func TestSoftDeleteAchievementService_NotFound(t *testing.T) {
	studentID := uuid.New()
	userID := uuid.New().String()
	achievementStudentRepo = &mockStudentRepo{}
	achievementRefRepo = &mockAchievementRefRepo{
		DeleteByStudentFn: func(ctx context.Context, refID string, sID uuid.UUID) error {
			return errors.New("tidak ditemukan")
		},
	}

	app := fiber.New()
	app.Put("/achievements/:id/soft-delete", func(c *fiber.Ctx) error {
		c.Locals("student_uuid", studentID)
		c.Locals("user_id", userID)
		return SoftDeleteAchievementService(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/achievements/ref-404/soft-delete", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHardDeleteAchievementService_WrongStatus(t *testing.T) {
	refID := "ref-1"
	achievementRefRepo = &mockAchievementRefRepo{
		GetByIDFn: func(ctx context.Context, id string) (*model.AchievementReference, error) {
			return &model.AchievementReference{
				ID:                 uuid.New(),
				MongoAchievementID: "mongo-x",
				Status:             model.AchievementStatusSubmitted,
			}, nil
		},
	}

	app := fiber.New()
	app.Delete("/achievements/:id/delete", HardDeleteAchievementService)

	req := httptest.NewRequest(http.MethodDelete, "/achievements/"+refID+"/delete", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	body := decodeMapAchievement(t, resp)
	if body["message"] != "Hard delete hanya boleh untuk status deleted" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestGetAchievementsService_AdminSuccess(t *testing.T) {
	roleID := "role-admin"
	achievementRoleRepo = &mockRoleRepo{
		GetRoleByIDFn: func(id string) (*model.Role, error) {
			return &model.Role{ID: id, Name: "Admin"}, nil
		},
	}
	refID := uuid.New()
	achievementRefRepo = &mockAchievementRefRepo{
		ListByStatusesFn: func(ctx context.Context, statuses []string, studentID *uuid.UUID, advisorID *uuid.UUID, page, limit int64) ([]model.AchievementReference, int64, error) {
			if len(statuses) == 0 {
				t.Fatalf("statuses empty")
			}
			return []model.AchievementReference{{
				ID:                 refID,
				MongoAchievementID: "mongo-1",
				Status:             model.AchievementStatusSubmitted,
				StudentID:          uuid.New(),
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}}, 1, nil
		},
	}
	achievementMongoRepo = &mockAchievementMongoRepo{
		GetByIDsFn: func(ctx context.Context, ids []string) ([]model.Achievement, error) {
			return []model.Achievement{{
				ID:              bson.NewObjectID(),
				AchievementType: "competition",
				Title:           "Juara",
				Description:     "Desc",
			}}, nil
		},
	}

	app := fiber.New()
	app.Get("/achievements", func(c *fiber.Ctx) error {
		c.Locals("role_id", roleID)
		return GetAchievementsService(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/achievements", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
}
