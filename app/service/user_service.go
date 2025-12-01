package service

import (
	"hello-fiber/app/model"
	"hello-fiber/app/repository"
	"hello-fiber/utils"
	"github.com/gofiber/fiber/v2"
	"regexp"
	"strings"
	"unicode"
	"database/sql"
)

var userRepo repository.UserRepository

func InitUserService(db *sql.DB) {
    userRepo = repository.NewUserRepositoryPostgres(db)
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}

func isValidPassword(password string) bool {
	if len(password) < 5 {
		return false
	}

	hasUpper := true
	hasLower := true
	hasNumber := true

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasNumber = true
		}
	}

	return hasUpper && hasLower && hasNumber
}

func toUserResponse(user *model.User) *model.UserResponse {
	return &model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		RoleID:    user.RoleID,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// Register godoc
// @Summary Daftar users baru
// @Description Membuat users baru dengan validasi email, username, password, dan full_name
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body model.RegisterRequest true "Data registrasi"
// @Success 201 {object} model.SuccessResponse "User berhasil terdaftar"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/auth/register [post]
func Register(c *fiber.Ctx, db *sql.DB) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Request body tidak valid", "error": err.Error()})
	}

	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Username, email, password, dan full_name harus diisi"})
	}

	if !isValidUsername(req.Username) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Username harus 3-50 karakter, hanya alphanumeric dan underscore"})
	}

	if !isValidEmail(req.Email) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Format email tidak valid"})
	}

	if !isValidPassword(req.Password) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Password minimal 5 karakter dengan uppercase, lowercase, dan number"})
	}

	existingUser, err := userRepo.GetUserByUsername(req.Username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal validasi username", "error": err.Error()})
	}
	if existingUser != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Username sudah terdaftar"})
	}

	id, err := userRepo.Register(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal mendaftarkan user", "error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "message": "User berhasil didaftarkan", "id": id})
}

// Login godoc
// @Summary Login users
// @Description Authenticate users dengan email dan password, return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Email dan password"
// @Success 200 {object} model.LoginResponse "Login berhasil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Email atau password salah"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/auth/login [post]
func Login(c *fiber.Ctx, db *sql.DB) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Request body tidak valid", "error": err.Error()})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Email dan password harus diisi"})
	}

	user, err := userRepo.Login(strings.ToLower(strings.TrimSpace(req.Email)), req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	token, err := utils.GenerateJWTPostgres(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal membuat token", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Login berhasil", "token": token, "user": toUserResponse(user)})
}

// GetUserByEmailService godoc
// @Summary Dapatkan detail user berdasarkan email (Admin)
// @Description Mengambil detail user berdasarkan email
// @Tags Users
// @Accept json
// @Produce json
// @Param email query string true "Email user"
// @Success 200 {object} model.UserDetailResponse "Data user berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/users/byemail [get]
// @Security BearerAuth
// func GetUserByEmailService(c *fiber.Ctx) error {
// 	email := strings.ToLower(strings.TrimSpace(c.Query("email")))
// 	if email == "" {
// 		return c.Status(400).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Email harus diisi",
// 		})
// 	}
// 	if !isValidEmail(email) {
// 		return c.Status(400).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Format email tidak valid",
// 		})
// 	}

// 	user, err := userRepo.GetUserByEmail(email)
// 	if err != nil {
// 		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
// 			return c.Status(404).JSON(fiber.Map{
// 				"success": false,
// 				"message": "User tidak ditemukan",
// 			})
// 		}
// 		return c.Status(500).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil data user",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Data user berhasil diambil",
// 		"data":    toUserResponse(user),
// 	})
// }

// GetUserByIDService godoc
// @Summary Dapatkan detail user (Admin)
// @Description Mengambil detail user berdasarkan User ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} model.UserDetailResponse "Data user berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/users/{id} [get]
// @Security BearerAuth
func GetUserByIDService(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "User ID harus diisi",
		})
	}

	user, err := userRepo.GetUserByID(id)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "User tidak ditemukan",
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data user berhasil diambil",
		"data":    toUserResponse(user),
	})
}

// GetAllUsersService godoc
// @Summary Dapatkan semua user (Admin)
// @Description Mengambil daftar semua user dengan pagination
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Success 200 {object} model.UserListResponse "User list berhasil diambil"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/users [get]
// @Security BearerAuth
func GetAllUsersService(c *fiber.Ctx) error {
	page := int64(1)
	limit := int64(10)

	if p := c.Query("page"); p != "" {
		page = int64(c.QueryInt("page", 1))
	}
	if l := c.Query("limit"); l != "" {
		limit = int64(c.QueryInt("limit", 10))
	}

	users, total, err := userRepo.GetAllUsers(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal mengambil data user", "error": err.Error()})
	}

	var userResponses []model.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, *toUserResponse(&user))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data user berhasil diambil",
		"data":    userResponses,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetUserByUsernameService godoc
// @Summary Dapatkan detail users berdasarkan username (Admin)
// @Description Mengambil detail users berdasarkan username
// @Tags Users
// @Accept json
// @Produce json
// @Param username query string true "Username"
// @Success 200 {object} model.UserDetailResponse "Data user berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/users/byusername [get]
// @Security BearerAuth
// func GetUserByUsernameService(c *fiber.Ctx) error {
// 	username := strings.TrimSpace(c.Query("username"))
// 	if username == "" {
// 		return c.Status(400).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Username harus diisi",
// 		})
// 	}
// 	if !isValidUsername(username) {
// 		return c.Status(400).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Username harus 3-50 karakter, hanya alphanumeric dan underscore",
// 		})
// 	}

// 	user, err := userRepo.GetUserByUsername(username)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil data user",
// 			"error":   err.Error(),
// 		})
// 	}
// 	if user == nil {
// 		return c.Status(404).JSON(fiber.Map{
// 			"success": false,
// 			"message": "User tidak ditemukan",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Data user berhasil diambil",
// 		"data":    toUserResponse(user),
// 	})
// }

// GetUsersByRoleNameService godoc
// @Summary Dapatkan user berdasarkan nama role (Admin)
// @Description Mengambil daftar user berdasarkan nama role dengan pagination
// @Tags Users
// @Accept json
// @Produce json
// @Param name query string true "Nama role (contoh: admin)"
// @Param page query int false "Halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Success 200 {object} model.UserListResponse "User list berhasil diambil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "Role tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/users/byrole [get]
// @Security BearerAuth
// func GetUsersByRoleNameService(c *fiber.Ctx) error {
// 	roleName := strings.TrimSpace(c.Query("name"))
// 	if roleName == "" {
// 		return c.Status(400).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Nama role harus diisi",
// 		})
// 	}

// 	page := int64(1)
// 	limit := int64(10)

// 	if p := c.Query("page"); p != "" {
// 		page = int64(c.QueryInt("page", 1))
// 	}
// 	if l := c.Query("limit"); l != "" {
// 		limit = int64(c.QueryInt("limit", 10))
// 	}

// 	users, total, err := userRepo.GetUsersByRoleName(roleName, page, limit)
// 	if err != nil {
// 		if strings.Contains(strings.ToLower(err.Error()), "tidak ditemukan") {
// 			return c.Status(404).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Role tidak ditemukan",
// 			})
// 		}
// 		if strings.Contains(strings.ToLower(err.Error()), "harus diisi") {
// 			return c.Status(400).JSON(fiber.Map{
// 				"success": false,
// 				"message": err.Error(),
// 			})
// 		}

// 		return c.Status(500).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil data user",
// 			"error":   err.Error(),
// 		})
// 	}

// 	var userResponses []model.UserResponse
// 	for _, u := range users {
// 		userResponses = append(userResponses, *toUserResponse(&u))
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Data user berhasil diambil",
// 		"data":    userResponses,
// 		"total":   total,
// 		"page":    page,
// 		"limit":   limit,
// 	})
// }

// CreateUserAdmin godoc
// @Summary Buat users baru (Admin)
// @Description Admin membuat users baru dengan validasi lengkap
// @Tags Users
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "Data user baru"
// @Success 201 {object} model.SuccessResponse "User berhasil dibuat"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/users [post]
// @Security BearerAuth
func CreateUserAdmin(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Request body tidak valid", "error": err.Error()})
	}

	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Username, email, password, dan full_name harus diisi"})
	}

	if !isValidUsername(req.Username) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Username harus 3-50 karakter, hanya alphanumeric dan underscore"})
	}

	if !isValidEmail(req.Email) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Format email tidak valid"})
	}

	if !isValidPassword(req.Password) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Password minimal 5 karakter dengan uppercase, lowercase, dan number"})
	}

	existingUser, err := userRepo.GetUserByUsername(req.Username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal validasi username", "error": err.Error()})
	}
	if existingUser != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Username sudah terdaftar"})
	}

	id, err := userRepo.CreateUser(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal membuat user", "error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "message": "User berhasil dibuat", "id": id})
}

// UpdateUserService godoc
// @Summary Update data users (Admin)
// @Description Admin dapat update username, email, password, role, atau is_active
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param body body model.UpdateUserRequest true "Data user yang diupdate"
// @Success 200 {object} model.SuccessResponse "User berhasil diupdate"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 404 {object} model.ErrorResponse "User tidak ditemukan"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/users/{id} [put]
// @Security BearerAuth
func UpdateUserService(c *fiber.Ctx) error {
	userID := c.Params("id")
	var req model.UpdateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Request body tidak valid", "error": err.Error()})
	}

	hasUpdate := req.Username != "" || req.Email != "" || req.Password != "" || req.RoleID != "" || req.FullName != "" || req.IsActive != nil
	if !hasUpdate {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Minimal ada satu field yang harus diupdate"})
	}

	if req.Username != "" && !isValidUsername(req.Username) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Username harus 3-50 karakter, hanya alphanumeric dan underscore"})
	}

	if req.Email != "" && !isValidEmail(req.Email) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Format email tidak valid"})
	}

	if req.Password != "" && !isValidPassword(req.Password) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Password minimal 5 karakter dengan uppercase, lowercase, dan number"})
	}

	if req.Username != "" {
		existingUser, err := userRepo.GetUserByUsername(req.Username)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal validasi username", "error": err.Error()})
		}
		if existingUser != nil && existingUser.ID != userID {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Username sudah terdaftar"})
		}
	}

	if err := userRepo.UpdateUser(userID, req); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal update user", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "User berhasil diupdate"})
}

// DeleteUserService godoc
// @Summary Hapus users (Admin)
// @Description Admin dapat menghapus users berdasarkan ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} model.SuccessResponse "User berhasil dihapus"
// @Failure 400 {object} model.ErrorResponse "User ID tidak valid"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /v1/users/{id} [delete]
// @Security BearerAuth
func DeleteUserService(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "User ID harus diisi"})
	}

	if err := userRepo.DeleteUser(userID); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal delete user", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "User berhasil dihapus"})
}
