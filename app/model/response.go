package model

import (
	"time"

	"github.com/google/uuid"
)

type MetaInfo struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Total  int    `json:"total"`
	Pages  int    `json:"pages"`
	SortBy string `json:"sortBy"`
	Order  string `json:"order"`
	Search string `json:"search"`
}

type LoginResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Token   string        `json:"token,omitempty"`
	User    *UserResponse `json:"user,omitempty"`
}

type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation successful"`
	ID      string `json:"id,omitempty" example:"507f1f77bcf86cd799439011"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Error message"`
	Error   string `json:"error,omitempty" example:"Detailed error"`
}

type UserListResponse struct {
	Success bool                  `json:"success" example:"true"`
	Message string                `json:"message" example:"Data user berhasil diambil"`
	Data    []UserResponse `json:"data"`
	Total   int64                 `json:"total"`
	Page    int64                 `json:"page"`
	Limit   int64                 `json:"limit"`
}

type RoleListResponse struct {
	Success bool     `json:"success" example:"true"`
	Message string   `json:"message" example:"Data role berhasil diambil"`
	Data    []Role   `json:"data"`
	Total   int64    `json:"total"`
	Page    int64    `json:"page"`
	Limit   int64    `json:"limit"`
}

type RoleDetailResponse struct {
	Success bool  `json:"success" example:"true"`
	Message string `json:"message" example:"Data role berhasil diambil"`
	Data    *Role `json:"data,omitempty"`
}

type StudentResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	StudentID   string    `json:"student_id"`
	ProgramStudy string   `json:"program_study"`
	AcademicYear string   `json:"academic_year"`
	AdvisorID   *uuid.UUID `json:"advisor_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
}

type LecturerResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	LecturerID string    `json:"lecturer_id"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserDetailResponse struct {
	Success bool          `json:"success" example:"true"`
	Message string        `json:"message" example:"Data user berhasil diambil"`
	Data    *UserResponse `json:"data,omitempty"`
}
