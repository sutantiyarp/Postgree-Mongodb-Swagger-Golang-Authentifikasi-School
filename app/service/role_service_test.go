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
)

type mockRoleRepo struct {
	GetAllRolesFn  func(page, limit int64) ([]model.Role, int64, error)
	GetRoleByIDFn  func(id string) (*model.Role, error)
	GetRoleByNameFn func(name string) (*model.Role, error)

	CreateRoleFn func(req model.CreateRoleRequest) (string, error)
	UpdateRoleFn func(id string, req model.UpdateRoleRequest) error
	DeleteRoleFn func(id string) error
}

func (m *mockRoleRepo) GetAllRoles(page, limit int64) ([]model.Role, int64, error) {
	if m.GetAllRolesFn != nil {
		return m.GetAllRolesFn(page, limit)
	}
	return nil, 0, nil
}

func (m *mockRoleRepo) GetRoleByID(id string) (*model.Role, error) {
	if m.GetRoleByIDFn != nil {
		return m.GetRoleByIDFn(id)
	}
	return nil, nil
}

func (m *mockRoleRepo) GetRoleByName(name string) (*model.Role, error) {
	if m.GetRoleByNameFn != nil {
		return m.GetRoleByNameFn(name)
	}
	return nil, nil
}

func (m *mockRoleRepo) CreateRole(req model.CreateRoleRequest) (string, error) {
	if m.CreateRoleFn != nil {
		return m.CreateRoleFn(req)
	}
	return "role-id-1", nil
}

func (m *mockRoleRepo) UpdateRole(id string, req model.UpdateRoleRequest) error {
	if m.UpdateRoleFn != nil {
		return m.UpdateRoleFn(id, req)
	}
	return nil
}

func (m *mockRoleRepo) DeleteRole(id string) error {
	if m.DeleteRoleFn != nil {
		return m.DeleteRoleFn(id)
	}
	return nil
}

func jsonBodyRole(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewReader(b)
}

func decodeMapRole(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return out
}

func TestGetAllRolesService_Success(t *testing.T) {
	roleRepo = &mockRoleRepo{
		GetAllRolesFn: func(page, limit int64) ([]model.Role, int64, error) {
			if page != 2 || limit != 5 {
				t.Fatalf("expected page=2 limit=5, got page=%d limit=%d", page, limit)
			}
			now := time.Now()
			return []model.Role{
				{ID: "r1", Name: "Admin", Description: "admin", CreatedAt: now},
				{ID: "r2", Name: "User", Description: "user", CreatedAt: now},
			}, 2, nil
		},
	}

	app := fiber.New()
	app.Get("/roles", GetAllRolesService)

	req := httptest.NewRequest(http.MethodGet, "/roles?page=2&limit=5", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMapRole(t, resp)
	if body["success"] != true {
		t.Fatalf("expected success=true, got %#v", body["success"])
	}
	if body["message"] != "Data role berhasil diambil" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["total"].(float64) != 2 {
		t.Fatalf("unexpected total: %#v", body["total"])
	}
}

func TestGetAllRolesService_Error(t *testing.T) {
	roleRepo = &mockRoleRepo{
		GetAllRolesFn: func(page, limit int64) ([]model.Role, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}

	app := fiber.New()
	app.Get("/roles", GetAllRolesService)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	body := decodeMapRole(t, resp)
	if body["message"] != "Gagal mengambil data role" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetRoleByIDService_EmptyID(t *testing.T) {
	roleRepo = &mockRoleRepo{}

	app := fiber.New()
	app.Get("/roles/:id", GetRoleByIDService)

	// id = "  " (spasi) -> strings.TrimSpace jadi ""
	req := httptest.NewRequest(http.MethodGet, "/roles/%20%20", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapRole(t, resp)
	if body["message"] != "Role ID harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetRoleByIDService_NotFound(t *testing.T) {
	roleRepo = &mockRoleRepo{
		GetRoleByIDFn: func(id string) (*model.Role, error) {
			return nil, errors.New("role tidak ditemukan")
		},
	}

	app := fiber.New()
	app.Get("/roles/:id", GetRoleByIDService)

	req := httptest.NewRequest(http.MethodGet, "/roles/r404", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	body := decodeMapRole(t, resp)
	if body["message"] != "Role tidak ditemukan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

// func TestGetRoleByNameService_EmptyName(t *testing.T) {
// 	roleRepo = &mockRoleRepo{}

// 	app := fiber.New()
// 	app.Get("/roles/byname", GetRoleByNameService)

// 	req := httptest.NewRequest(http.MethodGet, "/roles/byname", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusBadRequest {
// 		t.Fatalf("expected 400, got %d", resp.StatusCode)
// 	}
// 	body := decodeMapRole(t, resp)
// 	if body["message"] != "Query 'name' harus diisi" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

func TestCreateRoleService_Success(t *testing.T) {
	roleRepo = &mockRoleRepo{
		// create service kamu kemarin ngecek duplicate pakai GetRoleByName
		GetRoleByNameFn: func(name string) (*model.Role, error) {
			return nil, errors.New("role tidak ditemukan")
		},
		CreateRoleFn: func(req model.CreateRoleRequest) (string, error) {
			if req.Name != "Admin" {
				t.Fatalf("expected name=Admin, got %q", req.Name)
			}
			return "role-id-123", nil
		},
	}

	app := fiber.New()
	app.Post("/roles", CreateRoleService)

	req := httptest.NewRequest(http.MethodPost, "/roles", jsonBodyRole(t, model.CreateRoleRequest{
		Name:        "Admin",
		Description: "admin",
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
	body := decodeMapRole(t, resp)
	if body["message"] != "Role berhasil dibuat" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["id"] != "role-id-123" {
		t.Fatalf("unexpected id: %#v", body["id"])
	}
}

func TestUpdateRoleService_EmptyBody(t *testing.T) {
	roleRepo = &mockRoleRepo{}

	app := fiber.New()
	app.Put("/roles/:id", UpdateRoleService)

	req := httptest.NewRequest(http.MethodPut, "/roles/r1", jsonBodyRole(t, model.UpdateRoleRequest{}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapRole(t, resp)
	if body["message"] != "Minimal ada satu field yang harus diupdate" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateRoleService_NotFound(t *testing.T) {
	roleRepo = &mockRoleRepo{
		UpdateRoleFn: func(id string, req model.UpdateRoleRequest) error {
			return errors.New("role tidak ditemukan")
		},
	}

	app := fiber.New()
	app.Put("/roles/:id", UpdateRoleService)

	req := httptest.NewRequest(http.MethodPut, "/roles/r404", jsonBodyRole(t, model.UpdateRoleRequest{Name: "X"}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	body := decodeMapRole(t, resp)
	if body["message"] != "Role tidak ditemukan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestDeleteRoleService_Success(t *testing.T) {
	roleRepo = &mockRoleRepo{
		DeleteRoleFn: func(id string) error { return nil },
	}

	app := fiber.New()
	app.Delete("/roles/:id", DeleteRoleService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/r1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMapRole(t, resp)
	if body["message"] != "Role berhasil dihapus" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}
