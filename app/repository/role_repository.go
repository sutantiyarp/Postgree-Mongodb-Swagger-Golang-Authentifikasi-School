package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hello-fiber/app/model"
	// "hello-fiber/utils"
	"strings"
	"time"
)

type RoleRepository interface {
	GetAllRoles(page, limit int64) ([]model.Role, int64, error)
	GetRoleByID(id string) (*model.Role, error)
	GetRoleByName(name string) (*model.Role, error)
	CreateRole(req model.CreateRoleRequest) (string, error)
	UpdateRole(id string, req model.UpdateRoleRequest) error
	DeleteRole(id string) error
}

type RoleRepositoryPostgres struct {
	db *sql.DB
}

func NewRoleRepositoryPostgres(db *sql.DB) *RoleRepositoryPostgres {
	return &RoleRepositoryPostgres{db: db}
}

func (r *RoleRepositoryPostgres) GetAllRoles(page, limit int64) ([]model.Role, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// total
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM roles").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal count roles: %w", err)
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, name, description, created_at
		FROM roles
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal query roles: %w", err)
	}
	defer rows.Close()

	roles := make([]model.Role, 0)
	for rows.Next() {
		var role model.Role
		var desc sql.NullString

		if err := rows.Scan(&role.ID, &role.Name, &desc, &role.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan role: %w", err)
		}

		role.Description = ""
		if desc.Valid {
			role.Description = desc.String
		}

		roles = append(roles, role)
	}

	return roles, total, rows.Err()
}

func (r *RoleRepositoryPostgres) GetRoleByID(id string) (*model.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, description, created_at
		FROM roles
		WHERE id = $1
	`

	var role model.Role
	var desc sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(&role.ID, &role.Name, &desc, &role.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal query role: %w", err)
	}

	if desc.Valid {
		role.Description = desc.String
	} else {
		role.Description = ""
	}
	return &role, nil
}

func (r *RoleRepositoryPostgres) GetRoleByName(name string) (*model.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT id, name, description, created_at
	FROM roles
	WHERE LOWER(name) = LOWER($1)
	`

	var role model.Role
	var desc sql.NullString

	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(name)).Scan(&role.ID, &role.Name, &desc, &role.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal query role: %w", err)
	}

	if desc.Valid {
		role.Description = desc.String
	} else {
		role.Description = ""
	}
	return &role, nil
}

func (r *RoleRepositoryPostgres) CreateRole(req model.CreateRoleRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := strings.TrimSpace(req.Name)
	desc := strings.TrimSpace(req.Description)

	if name == "" {
		return "", errors.New("nama role tidak boleh kosong")
	}

	query := `
		INSERT INTO roles (id, name, description, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		RETURNING id
	`

	var roleID string
	err := r.db.QueryRowContext(ctx, query, name, desc).Scan(&roleID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return "", errors.New("role dengan nama tersebut sudah ada")
		}
		return "", fmt.Errorf("gagal membuat role: %w", err)
	}

	return roleID, nil
}

func (r *RoleRepositoryPostgres) UpdateRole(id string, req model.UpdateRoleRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if strings.TrimSpace(req.Name) != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, strings.TrimSpace(req.Name))
		argIndex++
	}
	if req.Description != "" {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, strings.TrimSpace(req.Description))
		argIndex++
	}

	if len(updates) == 0 {
		return errors.New("tidak ada field yang diupdate")
	}

	query := fmt.Sprintf(`
		UPDATE roles
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argIndex)

	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("gagal update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("role tidak ditemukan")
	}

	return nil
}

func (r *RoleRepositoryPostgres) DeleteRole(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, "DELETE FROM roles WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("gagal delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("role tidak ditemukan")
	}

	return nil
}
