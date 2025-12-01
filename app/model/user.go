package model

import (
	"time"
)

type User struct {
	ID        string    `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Email     string    `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"password_hash,omitempty"`
	FullName  string    `db:"full_name" json:"full_name"`
	RoleID    string    `db:"role_id" json:"role_id"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Role struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type Permission struct {
	ID          string `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Resource    string `db:"resource" json:"resource"`
	Action      string `db:"action" json:"action"`
	Description string `db:"description" json:"description"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	RoleID    string    `json:"role_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FullName  string `json:"full_name" binding:"required"`
}

type CreateUserRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FullName  string `json:"full_name" binding:"required"`
	IsActive  bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	RoleID   string `json:"role_id"`
	IsActive *bool  `json:"is_active"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}
