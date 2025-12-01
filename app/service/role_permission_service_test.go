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

type mockRolePermissionRepo struct {
	GetAllRolePermissionsFn func(page, limit int64, roleID, permissionID string) ([]model.RolePermission, int64, error)
	GetRolePermissionFn     func(roleID, permissionID string) (*model.RolePermission, error)
	GetPermissionsByRoleIDFn func(roleID string) ([]model.Permission, error)
	CreateRolePermissionFn  func(roleID, permissionID string) error
	UpdateRolePermissionFn  func(oldRoleID, oldPermissionID, newRoleID, newPermissionID string) error
	DeleteRolePermissionFn  func(roleID, permissionID string) error
}

func (m *mockRolePermissionRepo) GetAllRolePermissions(page, limit int64, roleID, permissionID string) ([]model.RolePermission, int64, error) {
	if m.GetAllRolePermissionsFn != nil {
		return m.GetAllRolePermissionsFn(page, limit, roleID, permissionID)
	}
	return nil, 0, nil
}
func (m *mockRolePermissionRepo) GetRolePermission(roleID, permissionID string) (*model.RolePermission, error) {
	if m.GetRolePermissionFn != nil {
		return m.GetRolePermissionFn(roleID, permissionID)
	}
	return nil, nil
}
func (m *mockRolePermissionRepo) GetPermissionsByRoleID(roleID string) ([]model.Permission, error) {
	if m.GetPermissionsByRoleIDFn != nil {
		return m.GetPermissionsByRoleIDFn(roleID)
	}
	return nil, nil
}
func (m *mockRolePermissionRepo) CreateRolePermission(roleID, permissionID string) error {
	if m.CreateRolePermissionFn != nil {
		return m.CreateRolePermissionFn(roleID, permissionID)
	}
	return nil
}
func (m *mockRolePermissionRepo) UpdateRolePermission(oldRoleID, oldPermissionID, newRoleID, newPermissionID string) error {
	if m.UpdateRolePermissionFn != nil {
		return m.UpdateRolePermissionFn(oldRoleID, oldPermissionID, newRoleID, newPermissionID)
	}
	return nil
}
func (m *mockRolePermissionRepo) DeleteRolePermission(roleID, permissionID string) error {
	if m.DeleteRolePermissionFn != nil {
		return m.DeleteRolePermissionFn(roleID, permissionID)
	}
	return nil
}

func toJSONReaderRolePermission(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewReader(b)
}

func decodeMapRolePermission(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return out
}

func TestGetAllRolePermissionsService_Success(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{
		GetAllRolePermissionsFn: func(page, limit int64, roleID, permissionID string) ([]model.RolePermission, int64, error) {
			if page != 2 || limit != 5 {
				t.Fatalf("expected page=2 limit=5, got page=%d limit=%d", page, limit)
			}
			if roleID != "r1" || permissionID != "p1" {
				t.Fatalf("unexpected filters: roleID=%q permissionID=%q", roleID, permissionID)
			}
			return []model.RolePermission{{RoleID: "r1", PermissionID: "p1"}}, 1, nil
		},
	}

	app := fiber.New()
	app.Get("/role-permissions", GetAllRolePermissionsService)

	req := httptest.NewRequest(http.MethodGet, "/role-permissions?page=2&limit=5&role_id=r1&permission_id=p1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeMapRolePermission(t, resp)
	if body["message"] != "Data role_permission berhasil diambil" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["total"] != float64(1) {
		t.Fatalf("unexpected total: %#v", body["total"])
	}
	if body["page"] != float64(2) || body["limit"] != float64(5) {
		t.Fatalf("unexpected page/limit: page=%#v limit=%#v", body["page"], body["limit"])
	}
}

func TestGetAllRolePermissionsService_DefaultPagination(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{
		GetAllRolePermissionsFn: func(page, limit int64, roleID, permissionID string) ([]model.RolePermission, int64, error) {
			if page != 1 || limit != 10 {
				t.Fatalf("expected default page=1 limit=10, got page=%d limit=%d", page, limit)
			}
			if roleID != "" || permissionID != "" {
				t.Fatalf("expected empty filters, got roleID=%q permissionID=%q", roleID, permissionID)
			}
			return []model.RolePermission{}, 0, nil
		},
	}

	app := fiber.New()
	app.Get("/role-permissions", GetAllRolePermissionsService)

	req := httptest.NewRequest(http.MethodGet, "/role-permissions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetAllRolePermissionsService_RepoError(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{
		GetAllRolePermissionsFn: func(page, limit int64, roleID, permissionID string) ([]model.RolePermission, int64, error) {
			return nil, 0, errors.New("db down")
		},
	}

	app := fiber.New()
	app.Get("/role-permissions", GetAllRolePermissionsService)

	req := httptest.NewRequest(http.MethodGet, "/role-permissions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	body := decodeMapRolePermission(t, resp)
	if body["message"] != "Gagal mengambil data role_permission" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

// func TestGetRolePermissionDetailService_EmptyRoleID(t *testing.T) {
// 	rolePermissionRepo = &mockRolePermissionRepo{}

// 	app := fiber.New()
// 	app.Get("/role-permissions/:role_id/:permission_id", GetRolePermissionDetailService)

// 	// pakai %20%20 supaya route tetap match, lalu TrimSpace -> ""
// 	req := httptest.NewRequest(http.MethodGet, "/role-permissions/%20%20/p1", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusBadRequest {
// 		t.Fatalf("expected 400, got %d", resp.StatusCode)
// 	}

// 	body := decodeMapRolePermission(t, resp)
// 	if body["message"] != "role_id dan permission_id harus diisi" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

// func TestGetRolePermissionDetailService_NotFound(t *testing.T) {
// 	rolePermissionRepo = &mockRolePermissionRepo{
// 		GetRolePermissionFn: func(roleID, permissionID string) (*model.RolePermission, error) {
// 			return nil, errors.New("role_permission tidak ditemukan")
// 		},
// 	}

// 	app := fiber.New()
// 	app.Get("/role-permissions/:role_id/:permission_id", GetRolePermissionDetailService)

// 	req := httptest.NewRequest(http.MethodGet, "/role-permissions/r1/p1", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusNotFound {
// 		t.Fatalf("expected 404, got %d", resp.StatusCode)
// 	}

// 	body := decodeMapRolePermission(t, resp)
// 	if body["message"] != "role_permission tidak ditemukan" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

// func TestGetRolePermissionDetailService_RepoError(t *testing.T) {
// 	rolePermissionRepo = &mockRolePermissionRepo{
// 		GetRolePermissionFn: func(roleID, permissionID string) (*model.RolePermission, error) {
// 			return nil, errors.New("db down")
// 		},
// 	}

// 	app := fiber.New()
// 	app.Get("/role-permissions/:role_id/:permission_id", GetRolePermissionDetailService)

// 	req := httptest.NewRequest(http.MethodGet, "/role-permissions/r1/p1", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusInternalServerError {
// 		t.Fatalf("expected 500, got %d", resp.StatusCode)
// 	}

// 	body := decodeMapRolePermission(t, resp)
// 	if body["message"] != "Gagal mengambil detail role_permission" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

// func TestGetRolePermissionDetailService_Success(t *testing.T) {
// 	rolePermissionRepo = &mockRolePermissionRepo{
// 		GetRolePermissionFn: func(roleID, permissionID string) (*model.RolePermission, error) {
// 			if roleID != "r1" || permissionID != "p1" {
// 				t.Fatalf("unexpected ids: %s %s", roleID, permissionID)
// 			}
// 			return &model.RolePermission{RoleID: roleID, PermissionID: permissionID}, nil
// 		},
// 	}

// 	app := fiber.New()
// 	app.Get("/role-permissions/:role_id/:permission_id", GetRolePermissionDetailService)

// 	req := httptest.NewRequest(http.MethodGet, "/role-permissions/r1/p1", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatalf("expected 200, got %d", resp.StatusCode)
// 	}

// 	body := decodeMapRolePermission(t, resp)
// 	if body["message"] != "Detail role_permission berhasil diambil" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

func TestGetPermissionsByRoleIDService_EmptyRoleID(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{}

	app := fiber.New()
	app.Get("/roles/:role_id/permissions", GetPermissionsByRoleIDService)

	req := httptest.NewRequest(http.MethodGet, "/roles/%20%20/permissions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "role_id harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetPermissionsByRoleIDService_Success(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{
		GetPermissionsByRoleIDFn: func(roleID string) ([]model.Permission, error) {
			if roleID != "r1" {
				t.Fatalf("expected roleID=r1 got %q", roleID)
			}
			return []model.Permission{
				{ID: "p1", Name: "achievement:read", Resource: "achievement", Action: "read"},
			}, nil
		},
	}

	app := fiber.New()
	app.Get("/roles/:role_id/permissions", GetPermissionsByRoleIDService)

	req := httptest.NewRequest(http.MethodGet, "/roles/r1/permissions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "Data permissions milik role berhasil diambil" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreateRolePermissionService_InvalidBody(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{}

	app := fiber.New()
	app.Post("/role-permissions", CreateRolePermissionService)

	req := httptest.NewRequest(http.MethodPost, "/role-permissions", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "Request body tidak valid" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreateRolePermissionService_MissingFields(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{}

	app := fiber.New()
	app.Post("/role-permissions", CreateRolePermissionService)

	req := httptest.NewRequest(http.MethodPost, "/role-permissions", toJSONReaderRolePermission(t, map[string]any{
		"role_id": "r1",
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
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "role_id dan permission_id harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreateRolePermissionService_Success(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{
		CreateRolePermissionFn: func(roleID, permissionID string) error {
			if roleID != "r1" || permissionID != "p1" {
				t.Fatalf("unexpected ids: %s %s", roleID, permissionID)
			}
			return nil
		},
	}

	app := fiber.New()
	app.Post("/role-permissions", CreateRolePermissionService)

	req := httptest.NewRequest(http.MethodPost, "/role-permissions", toJSONReaderRolePermission(t, map[string]any{
		"role_id":       "r1",
		"permission_id": "p1",
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
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "role_permission berhasil dibuat" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestCreateRolePermissionService_Duplicate(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{
		CreateRolePermissionFn: func(roleID, permissionID string) error {
			return errors.New("role_permission sudah ada")
		},
	}

	app := fiber.New()
	app.Post("/role-permissions", CreateRolePermissionService)

	req := httptest.NewRequest(http.MethodPost, "/role-permissions", toJSONReaderRolePermission(t, map[string]any{
		"role_id":       "r1",
		"permission_id": "p1",
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
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "role_permission sudah ada" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateRolePermissionService_EmptyParams(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{}

	app := fiber.New()
	app.Put("/role-permissions/:role_id/:permission_id", UpdateRolePermissionService)

	req := httptest.NewRequest(http.MethodPut, "/role-permissions/%20%20/p1", toJSONReaderRolePermission(t, map[string]any{
		"new_role_id": "r2",
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
}

func TestUpdateRolePermissionService_InvalidBody(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{}

	app := fiber.New()
	app.Put("/role-permissions/:role_id/:permission_id", UpdateRolePermissionService)

	req := httptest.NewRequest(http.MethodPut, "/role-permissions/r1/p1", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "Request body tidak valid" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateRolePermissionService_NoFields(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{}

	app := fiber.New()
	app.Put("/role-permissions/:role_id/:permission_id", UpdateRolePermissionService)

	req := httptest.NewRequest(http.MethodPut, "/role-permissions/r1/p1", toJSONReaderRolePermission(t, map[string]any{}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "Minimal salah satu dari new_role_id atau new_permission_id harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateRolePermissionService_Success_OnlyNewRoleID(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{
		UpdateRolePermissionFn: func(oldRoleID, oldPermissionID, newRoleID, newPermissionID string) error {
			if oldRoleID != "r1" || oldPermissionID != "p1" {
				t.Fatalf("unexpected old ids: %s %s", oldRoleID, oldPermissionID)
			}
			if newRoleID != "r2" || newPermissionID != "p1" { // new_permission_id kosong -> pakai yg lama
				t.Fatalf("unexpected new ids: %s %s", newRoleID, newPermissionID)
			}
			return nil
		},
	}

	app := fiber.New()
	app.Put("/role-permissions/:role_id/:permission_id", UpdateRolePermissionService)

	req := httptest.NewRequest(http.MethodPut, "/role-permissions/r1/p1", toJSONReaderRolePermission(t, map[string]any{
		"new_role_id": "r2",
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
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "role_permission berhasil diupdate" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateRolePermissionService_NotFound(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{
		UpdateRolePermissionFn: func(oldRoleID, oldPermissionID, newRoleID, newPermissionID string) error {
			return errors.New("role_permission tidak ditemukan")
		},
	}

	app := fiber.New()
	app.Put("/role-permissions/:role_id/:permission_id", UpdateRolePermissionService)

	req := httptest.NewRequest(http.MethodPut, "/role-permissions/r1/p1", toJSONReaderRolePermission(t, map[string]any{
		"new_permission_id": "p2",
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
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "role_permission tidak ditemukan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestDeleteRolePermissionService_Success(t *testing.T) {
	rolePermissionRepo = &mockRolePermissionRepo{
		DeleteRolePermissionFn: func(roleID, permissionID string) error {
			if roleID != "r1" || permissionID != "p1" {
				t.Fatalf("unexpected ids: %s %s", roleID, permissionID)
			}
			return nil
		},
	}

	app := fiber.New()
	app.Delete("/role-permissions/:role_id/:permission_id", DeleteRolePermissionService)

	req := httptest.NewRequest(http.MethodDelete, "/role-permissions/r1/p1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMapRolePermission(t, resp)
	if body["message"] != "role_permission berhasil dihapus" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}
