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

type mockLecturerRepo struct {
	GetAllLecturersFn     func(page, limit int64) ([]model.Lecturer, int64, error)
	GetLecturerByIDFn     func(id string) (*model.Lecturer, error)
	GetLecturerByUserIDFn func(userID string) (*model.Lecturer, error)
	CreateLecturerFn      func(req model.CreateLecturerRequest) (string, error)
	UpdateLecturerFn      func(id string, req model.UpdateLecturerRequest) error
	DeleteLecturerFn      func(id string) error
}

func (m *mockLecturerRepo) GetAllLecturers(page, limit int64) ([]model.Lecturer, int64, error) {
	if m.GetAllLecturersFn != nil {
		return m.GetAllLecturersFn(page, limit)
	}
	return nil, 0, nil
}

func (m *mockLecturerRepo) GetLecturerByID(id string) (*model.Lecturer, error) {
	if m.GetLecturerByIDFn != nil {
		return m.GetLecturerByIDFn(id)
	}
	return nil, nil
}

func (m *mockLecturerRepo) GetLecturerByUserID(userID string) (*model.Lecturer, error) {
	if m.GetLecturerByUserIDFn != nil {
		return m.GetLecturerByUserIDFn(userID)
	}
	return nil, nil
}

func (m *mockLecturerRepo) CreateLecturer(req model.CreateLecturerRequest) (string, error) {
	if m.CreateLecturerFn != nil {
		return m.CreateLecturerFn(req)
	}
	return "", nil
}

func (m *mockLecturerRepo) UpdateLecturer(id string, req model.UpdateLecturerRequest) error {
	if m.UpdateLecturerFn != nil {
		return m.UpdateLecturerFn(id, req)
	}
	return nil
}

func (m *mockLecturerRepo) DeleteLecturer(id string) error {
	if m.DeleteLecturerFn != nil {
		return m.DeleteLecturerFn(id)
	}
	return nil
}

func toJSONReaderLecturer(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewReader(b)
}

func decodeMapLecturer(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return out
}

func TestGetAllLecturersService_Success(t *testing.T) {
	now := time.Now()
	lecturerRepo = &mockLecturerRepo{
		GetAllLecturersFn: func(page, limit int64) ([]model.Lecturer, int64, error) {
			if page != 2 || limit != 5 {
				t.Fatalf("unexpected pagination page=%d limit=%d", page, limit)
			}
			return []model.Lecturer{{
				ID:         uuid.New(),
				UserID:     uuid.New(),
				LecturerID: "L001",
				Department: "TI",
				CreatedAt:  now,
			}}, 1, nil
		},
	}

	app := fiber.New()
	app.Get("/lecturers", GetAllLecturersService)

	req := httptest.NewRequest(http.MethodGet, "/lecturers?page=2&limit=5", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Data lecturer berhasil diambil" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestGetLecturerByIDService_InvalidUUID(t *testing.T) {
	app := fiber.New()
	app.Get("/lecturers/:id", GetLecturerByIDService)

	req := httptest.NewRequest(http.MethodGet, "/lecturers/not-a-uuid", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Format Lecturer ID tidak valid" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestCreateLecturerService_Validation(t *testing.T) {
	app := fiber.New()
	app.Post("/lecturers", CreateLecturerService)

	req := httptest.NewRequest(http.MethodPost, "/lecturers", toJSONReaderLecturer(t, map[string]any{}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "user_id, lecturer_id, dan department harus diisi" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestCreateLecturerService_Success(t *testing.T) {
	uid := uuid.New()
	lecturerRepo = &mockLecturerRepo{
		CreateLecturerFn: func(req model.CreateLecturerRequest) (string, error) {
			if req.UserID != uid {
				t.Fatalf("unexpected user_id: %v", req.UserID)
			}
			return "lec-123", nil
		},
	}

	app := fiber.New()
	app.Post("/lecturers", CreateLecturerService)

	payload := map[string]any{
		"user_id":     uid.String(),
		"lecturer_id": "L002",
		"department":  "SI",
	}
	req := httptest.NewRequest(http.MethodPost, "/lecturers", toJSONReaderLecturer(t, payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Lecturer berhasil dibuat" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestUpdateLecturerService_NoFields(t *testing.T) {
	app := fiber.New()
	app.Put("/lecturers/:id", UpdateLecturerService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/lecturers/"+id, bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Minimal satu field harus diisi untuk update" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestDeleteLecturerService_NotFound(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{
		DeleteLecturerFn: func(id string) error {
			return errors.New("lecturer tidak ditemukan")
		},
	}

	app := fiber.New()
	app.Delete("/lecturers/:id", DeleteLecturerService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/lecturers/"+id, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Lecturer tidak ditemukan" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}
