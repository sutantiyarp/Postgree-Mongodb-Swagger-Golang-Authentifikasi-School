package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hello-fiber/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type mockLecturerRepo struct {
	GetAllLecturersFn  func(page, limit int64) ([]model.Lecturer, int64, error)
	GetLecturerByIDFn  func(id string) (*model.Lecturer, error)
	CreateLecturerFn   func(req model.CreateLecturerRequest) (string, error)
	UpdateLecturerFn   func(id string, req model.UpdateLecturerRequest) error
	DeleteLecturerFn   func(id string) error
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
	id1 := uuid.New()
	id2 := uuid.New()

	lecturerRepo = &mockLecturerRepo{
		GetAllLecturersFn: func(page, limit int64) ([]model.Lecturer, int64, error) {
			if page != 2 || limit != 5 {
				t.Fatalf("expected page=2 limit=5, got page=%d limit=%d", page, limit)
			}
			return []model.Lecturer{
				{ID: id1, LecturerID: "D001", Department: "Informatics"},
				{ID: id2, LecturerID: "D002", Department: "Math"},
			}, 2, nil
		},
	}

	app := fiber.New()
	app.Get("/lecturer", GetAllLecturersService)

	req := httptest.NewRequest(http.MethodGet, "/lecturer?page=2&limit=5", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Data lecturer berhasil diambil" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["total"] != float64(2) {
		t.Fatalf("unexpected total: %#v", body["total"])
	}
}

func TestGetAllLecturersService_RepoError(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{
		GetAllLecturersFn: func(page, limit int64) ([]model.Lecturer, int64, error) {
			return nil, 0, errors.New("db down")
		},
	}

	app := fiber.New()
	app.Get("/lecturer", GetAllLecturersService)

	req := httptest.NewRequest(http.MethodGet, "/lecturer", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Gagal mengambil data lecturer" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetLecturerByIDService_EmptyID(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{}

	app := fiber.New()
	app.Get("/lecturer/:id", GetLecturerByIDService)

	req := httptest.NewRequest(http.MethodGet, "/lecturer/%20%20", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Lecturer ID harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetLecturerByIDService_InvalidUUID(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{}

	app := fiber.New()
	app.Get("/lecturer/:id", GetLecturerByIDService)

	req := httptest.NewRequest(http.MethodGet, "/lecturer/not-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Format Lecturer ID tidak valid" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetLecturerByIDService_NotFound(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{
		GetLecturerByIDFn: func(id string) (*model.Lecturer, error) {
			return nil, errors.New("lecturer tidak ditemukan")
		},
	}

	app := fiber.New()
	app.Get("/lecturer/:id", GetLecturerByIDService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/lecturer/"+id, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Lecturer tidak ditemukan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetLecturerByIDService_Success(t *testing.T) {
	lecID := uuid.New()

	lecturerRepo = &mockLecturerRepo{
		GetLecturerByIDFn: func(id string) (*model.Lecturer, error) {
			// optional: pastiin service ngirim id yang bener
			if id != lecID.String() {
				t.Fatalf("expected id=%s, got %s", lecID.String(), id)
			}
			return &model.Lecturer{
				ID:         lecID,
				UserID:     uuid.New(),
				LecturerID: "D001",
				Department: "Informatics",
			}, nil
		},
	}

	app := fiber.New()
	app.Get("/lecturer/:id", GetLecturerByIDService)

	req := httptest.NewRequest(http.MethodGet, "/lecturer/"+lecID.String(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Data lecturer berhasil diambil" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreateLecturerService_InvalidBody(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{}

	app := fiber.New()
	app.Post("/lecturer", CreateLecturerService)

	req := httptest.NewRequest(http.MethodPost, "/lecturer", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Request body tidak valid" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreateLecturerService_MissingFields(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{}

	app := fiber.New()
	app.Post("/lecturer", CreateLecturerService)

	req := httptest.NewRequest(http.MethodPost, "/lecturer", toJSONReaderLecturer(t, map[string]any{
		"user_id":     uuid.Nil.String(),
		"lecturer_id": "",
		"department":  "",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "user_id, lecturer_id, dan department harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreateLecturerService_Success(t *testing.T) {
	u := uuid.New()

	lecturerRepo = &mockLecturerRepo{
		CreateLecturerFn: func(req model.CreateLecturerRequest) (string, error) {
			if req.UserID != u {
				t.Fatalf("unexpected user_id: %v", req.UserID)
			}
			if req.LecturerID != "D001" || req.Department != "Informatics" {
				t.Fatalf("unexpected payload: %+v", req)
			}
			return "new-lecturer-id", nil
		},
	}

	app := fiber.New()
	app.Post("/lecturer", CreateLecturerService)

	req := httptest.NewRequest(http.MethodPost, "/lecturer", toJSONReaderLecturer(t, map[string]any{
		"user_id":     u.String(),
		"lecturer_id": "D001",
		"department":  "Informatics",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Lecturer berhasil dibuat" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["id"] != "new-lecturer-id" {
		t.Fatalf("unexpected id: %#v", body["id"])
	}
}

func TestCreateLecturerService_DuplicateLecturerID(t *testing.T) {
	u := uuid.New()
	lecturerRepo = &mockLecturerRepo{
		CreateLecturerFn: func(req model.CreateLecturerRequest) (string, error) {
			return "", errors.New("lecturer_id sudah digunakan")
		},
	}

	app := fiber.New()
	app.Post("/lecturer", CreateLecturerService)

	req := httptest.NewRequest(http.MethodPost, "/lecturer", toJSONReaderLecturer(t, map[string]any{
		"user_id":     u.String(),
		"lecturer_id": "D001",
		"department":  "Informatics",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "lecturer_id sudah digunakan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateLecturerService_NoFields(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{}

	app := fiber.New()
	app.Put("/lecturer/:id", UpdateLecturerService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/lecturer/"+id, toJSONReaderLecturer(t, map[string]any{}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Minimal satu field harus diisi untuk update" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateLecturerService_Success(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{
		UpdateLecturerFn: func(id string, req model.UpdateLecturerRequest) error {
			if req.Department == nil || *req.Department != "Updated Dept" {
				t.Fatalf("unexpected req: %+v", req)
			}
			return nil
		},
	}

	app := fiber.New()
	app.Put("/lecturer/:id", UpdateLecturerService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/lecturer/"+id, toJSONReaderLecturer(t, map[string]any{
		"department": "Updated Dept",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Lecturer berhasil diupdate" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateLecturerService_NotFound(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{
		UpdateLecturerFn: func(id string, req model.UpdateLecturerRequest) error {
			return errors.New("lecturer tidak ditemukan")
		},
	}

	app := fiber.New()
	app.Put("/lecturer/:id", UpdateLecturerService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/lecturer/"+id, toJSONReaderLecturer(t, map[string]any{
		"department": "X",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Lecturer tidak ditemukan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestDeleteLecturerService_Success(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{
		DeleteLecturerFn: func(id string) error { return nil },
	}

	app := fiber.New()
	app.Delete("/lecturer/:id", DeleteLecturerService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/lecturer/"+id, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Lecturer berhasil dihapus" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestDeleteLecturerService_NotFound(t *testing.T) {
	lecturerRepo = &mockLecturerRepo{
		DeleteLecturerFn: func(id string) error { return errors.New("lecturer tidak ditemukan") },
	}

	app := fiber.New()
	app.Delete("/lecturer/:id", DeleteLecturerService)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/lecturer/"+id, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	body := decodeMapLecturer(t, resp)
	if body["message"] != "Lecturer tidak ditemukan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}
