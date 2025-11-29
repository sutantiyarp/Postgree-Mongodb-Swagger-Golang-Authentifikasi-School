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
