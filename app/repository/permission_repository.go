package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hello-fiber/app/model"
	"strings"
	"time"
)

type PermissionRepository interface {
	GetAllPermissions(page, limit int64) ([]model.Permission, int64, error)
	GetPermissionByID(id string) (*model.Permission, error)
	CreatePermission(req model.CreatePermissionRequest) (string, error)
	UpdatePermission(id string, req model.UpdatePermissionRequest) error
	DeletePermission(id string) error
}

type PermissionRepositoryPostgres struct {
	db *sql.DB
}

func NewPermissionRepositoryPostgres(db *sql.DB) *PermissionRepositoryPostgres {
	return &PermissionRepositoryPostgres{db: db}
}

func (r *PermissionRepositoryPostgres) GetAllPermissions(page, limit int64) ([]model.Permission, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM permissions").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal count permissions: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT id, name, resource, action, description
		FROM permissions
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var perm model.Permission
		if err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
		); err != nil {
			return nil, 0, fmt.Errorf("gagal scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error saat iterasi permissions: %w", err)
	}

	return permissions, total, nil
}

func (r *PermissionRepositoryPostgres) GetPermissionByID(id string) (*model.Permission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, resource, action, description
		FROM permissions
		WHERE id = $1
	`

	var perm model.Permission
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Resource,
		&perm.Action,
		&perm.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("permission tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal query permission: %w", err)
	}

	return &perm, nil
}

func (r *PermissionRepositoryPostgres) CreatePermission(req model.CreatePermissionRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO permissions (id, name, resource, action, description)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		req.Name,
		req.Resource,
		req.Action,
		req.Description,
	).Scan(&id)
	if err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "duplicate key") || strings.Contains(lowerErr, "unique") {
			return "", errors.New("permission dengan kombinasi name, resource, dan action tersebut sudah ada")
		}
		return "", fmt.Errorf("gagal membuat permission: %w", err)
	}

	return id, nil
}

func (r *PermissionRepositoryPostgres) UpdatePermission(id string, req model.UpdatePermissionRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var updates []string
	var args []interface{}
	argIndex := 1

	if strings.TrimSpace(req.Name) != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, strings.TrimSpace(req.Name))
		argIndex++
	}
	if strings.TrimSpace(req.Resource) != "" {
		updates = append(updates, fmt.Sprintf("resource = $%d", argIndex))
		args = append(args, strings.TrimSpace(req.Resource))
		argIndex++
	}
	if strings.TrimSpace(req.Action) != "" {
		updates = append(updates, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, strings.TrimSpace(req.Action))
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
		UPDATE permissions
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argIndex)

	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "duplicate key") || strings.Contains(lowerErr, "unique") {
			return errors.New("permission dengan kombinasi name, resource, dan action tersebut sudah ada")
		}
		return fmt.Errorf("gagal update permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("permission tidak ditemukan")
	}

	return nil
}

func (r *PermissionRepositoryPostgres) DeletePermission(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, "DELETE FROM permissions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("gagal delete permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("permission tidak ditemukan")
	}

	return nil
}
