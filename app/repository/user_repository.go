package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hello-fiber/app/model"
	"hello-fiber/utils"
	"strings"
	"time"
)

type UserRepository interface {
	Register(req model.RegisterRequest) (string, error)
	Login(email, password string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByID(id string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetAllUsers(page, limit int64) ([]model.User, int64, error)
	CreateUser(req model.CreateUserRequest) (string, error)
	UpdateUser(id string, req model.UpdateUserRequest) error
	DeleteUser(id string) error
	GetRoleByID(id string) (*model.Role, error)
	GetRoleByName(name string) (*model.Role, error)
	GetUserPermissions(userID string) ([]model.Permission, error)
}

type UserRepositoryPostgres struct {
	db *sql.DB
}

func NewUserRepositoryPostgres(db *sql.DB) *UserRepositoryPostgres {
	return &UserRepositoryPostgres{db: db}
}

// Register menambahkan user baru ke database PostgreSQL
func (r *UserRepositoryPostgres) Register(req model.RegisterRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return "", fmt.Errorf("gagal hash password: %w", err)
	}

	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, is_active, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, true, NOW(), NOW())
		RETURNING id
	`

	var userID string
	err = r.db.QueryRowContext(
		ctx,
		query,
		strings.TrimSpace(req.Username),
		strings.ToLower(strings.TrimSpace(req.Email)),
		hashedPassword,
		req.FullName,
	).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return "", errors.New("email atau username sudah terdaftar")
		}
		return "", fmt.Errorf("gagal membuat user: %w", err)
	}

	return userID, nil
}

func (r *UserRepositoryPostgres) Login(email, password string) (*model.User, error) {
	user, err := r.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("user tidak aktif")
	}

	// Validasi password
	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("email atau password salah")
	}

	return user, nil
}

// GetUserByEmail mengambil user berdasarkan email
func (r *UserRepositoryPostgres) GetUserByEmail(email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	var roleID sql.NullString
	err := r.db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(email))).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
    	&roleID, // <- ganti dari &user.RoleID
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal query user: %w", err)
	}

	user.RoleID = ""
	if roleID.Valid {
		user.RoleID = roleID.String
	}

	return &user, nil
}

// GetUserByID mengambil user berdasarkan ID
func (r *UserRepositoryPostgres) GetUserByID(id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	var roleID sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&roleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal query user: %w", err)
	}

	user.RoleID = ""
	if roleID.Valid {
		user.RoleID = roleID.String
	}

	return &user, nil
}

// GetUserByUsername mengambil user berdasarkan username
func (r *UserRepositoryPostgres) GetUserByUsername(username string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user model.User
	var roleID sql.NullString
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(username)).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&roleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Username belum ada
		}
		return nil, fmt.Errorf("gagal query user: %w", err)
	}

	user.RoleID = ""
	if roleID.Valid {
		user.RoleID = roleID.String
	}

	return &user, nil
}

// GetAllUsers mengambil semua users dengan pagination
func (r *UserRepositoryPostgres) GetAllUsers(page, limit int64) ([]model.User, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count total
	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal count users: %w", err)
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal query users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		var roleID sql.NullString
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.FullName,
			&roleID,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("[WARNING] Gagal decode user: %v\n", err)
			continue
		}
		user.RoleID = ""
		if roleID.Valid {
			user.RoleID = roleID.String
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}

// CreateUser membuat user baru (Admin)
func (r *UserRepositoryPostgres) CreateUser(req model.CreateUserRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return "", fmt.Errorf("gagal hash password: %w", err)
	}

	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, is_active, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`

	var userID string
	err = r.db.QueryRowContext(
		ctx,
		query,
		strings.TrimSpace(req.Username),
		strings.ToLower(strings.TrimSpace(req.Email)),
		hashedPassword,
		req.FullName,
		req.IsActive,
	).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return "", errors.New("email atau username sudah terdaftar")
		}
		return "", fmt.Errorf("gagal membuat user: %w", err)
	}

	return userID, nil
}

// UpdateUser mengupdate data user
func (r *UserRepositoryPostgres) UpdateUser(id string, req model.UpdateUserRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Username != "" {
		updates = append(updates, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, strings.TrimSpace(req.Username))
		argIndex++
	}
	if req.Email != "" {
		updates = append(updates, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, strings.ToLower(strings.TrimSpace(req.Email)))
		argIndex++
	}
	if req.Password != "" {
		hashed, err := utils.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("gagal hash password: %w", err)
		}
		updates = append(updates, fmt.Sprintf("password_hash = $%d", argIndex))
		args = append(args, hashed)
		argIndex++
	}
	if req.FullName != "" {
		updates = append(updates, fmt.Sprintf("full_name = $%d", argIndex))
		args = append(args, req.FullName)
		argIndex++
	}
	if req.RoleID != "" {
		updates = append(updates, fmt.Sprintf("role_id = $%d", argIndex))
		args = append(args, req.RoleID)
		argIndex++
	}
	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(updates) == 0 {
		return errors.New("tidak ada data yang diupdate")
	}

	updates = append(updates, fmt.Sprintf("updated_at = NOW()"))

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(updates, ", "), argIndex)
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("email atau username sudah terdaftar")
		}
		return fmt.Errorf("gagal update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}

	return nil
}

// DeleteUser menghapus user
func (r *UserRepositoryPostgres) DeleteUser(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("gagal delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}

	return nil
}

// GetRoleByID mengambil role berdasarkan ID
func (r *UserRepositoryPostgres) GetRoleByID(id string) (*model.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, description, created_at
		FROM roles
		WHERE id = $1
	`

	var role model.Role
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal query role: %w", err)
	}

	return &role, nil
}

// GetRoleByName mengambil role berdasarkan nama
func (r *UserRepositoryPostgres) GetRoleByName(name string) (*model.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, description, created_at
		FROM roles
		WHERE name = $1
	`

	var role model.Role
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal query role: %w", err)
	}

	return &role, nil
}

func (r *UserRepositoryPostgres) GetUserPermissions(userID string) ([]model.Permission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN users u ON u.role_id = rp.role_id
		WHERE u.id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var perm model.Permission
		err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description)
		if err != nil {
			fmt.Printf("[WARNING] Gagal decode permission: %v\n", err)
			continue
		}
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}
