package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hello-fiber/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type mockStudentRepoStd struct {
	GetAllStudentsFn     func(page, limit int64) ([]model.Student, int64, error)
	GetStudentByIDFn     func(id string) (*model.Student, error)
	GetStudentByUserIDFn func(userID string) (*model.Student, error)
	CreateStudentFn      func(req model.CreateStudentRequest) (string, error)
	UpdateStudentFn      func(id string, req model.UpdateStudentRequest) error
	DeleteStudentFn      func(id string) error
}

func (m *mockStudentRepoStd) GetAllStudents(page, limit int64) ([]model.Student, int64, error) {
	if m.GetAllStudentsFn != nil {
		return m.GetAllStudentsFn(page, limit)
	}
	return nil, 0, nil
}

func (m *mockStudentRepoStd) GetStudentByID(id string) (*model.Student, error) {
	if m.GetStudentByIDFn != nil {
		return m.GetStudentByIDFn(id)
	}
	return nil, nil
}

func (m *mockStudentRepoStd) GetStudentByUserID(userID string) (*model.Student, error) {
	if m.GetStudentByUserIDFn != nil {
		return m.GetStudentByUserIDFn(userID)
	}
	return nil, nil
}

func (m *mockStudentRepoStd) CreateStudent(req model.CreateStudentRequest) (string, error) {
	if m.CreateStudentFn != nil {
		return m.CreateStudentFn(req)
	}
	return "", nil
}

func (m *mockStudentRepoStd) UpdateStudent(id string, req model.UpdateStudentRequest) error {
	if m.UpdateStudentFn != nil {
		return m.UpdateStudentFn(id, req)
	}
	return nil
}

func (m *mockStudentRepoStd) DeleteStudent(id string) error {
	if m.DeleteStudentFn != nil {
		return m.DeleteStudentFn(id)
	}
	return nil
}

func jsonBodyStudent(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewReader(b)
}

func decodeMapStudent(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return out
}

func TestGetAllStudentsService_Success(t *testing.T) {
	studentRepo = &mockStudentRepoStd{
		GetAllStudentsFn: func(page, limit int64) ([]model.Student, int64, error) {
			if page != 2 || limit != 5 {
				t.Fatalf("unexpected pagination: %d %d", page, limit)
			}
			return []model.Student{{
				ID:           uuid.New(),
				UserID:       uuid.New(),
				StudentID:    "S001",
				ProgramStudy: "TI",
				AcademicYear: "2025",
				CreatedAt:    time.Now(),
			}}, 1, nil
		},
	}

	app := fiber.New()
	app.Get("/students", GetAllStudentsService)

	req := httptest.NewRequest(http.MethodGet, "/students?page=2&limit=5", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	body := decodeMapStudent(t, resp)
	if body["message"] != "Data student berhasil diambil" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestGetStudentByIDService_InvalidUUID(t *testing.T) {
	app := fiber.New()
	app.Get("/students/:id", GetStudentByIDService)

	req := httptest.NewRequest(http.MethodGet, "/students/not-uuid", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCreateStudentService_Validation(t *testing.T) {
	app := fiber.New()
	app.Post("/students", CreateStudentService)

	req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	body := decodeMapStudent(t, resp)
	if body["message"] != "user_id dan student_id harus diisi" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestCreateStudentService_Success(t *testing.T) {
	uid := uuid.New()
	studentRepo = &mockStudentRepoStd{
		CreateStudentFn: func(req model.CreateStudentRequest) (string, error) {
			if req.UserID != uid {
				t.Fatalf("unexpected user_id: %v", req.UserID)
			}
			if req.StudentID != "S123" {
				t.Fatalf("unexpected student_id: %s", req.StudentID)
			}
			return "stud-1", nil
		},
	}

	app := fiber.New()
	app.Post("/students", CreateStudentService)

	payload := map[string]any{
		"user_id":       uid.String(),
		"student_id":    "S123",
		"program_study": "TI",
		"academic_year": "2025",
	}
	req := httptest.NewRequest(http.MethodPost, "/students", jsonBodyStudent(t, payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	body := decodeMapStudent(t, resp)
	if body["message"] != "Student berhasil dibuat" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestUpdateStudentService_NoFields(t *testing.T) {
	app := fiber.New()
	app.Put("/students/:id", UpdateStudentService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/students/"+id, bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	body := decodeMapStudent(t, resp)
	if body["message"] != "Minimal satu field harus diisi untuk update" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestDeleteStudentService_NotFound(t *testing.T) {
	studentRepo = &mockStudentRepoStd{
		DeleteStudentFn: func(id string) error {
			return errors.New("student tidak ditemukan")
		},
	}

	app := fiber.New()
	app.Delete("/students/:id", DeleteStudentService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/students/"+id, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	body := decodeMapStudent(t, resp)
	if body["message"] != "Student tidak ditemukan" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}
