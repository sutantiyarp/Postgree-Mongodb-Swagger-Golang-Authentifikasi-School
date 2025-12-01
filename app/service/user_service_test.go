package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	// "net/url"
	"testing"

	"hello-fiber/app/model"

	"github.com/gofiber/fiber/v2"
)

type mockUserRepo struct {
	GetUserByUsernameFn func(username string) (*model.User, error)
	RegisterFn          func(req model.RegisterRequest) (string, error)
	LoginFn             func(email, password string) (*model.User, error)

	GetUserByEmailFn func(email string) (*model.User, error)
	GetUserByIDFn  func(id string) (*model.User, error)
	GetAllUsersFn  func(page, limit int64) ([]model.User, int64, error)
	GetUsersByRoleNameFn func(roleName string, page, limit int64) ([]model.User, int64, error)
	CreateUserFn   func(req model.CreateUserRequest) (string, error)
	UpdateUserFn   func(id string, req model.UpdateUserRequest) error
	DeleteUserFn   func(id string) error

	GetAllRolesFn       func(page, limit int64) ([]model.Role, int64, error)
	GetRoleByIDFn        func(id string) (*model.Role, error)
	GetRoleByNameFn      func(name string) (*model.Role, error)
	GetUserPermissionsFn func(userID string) ([]model.Permission, error)

	LastLoginEmail    string
	LastLoginPassword string
	LastRegisterReq   *model.RegisterRequest
}

func (m *mockUserRepo) Register(req model.RegisterRequest) (string, error) {
	m.LastRegisterReq = &req
	if m.RegisterFn != nil {
		return m.RegisterFn(req)
	}
	return "mock-id", nil
}

func (m *mockUserRepo) Login(email, password string) (*model.User, error) {
	m.LastLoginEmail = email
	m.LastLoginPassword = password
	if m.LoginFn != nil {
		return m.LoginFn(email, password)
	}
	return &model.User{ID: "u1", Email: email, RoleID: "user", IsActive: true}, nil
}

func (m *mockUserRepo) GetUserByUsername(username string) (*model.User, error) {
	if m.GetUserByUsernameFn != nil {
		return m.GetUserByUsernameFn(username)
	}
	return nil, nil
}

func (m *mockUserRepo) GetUserByEmail(email string) (*model.User, error) {
	if m.GetUserByEmailFn != nil {
		return m.GetUserByEmailFn(email)
	}
	return nil, nil
}

func (m *mockUserRepo) GetUserByID(id string) (*model.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetAllUsers(page, limit int64) ([]model.User, int64, error) {
	if m.GetAllUsersFn != nil {
		return m.GetAllUsersFn(page, limit)
	}
	return nil, 0, nil
}

func (m *mockUserRepo) GetUsersByRoleName(roleName string, page, limit int64) ([]model.User, int64, error) {
	if m.GetUsersByRoleNameFn != nil {
		return m.GetUsersByRoleNameFn(roleName, page, limit)
	}
	return nil, 0, nil
}

func (m *mockUserRepo) CreateUser(req model.CreateUserRequest) (string, error) {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(req)
	}
	return "", nil
}

func (m *mockUserRepo) UpdateUser(id string, req model.UpdateUserRequest) error {
	if m.UpdateUserFn != nil {
		return m.UpdateUserFn(id, req)
	}
	return nil
}

func (m *mockUserRepo) DeleteUser(id string) error {
	if m.DeleteUserFn != nil {
		return m.DeleteUserFn(id)
	}
	return nil
}

func (m *mockUserRepo) GetAllRoles(page, limit int64) ([]model.Role, int64, error) { return nil, 0, nil }
func (m *mockUserRepo) GetRoleByID(id string) (*model.Role, error)                        { return nil, nil }
func (m *mockUserRepo) GetRoleByName(name string) (*model.Role, error)                    { return nil, nil }
func (m *mockUserRepo) GetUserPermissions(userID string) ([]model.Permission, error)      { return nil, nil }

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewReader(b)
}

func decodeMap(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return out
}

//REGISTER Test
func TestRegister_Success(t *testing.T) {
	mock := &mockUserRepo{
		GetUserByUsernameFn: func(username string) (*model.User, error) {
			return nil, nil
		},
		RegisterFn: func(req model.RegisterRequest) (string, error) {
			return "user-id-123", nil
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Post("/register", func(c *fiber.Ctx) error { return Register(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, model.RegisterRequest{
		Username: "user_1",
		Email:    "test@example.com",
		Password: "Abcd1",
		FullName: "User One",
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

	body := decodeMap(t, resp)
	if body["success"] != true {
		t.Fatalf("expected success=true, got %#v", body["success"])
	}
	if body["message"] != "User berhasil didaftarkan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["id"] != "user-id-123" {
		t.Fatalf("unexpected id: %#v", body["id"])
	}
}

func TestRegister_UsernameAlreadyExists(t *testing.T) {
	mock := &mockUserRepo{
		GetUserByUsernameFn: func(username string) (*model.User, error) {
			return &model.User{ID: "existing"}, nil
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Post("/register", func(c *fiber.Ctx) error { return Register(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, model.RegisterRequest{
		Username: "user_1",
		Email:    "test@example.com",
		Password: "Abcd1",
		FullName: "User One",
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
	body := decodeMap(t, resp)
	if body["message"] != "Username sudah terdaftar" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	userRepo = &mockUserRepo{}

	app := fiber.New()
	app.Post("/register", func(c *fiber.Ctx) error { return Register(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, model.RegisterRequest{
		Username: "user_1",
		Email:    "bukan-email",
		Password: "Abcd1",
		FullName: "User One",
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
	body := decodeMap(t, resp)
	if body["message"] != "Format email tidak valid" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestRegister_InvalidPasswordTooShort(t *testing.T) {
	userRepo = &mockUserRepo{}

	app := fiber.New()
	app.Post("/register", func(c *fiber.Ctx) error { return Register(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, model.RegisterRequest{
		Username: "user_1",
		Email:    "test@example.com",
		Password: "Ab1", // < 5
		FullName: "User One",
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
	body := decodeMap(t, resp)
	if body["message"] != "Password minimal 5 karakter dengan uppercase, lowercase, dan number" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestRegister_GetUserByUsernameError(t *testing.T) {
	mock := &mockUserRepo{
		GetUserByUsernameFn: func(username string) (*model.User, error) {
			return nil, errors.New("db error")
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Post("/register", func(c *fiber.Ctx) error { return Register(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, model.RegisterRequest{
		Username: "user_1",
		Email:    "test@example.com",
		Password: "Abcd1",
		FullName: "User One",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "Gagal validasi username" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestRegister_EmptyPassword(t *testing.T) {
	userRepo = &mockUserRepo{}

	app := fiber.New()
	app.Post("/register", func(c *fiber.Ctx) error { return Register(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, model.RegisterRequest{
		Username: "user_1",
		Email:    "test@example.com",
		Password: "",
		FullName: "User One",
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

	body := decodeMap(t, resp)
	if body["message"] != "Username, email, password, dan full_name harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestRegister_EmptyFullName(t *testing.T) {
	userRepo = &mockUserRepo{}

	app := fiber.New()
	app.Post("/register", func(c *fiber.Ctx) error { return Register(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, model.RegisterRequest{
		Username: "user_1",
		Email:    "test@example.com",
		Password: "Abcd1",
		FullName: "",
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

	body := decodeMap(t, resp)
	if body["message"] != "Username, email, password, dan full_name harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestRegister_EmptyEmail(t *testing.T) {
	userRepo = &mockUserRepo{}

	app := fiber.New()
	app.Post("/register", func(c *fiber.Ctx) error { return Register(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, model.RegisterRequest{
		Username: "user_1",
		Email:    "",
		Password: "Abcd1",
		FullName: "User One",
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

	body := decodeMap(t, resp)
	if body["message"] != "Username, email, password, dan full_name harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}


//LOGIN Test
func TestLogin_Success(t *testing.T) {
	mock := &mockUserRepo{
		LoginFn: func(email, password string) (*model.User, error) {
			return &model.User{
				ID:       "u1",
				Email:    email,
				Username: "user_1",
				FullName: "User One",
				RoleID:   "user",
				IsActive: true,
			}, nil
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Post("/login", func(c *fiber.Ctx) error { return Login(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/login", jsonBody(t, model.LoginRequest{
		Email:    "  TEST@Example.com  ",
		Password: "whatever",
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

	if mock.LastLoginEmail != "test@example.com" {
		t.Fatalf("expected normalized email 'test@example.com', got %q", mock.LastLoginEmail)
	}

	body := decodeMap(t, resp)
	if body["success"] != true {
		t.Fatalf("expected success=true, got %#v", body["success"])
	}
	if body["message"] != "Login berhasil" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if tok, _ := body["token"].(string); tok == "" {
		t.Fatalf("expected non-empty token")
	}
	if body["user"] == nil {
		t.Fatalf("expected user object in response")
	}
}

func TestLogin_MissingFields(t *testing.T) {
	userRepo = &mockUserRepo{}

	app := fiber.New()
	app.Post("/login", func(c *fiber.Ctx) error { return Login(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/login", jsonBody(t, model.LoginRequest{
		Email:    "",
		Password: "",
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
	body := decodeMap(t, resp)
	if body["message"] != "Email dan password harus diisi" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestLogin_UnauthorizedFromRepo(t *testing.T) {
	mock := &mockUserRepo{
		LoginFn: func(email, password string) (*model.User, error) {
			return nil, errors.New("email atau password salah")
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Post("/login", func(c *fiber.Ctx) error { return Login(c, nil) })

	req := httptest.NewRequest(http.MethodPost, "/login", jsonBody(t, model.LoginRequest{
		Email:    "test@example.com",
		Password: "bad",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "email atau password salah" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

//GET ALL USERS Test
func TestGetAllUsersService_Success_DefaultPagination(t *testing.T) {
	mock := &mockUserRepo{
		GetAllUsersFn: func(page, limit int64) ([]model.User, int64, error) {
			if page != 1 || limit != 10 {
				t.Fatalf("expected default page=1 limit=10, got page=%d limit=%d", page, limit)
			}
			return []model.User{
				{
					ID:       "u1",
					Username: "user1",
					Email:    "u1@mail.com",
					FullName: "User One",
					RoleID:   "",
					IsActive: true,
				},
				{
					ID:       "u2",
					Username: "user2",
					Email:    "u2@mail.com",
					FullName: "User Two",
					RoleID:   "role-x",
					IsActive: false,
				},
			}, 2, nil
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Get("/users", GetAllUsersService)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeMap(t, resp)
	if body["success"] != true {
		t.Fatalf("expected success=true, got %#v", body["success"])
	}
	if body["message"] != "Data user berhasil diambil" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["total"] != float64(2) { // angka JSON -> float64
		t.Fatalf("unexpected total: %#v", body["total"])
	}

	data, ok := body["data"].([]any)
	if !ok || len(data) != 2 {
		t.Fatalf("expected 2 users in data, got %#v", body["data"])
	}
}

func TestGetAllUsersService_RepoError(t *testing.T) {
	mock := &mockUserRepo{
		GetAllUsersFn: func(page, limit int64) ([]model.User, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Get("/users", GetAllUsersService)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	body := decodeMap(t, resp)
	if body["message"] != "Gagal mengambil data user" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetUserByIDService_Success(t *testing.T) {
	mock := &mockUserRepo{
		GetUserByIDFn: func(id string) (*model.User, error) {
			if id != "u1" {
				t.Fatalf("expected id u1, got %q", id)
			}
			return &model.User{
				ID:       "u1",
				Username: "user1",
				Email:    "u1@mail.com",
				FullName: "User One",
				RoleID:   "",
				IsActive: true,
			}, nil
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Get("/users/:id", GetUserByIDService)

	req := httptest.NewRequest(http.MethodGet, "/users/u1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeMap(t, resp)
	if body["success"] != true {
		t.Fatalf("expected success=true, got %#v", body["success"])
	}
	if body["message"] != "Data user berhasil diambil" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %#v", body["data"])
	}
	if data["id"] != "u1" {
		t.Fatalf("unexpected id: %#v", data["id"])
	}
}

func TestGetUserByIDService_NotFound(t *testing.T) {
	mock := &mockUserRepo{
		GetUserByIDFn: func(id string) (*model.User, error) {
			return nil, errors.New("user tidak ditemukan")
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Get("/users/:id", GetUserByIDService)

	req := httptest.NewRequest(http.MethodGet, "/users/u404", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "User tidak ditemukan" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestGetUserByIDService_RepoError(t *testing.T) {
	mock := &mockUserRepo{
		GetUserByIDFn: func(id string) (*model.User, error) {
			return nil, errors.New("db error")
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Get("/users/:id", GetUserByIDService)

	req := httptest.NewRequest(http.MethodGet, "/users/u1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "Gagal mengambil data user" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

// func TestGetUserByEmailService_MissingEmail(t *testing.T) {
// 	userRepo = &mockUserRepo{}

// 	app := fiber.New()
// 	app.Get("/users/byemail", GetUserByEmailService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byemail", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusBadRequest {
// 		t.Fatalf("expected 400, got %d", resp.StatusCode)
// 	}

// 	body := decodeMap(t, resp)
// 	if body["message"] != "Email harus diisi" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

// func TestGetUserByEmailService_InvalidEmail(t *testing.T) {
// 	userRepo = &mockUserRepo{}

// 	app := fiber.New()
// 	app.Get("/users/byemail", GetUserByEmailService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byemail?email=bukan-email", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusBadRequest {
// 		t.Fatalf("expected 400, got %d", resp.StatusCode)
// 	}

// 	body := decodeMap(t, resp)
// 	if body["message"] != "Format email tidak valid" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

// func TestGetUserByEmailService_NotFound(t *testing.T) {
// 	mock := &mockUserRepo{
// 		GetUserByEmailFn: func(email string) (*model.User, error) {
// 			return nil, errors.New("user tidak ditemukan")
// 		},
// 	}
// 	userRepo = mock

// 	app := fiber.New()
// 	app.Get("/users/byemail", GetUserByEmailService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byemail?email=test@example.com", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusNotFound {
// 		t.Fatalf("expected 404, got %d", resp.StatusCode)
// 	}

// 	body := decodeMap(t, resp)
// 	if body["message"] != "User tidak ditemukan" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

// func TestGetUserByEmailService_Success(t *testing.T) {
// 	mock := &mockUserRepo{
// 		GetUserByEmailFn: func(email string) (*model.User, error) {
// 			// service normalisasi ke lower+trim
// 			if email != "test@example.com" {
// 				t.Fatalf("expected email=test@example.com, got %q", email)
// 			}
// 			return &model.User{
// 				ID:       "u1",
// 				Username: "user1",
// 				Email:    email,
// 				FullName: "User One",
// 				RoleID:   "",
// 				IsActive: true,
// 			}, nil
// 		},
// 	}
// 	userRepo = mock

//     app := fiber.New()
//     app.Get("/users/byemail", GetUserByEmailService)

//     email := url.QueryEscape("  TEST@Example.com  ")
//     req := httptest.NewRequest(http.MethodGet, "/users/byemail?email="+email, nil)

// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatalf("expected 200, got %d", resp.StatusCode)
// 	}

// 	body := decodeMap(t, resp)
// 	if body["success"] != true {
// 		t.Fatalf("expected success=true, got %#v", body["success"])
// 	}
// }

// func TestGetUserByUsernameService_MissingUsername(t *testing.T) {
// 	userRepo = &mockUserRepo{}

// 	app := fiber.New()
// 	app.Get("/users/byusername", GetUserByUsernameService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byusername", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusBadRequest {
// 		t.Fatalf("expected 400, got %d", resp.StatusCode)
// 	}

// 	body := decodeMap(t, resp)
// 	if body["message"] != "Username harus diisi" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

// func TestGetUserByUsernameService_InvalidUsername(t *testing.T) {
// 	userRepo = &mockUserRepo{}

// 	app := fiber.New()
// 	app.Get("/users/byusername", GetUserByUsernameService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byusername?username=!!", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusBadRequest {
// 		t.Fatalf("expected 400, got %d", resp.StatusCode)
// 	}

// 	body := decodeMap(t, resp)
// 	if body["message"] != "Username harus 3-50 karakter, hanya alphanumeric dan underscore" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

// func TestGetUserByUsernameService_NotFound(t *testing.T) {
// 	mock := &mockUserRepo{
// 		GetUserByUsernameFn: func(username string) (*model.User, error) {
// 			return nil, nil // repo kamu: not found => nil, nil
// 		},
// 	}
// 	userRepo = mock

// 	app := fiber.New()
// 	app.Get("/users/byusername", GetUserByUsernameService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byusername?username=user_1", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusNotFound {
// 		t.Fatalf("expected 404, got %d", resp.StatusCode)
// 	}

// 	body := decodeMap(t, resp)
// 	if body["message"] != "User tidak ditemukan" {
// 		t.Fatalf("unexpected message: %#v", body["message"])
// 	}
// }

// func TestGetUserByUsernameService_Success(t *testing.T) {
// 	mock := &mockUserRepo{
// 		GetUserByUsernameFn: func(username string) (*model.User, error) {
// 			if username != "user_1" {
// 				t.Fatalf("expected username=user_1, got %q", username)
// 			}
// 			return &model.User{
// 				ID:       "u1",
// 				Username: username,
// 				Email:    "u1@mail.com",
// 				FullName: "User One",
// 				RoleID:   "",
// 				IsActive: true,
// 			}, nil
// 		},
// 	}
// 	userRepo = mock

// 	app := fiber.New()
// 	app.Get("/users/byusername", GetUserByUsernameService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byusername?username=user_1", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatalf("expected 200, got %d", resp.StatusCode)
// 	}

// 	body := decodeMap(t, resp)
// 	if body["success"] != true {
// 		t.Fatalf("expected success=true, got %#v", body["success"])
// 	}
// }

// func TestGetUsersByRoleNameService_MissingName(t *testing.T) {
// 	userRepo = &mockUserRepo{}

// 	app := fiber.New()
// 	app.Get("/users/byrole", GetUsersByRoleNameService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byrole", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusBadRequest {
// 		t.Fatalf("expected 400, got %d", resp.StatusCode)
// 	}
// }

// func TestGetUsersByRoleNameService_Success(t *testing.T) {
// 	mock := &mockUserRepo{
// 		GetUsersByRoleNameFn: func(roleName string, page, limit int64) ([]model.User, int64, error) {
// 			if roleName != "admin" {
// 				t.Fatalf("expected roleName=admin, got %q", roleName)
// 			}
// 			return []model.User{
// 				{ID: "u1", Username: "user1", Email: "u1@mail.com", FullName: "User One", RoleID: "role-admin", IsActive: true},
// 			}, 1, nil
// 		},
// 	}
// 	userRepo = mock

// 	app := fiber.New()
// 	app.Get("/users/byrole", GetUsersByRoleNameService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byrole?name=admin&page=1&limit=10", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatalf("expected 200, got %d", resp.StatusCode)
// 	}

// 	body := decodeMap(t, resp)
// 	if body["success"] != true {
// 		t.Fatalf("expected success=true, got %#v", body["success"])
// 	}
// }

// func TestGetUsersByRoleNameService_RoleNotFound(t *testing.T) {
// 	mock := &mockUserRepo{
// 		GetUsersByRoleNameFn: func(roleName string, page, limit int64) ([]model.User, int64, error) {
// 			return nil, 0, errors.New("role tidak ditemukan")
// 		},
// 	}
// 	userRepo = mock

// 	app := fiber.New()
// 	app.Get("/users/byrole", GetUsersByRoleNameService)

// 	req := httptest.NewRequest(http.MethodGet, "/users/byrole?name=unknown", nil)
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("app.Test: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusNotFound {
// 		t.Fatalf("expected 404, got %d", resp.StatusCode)
// 	}
// }

//CREATE USER ADMIN Test
func TestCreateUserAdmin_Success(t *testing.T) {
	mock := &mockUserRepo{
		GetUserByUsernameFn: func(username string) (*model.User, error) {
			return nil, nil
		},
		CreateUserFn: func(req model.CreateUserRequest) (string, error) {
			if req.Username != "user1" {
				t.Fatalf("unexpected username: %q", req.Username)
			}
			return "new-id-123", nil
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Post("/users", CreateUserAdmin)

	reqBody := model.CreateUserRequest{
		Username: "user1",
		Email:    "user1@mail.com",
		Password: "Abcd1",
		FullName: "User One",
	}

	req := httptest.NewRequest(http.MethodPost, "/users", jsonBody(t, reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["success"] != true {
		t.Fatalf("expected success=true, got %#v", body["success"])
	}
	if body["message"] != "User berhasil dibuat" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
	if body["id"] != "new-id-123" {
		t.Fatalf("unexpected id: %#v", body["id"])
	}
}

func TestCreateUserAdmin_UsernameAlreadyExists(t *testing.T) {
	mock := &mockUserRepo{
		GetUserByUsernameFn: func(username string) (*model.User, error) {
			return &model.User{ID: "exists"}, nil
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Post("/users", CreateUserAdmin)

	reqBody := model.CreateUserRequest{
		Username: "user1",
		Email:    "user1@mail.com",
		Password: "Abcd1",
		FullName: "User One",
	}

	req := httptest.NewRequest(http.MethodPost, "/users", jsonBody(t, reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "Username sudah terdaftar" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

//UPDATE USER Test
func TestUpdateUserService_Success(t *testing.T) {
	mock := &mockUserRepo{
		UpdateUserFn: func(id string, req model.UpdateUserRequest) error {
			if id != "u1" {
				t.Fatalf("expected id u1, got %q", id)
			}
			if req.FullName != "Nama Baru" {
				t.Fatalf("unexpected full_name: %#v", req.FullName)
			}
			return nil
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Put("/users/:id", UpdateUserService)

	reqBody := model.UpdateUserRequest{
		FullName: "Nama Baru",
	}

	req := httptest.NewRequest(http.MethodPut, "/users/u1", jsonBody(t, reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "User berhasil diupdate" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateUserService_NoFields(t *testing.T) {
	userRepo = &mockUserRepo{} // repo tidak kepakai

	app := fiber.New()
	app.Put("/users/:id", UpdateUserService)

	req := httptest.NewRequest(http.MethodPut, "/users/u1", jsonBody(t, model.UpdateUserRequest{}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "Minimal ada satu field yang harus diupdate" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateUserService_UsernameAlreadyExists(t *testing.T) {
	mock := &mockUserRepo{
		GetUserByUsernameFn: func(username string) (*model.User, error) {
			return &model.User{ID: "u2"}, nil // beda ID dengan yang di path
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Put("/users/:id", UpdateUserService)

	reqBody := model.UpdateUserRequest{
		Username: "user1",
	}

	req := httptest.NewRequest(http.MethodPut, "/users/u1", jsonBody(t, reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "Username sudah terdaftar" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestUpdateUserService_UpdateError(t *testing.T) {
	mock := &mockUserRepo{
		UpdateUserFn: func(id string, req model.UpdateUserRequest) error {
			return errors.New("db error")
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Put("/users/:id", UpdateUserService)

	reqBody := model.UpdateUserRequest{
		FullName: "Nama Baru",
	}

	req := httptest.NewRequest(http.MethodPut, "/users/u1", jsonBody(t, reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "Gagal update user" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

//DELETE USER Test
func TestDeleteUserService_Success(t *testing.T) {
	mock := &mockUserRepo{
		DeleteUserFn: func(id string) error {
			if id != "u1" {
				t.Fatalf("expected id u1, got %q", id)
			}
			return nil
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Delete("/users/:id", DeleteUserService)

	req := httptest.NewRequest(http.MethodDelete, "/users/u1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "User berhasil dihapus" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}

func TestDeleteUserService_Error(t *testing.T) {
	mock := &mockUserRepo{
		DeleteUserFn: func(id string) error {
			return errors.New("db error")
		},
	}
	userRepo = mock

	app := fiber.New()
	app.Delete("/users/:id", DeleteUserService)

	req := httptest.NewRequest(http.MethodDelete, "/users/u1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["message"] != "Gagal delete user" {
		t.Fatalf("unexpected message: %#v", body["message"])
	}
}
