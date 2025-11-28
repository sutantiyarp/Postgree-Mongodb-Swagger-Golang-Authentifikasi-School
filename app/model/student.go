package model

import (
	"time"
	"github.com/google/uuid"
)

type Student struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	StudentID   string     `db:"student_id" json:"student_id"`
	ProgramStudy string    `db:"program_study" json:"program_study"`
	AcademicYear string    `db:"academic_year" json:"academic_year"`
	AdvisorID   *uuid.UUID `db:"advisor_id" json:"advisor_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

type CreateStudentRequest struct {
	UserID      uuid.UUID  `json:"user_id" validate:"required"`
	StudentID   string     `json:"student_id" validate:"required,min=5"`
	ProgramStudy string    `json:"program_study" validate:"required"`
	AcademicYear string    `json:"academic_year" validate:"required"`
	AdvisorID   *uuid.UUID `json:"advisor_id"`
}

type UpdateStudentRequest struct {
	StudentID   *string    `json:"student_id"`
	ProgramStudy *string   `json:"program_study"`
	AcademicYear *string   `json:"academic_year"`
	AdvisorID   *uuid.UUID `json:"advisor_id"`
}
