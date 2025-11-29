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
// @Summary Daftar user baru
// @Description Membuat user baru dengan validasi email, username, password, dan full_name
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.RegisterRequest true "Data registrasi"
// @Success 201 {object} model.SuccessResponse "User berhasil terdaftar"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /register [post]
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
// @Summary Login user
// @Description Authenticate user dengan email dan password, return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Email dan password"
// @Success 200 {object} model.LoginResponse "Login berhasil"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Email atau password salah"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /login [post]
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

// GetAllUsersService godoc
// @Summary Dapatkan semua user (Admin only)
// @Description Mengambil daftar semua user dengan pagination
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Halaman (default: 1)"
// @Param limit query int false "Jumlah data per halaman (default: 10)"
// @Success 200 {object} model.UserListResponse "User list berhasil diambil"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /users [get]
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

// GetUserByIDService godoc
// @Summary Dapatkan detail user (Admin only)
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
// @Router /users/{id} [get]
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
		// repo kamu pakai error message: "user tidak ditemukan"
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


// CreateUserAdmin godoc
// @Summary Buat user baru (Admin only)
// @Description Admin membuat user baru dengan validasi lengkap
// @Tags Users
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "Data user baru"
// @Success 201 {object} model.SuccessResponse "User berhasil dibuat"
// @Failure 400 {object} model.ErrorResponse "Validasi gagal"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /users [post]
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
// @Summary Update data user (Admin only)
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
// @Router /users/{id} [put]
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
// @Summary Hapus user (Admin only)
// @Description Admin dapat menghapus user berdasarkan ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} model.SuccessResponse "User berhasil dihapus"
// @Failure 400 {object} model.ErrorResponse "User ID tidak valid"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Error server"
// @Router /users/{id} [delete]
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
