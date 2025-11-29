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

type mockUserRepo struct {
	GetUserByUsernameFn func(username string) (*model.User, error)
	RegisterFn          func(req model.RegisterRequest) (string, error)
	LoginFn             func(email, password string) (*model.User, error)

	// capture args (buat assert)
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

func (m *mockUserRepo) GetUserByEmail(email string) (*model.User, error)                  { return nil, nil }
func (m *mockUserRepo) GetUserByID(id string) (*model.User, error)                        { return nil, nil }
func (m *mockUserRepo) GetAllUsers(page, limit int64) ([]model.User, int64, error)        { return nil, 0, nil }
func (m *mockUserRepo) CreateUser(req model.CreateUserRequest) (string, error)            { return "", nil }
func (m *mockUserRepo) UpdateUser(id string, req model.UpdateUserRequest) error           { return nil }
func (m *mockUserRepo) DeleteUser(id string) error                                        { return nil }
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
