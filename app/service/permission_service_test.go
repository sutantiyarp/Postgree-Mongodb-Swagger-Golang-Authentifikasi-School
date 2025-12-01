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
)

type mockPermissionRepo struct {
	GetAllPermissionsFn func(page, limit int64) ([]model.Permission, int64, error)
	GetPermissionByIDFn func(id string) (*model.Permission, error)
	CreatePermissionFn  func(req model.CreatePermissionRequest) (string, error)
	UpdatePermissionFn  func(id string, req model.UpdatePermissionRequest) error
	DeletePermissionFn  func(id string) error
}

func (m *mockPermissionRepo) GetAllPermissions(page, limit int64) ([]model.Permission, int64, error) {
	if m.GetAllPermissionsFn != nil {
		return m.GetAllPermissionsFn(page, limit)
	}
	return nil, 0, nil
}

func (m *mockPermissionRepo) GetPermissionByID(id string) (*model.Permission, error) {
	if m.GetPermissionByIDFn != nil {
		return m.GetPermissionByIDFn(id)
	}
	return nil, nil
}

func (m *mockPermissionRepo) CreatePermission(req model.CreatePermissionRequest) (string, error) {
	if m.CreatePermissionFn != nil {
		return m.CreatePermissionFn(req)
	}
	return "", nil
}

func (m *mockPermissionRepo) UpdatePermission(id string, req model.UpdatePermissionRequest) error {
	if m.UpdatePermissionFn != nil {
		return m.UpdatePermissionFn(id, req)
	}
	return nil
}

func (m *mockPermissionRepo) DeletePermission(id string) error {
	if m.DeletePermissionFn != nil {
		return m.DeletePermissionFn(id)
	}
	return nil
}

func toJSONReaderPermission(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewReader(b)
}

func decodeMapPermission(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return out
}

func TestGetAllPermissionsService_Success(t *testing.T) {
	permissionRepo = &mockPermissionRepo{
		GetAllPermissionsFn: func(page, limit int64) ([]model.Permission, int64, error) {
			if page != 2 || limit != 5 {
				t.Fatalf("expected page=2 limit=5, got page=%d limit=%d", page, limit)
			}
			return []model.Permission{
				{ID: "p1", Name: "achievement:create", Resource: "achievement", Action: "create", Description: "Create achievement"},
				{ID: "p2", Name: "achievement:read", Resource: "achievement", Action: "read", Description: "Read achievement"},
			}, 2, nil
		},
	}

	app := fiber.New()
	app.Get("/permissions", GetAllPermissionsService)

	req := httptest.NewRequest(http.MethodGet, "/permissions?page=2&limit=5", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeMapPermission(t, resp)
	if body["message"] != "Data permission berhasil diambil" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["total"] != float64(2) {
		t.Fatalf("unexpected total: %#v", body["total"])
	}
}

func TestGetAllPermissionsService_DefaultPagination(t *testing.T) {
	permissionRepo = &mockPermissionRepo{
		GetAllPermissionsFn: func(page, limit int64) ([]model.Permission, int64, error) {
			if page != 1 || limit != 10 {
				t.Fatalf("expected page=1 limit=10, got page=%d limit=%d", page, limit)
			}
			return []model.Permission{}, 0, nil
		},
	}

	app := fiber.New()
	app.Get("/permissions", GetAllPermissionsService)

	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetAllPermissionsService_RepoError(t *testing.T) {
	permissionRepo = &mockPermissionRepo{
		GetAllPermissionsFn: func(page, limit int64) ([]model.Permission, int64, error) {
			return nil, 0, errors.New("db down")
		},
	}

	app := fiber.New()
	app.Get("/permissions", GetAllPermissionsService)

	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	body := decodeMapPermission(t, resp)
	if body["message"] != "Gagal mengambil data permission" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetPermissionByIDService_EmptyID(t *testing.T) {
	permissionRepo = &mockPermissionRepo{}

	app := fiber.New()
	app.Get("/permissions/:id", GetPermissionByIDService)

	req := httptest.NewRequest(http.MethodGet, "/permissions/%20%20", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeMapPermission(t, resp)
	if body["message"] != "Permission ID harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetPermissionByIDService_NotFound(t *testing.T) {
	permissionRepo = &mockPermissionRepo{
		GetPermissionByIDFn: func(id string) (*model.Permission, error) {
			return nil, errors.New("permission tidak ditemukan")
		},
	}

	app := fiber.New()
	app.Get("/permissions/:id", GetPermissionByIDService)

	req := httptest.NewRequest(http.MethodGet, "/permissions/p404", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeMapPermission(t, resp)
	if body["message"] != "Permission tidak ditemukan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetPermissionByIDService_RepoError(t *testing.T) {
	permissionRepo = &mockPermissionRepo{
		GetPermissionByIDFn: func(id string) (*model.Permission, error) {
			return nil, errors.New("some db error")
		},
	}

	app := fiber.New()
	app.Get("/permissions/:id", GetPermissionByIDService)

	req := httptest.NewRequest(http.MethodGet, "/permissions/p1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	body := decodeMapPermission(t, resp)
	if body["message"] != "Gagal mengambil data permission" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreatePermissionService_InvalidBody(t *testing.T) {
	permissionRepo = &mockPermissionRepo{}

	app := fiber.New()
	app.Post("/permissions", CreatePermissionService)

	req := httptest.NewRequest(http.MethodPost, "/permissions", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeMapPermission(t, resp)
	if body["message"] != "Request body tidak valid" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreatePermissionService_MissingFields(t *testing.T) {
	permissionRepo = &mockPermissionRepo{}

	app := fiber.New()
	app.Post("/permissions", CreatePermissionService)

	req := httptest.NewRequest(http.MethodPost, "/permissions", toJSONReaderPermission(t, map[string]any{
		"name": "x",
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

	body := decodeMapPermission(t, resp)
	if body["message"] != "Name, resource, dan action harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreatePermissionService_Success(t *testing.T) {
	permissionRepo = &mockPermissionRepo{
		CreatePermissionFn: func(req model.CreatePermissionRequest) (string, error) {
			if req.Name != "achievement:create" || req.Resource != "achievement" || req.Action != "create" {
				t.Fatalf("unexpected req: %+v", req)
			}
			return "new-id", nil
		},
	}

	app := fiber.New()
	app.Post("/permissions", CreatePermissionService)

	req := httptest.NewRequest(http.MethodPost, "/permissions", toJSONReaderPermission(t, map[string]any{
		"name":        "achievement:create",
		"resource":    "achievement",
		"action":      "create",
		"description": "Create achievement",
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

	body := decodeMapPermission(t, resp)
	if body["message"] != "Permission berhasil dibuat" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["id"] != "new-id" {
		t.Fatalf("unexpected id: %#v", body["id"])
	}
}

func TestUpdatePermissionService_NoFields(t *testing.T) {
	permissionRepo = &mockPermissionRepo{}

	app := fiber.New()
	app.Put("/permissions/:id", UpdatePermissionService)

	req := httptest.NewRequest(http.MethodPut, "/permissions/p1", toJSONReaderPermission(t, map[string]any{}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeMapPermission(t, resp)
	if body["message"] != "Minimal satu field harus diisi untuk update" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdatePermissionService_Success(t *testing.T) {
	permissionRepo = &mockPermissionRepo{
		UpdatePermissionFn: func(id string, req model.UpdatePermissionRequest) error {
			if id != "p1" {
				t.Fatalf("expected id=p1, got %s", id)
			}
			if req.Description != "updated" {
				t.Fatalf("unexpected req: %+v", req)
			}
			return nil
		},
	}

	app := fiber.New()
	app.Put("/permissions/:id", UpdatePermissionService)

	req := httptest.NewRequest(http.MethodPut, "/permissions/p1", toJSONReaderPermission(t, map[string]any{
		"description": "updated",
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

	body := decodeMapPermission(t, resp)
	if body["message"] != "Permission berhasil diupdate" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestDeletePermissionService_Success(t *testing.T) {
	permissionRepo = &mockPermissionRepo{
		DeletePermissionFn: func(id string) error {
			if id != "p1" {
				t.Fatalf("expected id=p1, got %s", id)
			}
			return nil
		},
	}

	app := fiber.New()
	app.Delete("/permissions/:id", DeletePermissionService)

	req := httptest.NewRequest(http.MethodDelete, "/permissions/p1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeMapPermission(t, resp)
	if body["message"] != "Permission berhasil dihapus" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}
