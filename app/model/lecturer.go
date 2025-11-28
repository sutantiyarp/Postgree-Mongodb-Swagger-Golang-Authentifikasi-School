package model

import (
	"time"
	"github.com/google/uuid"
)

type Lecturer struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	LecturerID string    `db:"lecturer_id" json:"lecturer_id"`
	Department string    `db:"department" json:"department"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type CreateLecturerRequest struct {
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	LecturerID string    `json:"lecturer_id" validate:"required,min=5"`
	Department string    `json:"department" validate:"required"`
}

type UpdateLecturerRequest struct {
	LecturerID *string `json:"lecturer_id"`
	Department *string `json:"department"`
}
