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

type RolePermissionRepository interface {
	GetAllRolePermissions(page, limit int64, roleID, permissionID string) ([]model.RolePermission, int64, error)
	GetRolePermission(roleID, permissionID string) (*model.RolePermission, error)
	GetPermissionsByRoleID(roleID string) ([]model.Permission, error)
	CreateRolePermission(roleID, permissionID string) error
	UpdateRolePermission(oldRoleID, oldPermissionID, newRoleID, newPermissionID string) error
	DeleteRolePermission(roleID, permissionID string) error
}

type RolePermissionRepositoryPostgres struct {
	db *sql.DB
}

func NewRolePermissionRepositoryPostgres(db *sql.DB) *RolePermissionRepositoryPostgres {
	return &RolePermissionRepositoryPostgres{db: db}
}

func (r *RolePermissionRepositoryPostgres) GetAllRolePermissions(page, limit int64, roleID, permissionID string) ([]model.RolePermission, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	where := []string{}
	args := []interface{}{}
	i := 1

	roleID = strings.TrimSpace(roleID)
	permissionID = strings.TrimSpace(permissionID)

	if roleID != "" {
		where = append(where, fmt.Sprintf("role_id = $%d", i))
		args = append(args, roleID)
		i++
	}
	if permissionID != "" {
		where = append(where, fmt.Sprintf("permission_id = $%d", i))
		args = append(args, permissionID)
		i++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	// total
	var total int64
	countQuery := "SELECT COUNT(*) FROM role_permissions" + whereSQL
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal count role_permissions: %w", err)
	}

	// list
	query := fmt.Sprintf(`
		SELECT role_id, permission_id
		FROM role_permissions
		%s
		ORDER BY role_id, permission_id
		LIMIT $%d OFFSET $%d
	`, whereSQL, i, i+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal query role_permissions: %w", err)
	}
	defer rows.Close()

	var out []model.RolePermission
	for rows.Next() {
		var rp model.RolePermission
		if err := rows.Scan(&rp.RoleID, &rp.PermissionID); err != nil {
			return nil, 0, fmt.Errorf("gagal scan role_permission: %w", err)
		}
		out = append(out, rp)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterasi role_permissions: %w", err)
	}

	return out, total, nil
}

func (r *RolePermissionRepositoryPostgres) GetRolePermission(roleID, permissionID string) (*model.RolePermission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT role_id, permission_id
		FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
	`

	var rp model.RolePermission
	err := r.db.QueryRowContext(ctx, query, roleID, permissionID).Scan(&rp.RoleID, &rp.PermissionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role_permission tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal get role_permission: %w", err)
	}
	return &rp, nil
}

func (r *RolePermissionRepositoryPostgres) GetPermissionsByRoleID(roleID string) ([]model.Permission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("gagal query permissions by role: %w", err)
	}
	defer rows.Close()

	var out []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description); err != nil {
			return nil, fmt.Errorf("gagal scan permission: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterasi permissions: %w", err)
	}

	return out, nil
}

func (r *RolePermissionRepositoryPostgres) CreateRolePermission(roleID, permissionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		l := strings.ToLower(err.Error())
		if strings.Contains(l, "duplicate key") || strings.Contains(l, "unique") {
			return errors.New("role_permission sudah ada")
		}
		if strings.Contains(l, "violates foreign key constraint") {
			return errors.New("role_id atau permission_id tidak valid")
		}
		return fmt.Errorf("gagal create role_permission: %w", err)
	}
	return nil
}

func (r *RolePermissionRepositoryPostgres) UpdateRolePermission(oldRoleID, oldPermissionID, newRoleID, newPermissionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE role_permissions
		SET role_id = $1, permission_id = $2
		WHERE role_id = $3 AND permission_id = $4
	`

	res, err := r.db.ExecContext(ctx, query, newRoleID, newPermissionID, oldRoleID, oldPermissionID)
	if err != nil {
		l := strings.ToLower(err.Error())
		if strings.Contains(l, "duplicate key") || strings.Contains(l, "unique") {
			return errors.New("role_permission sudah ada")
		}
		if strings.Contains(l, "violates foreign key constraint") {
			return errors.New("role_id atau permission_id tidak valid")
		}
		return fmt.Errorf("gagal update role_permission: %w", err)
	}

	ra, _ := res.RowsAffected()
	if ra == 0 {
		return errors.New("role_permission tidak ditemukan")
	}
	return nil
}

func (r *RolePermissionRepositoryPostgres) DeleteRolePermission(roleID, permissionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.db.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("gagal delete role_permission: %w", err)
	}

	ra, _ := res.RowsAffected()
	if ra == 0 {
		return errors.New("role_permission tidak ditemukan")
	}
	return nil
}