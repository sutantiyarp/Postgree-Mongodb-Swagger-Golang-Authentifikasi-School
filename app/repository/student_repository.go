package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"hello-fiber/app/model"

	"github.com/google/uuid"
)

type StudentRepository interface {
	GetAllStudents(page, limit int64) ([]model.Student, int64, error)
	GetStudentByID(id string) (*model.Student, error)
	CreateStudent(req model.CreateStudentRequest) (string, error)
	UpdateStudent(id string, req model.UpdateStudentRequest) error
	DeleteStudent(id string) error
}

type StudentRepositoryPostgres struct {
	db *sql.DB
}

func NewStudentRepositoryPostgres(db *sql.DB) *StudentRepositoryPostgres {
	return &StudentRepositoryPostgres{db: db}
}

func (r *StudentRepositoryPostgres) GetAllStudents(page, limit int64) ([]model.Student, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM students`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal count students: %w", err)
	}

	offset := (page - 1) * limit
	query := `
		SELECT
			id,
			user_id,
			student_id,
			COALESCE(program_study, ''),
			COALESCE(academic_year, ''),
			advisor_id::text,
			created_at
		FROM students
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal query students: %w", err)
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		var advisorStr sql.NullString

		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.StudentID,
			&s.ProgramStudy,
			&s.AcademicYear,
			&advisorStr,
			&s.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("gagal scan student: %w", err)
		}

		s.AdvisorID = nil
		if advisorStr.Valid && strings.TrimSpace(advisorStr.String) != "" {
			aid, err := uuid.Parse(strings.TrimSpace(advisorStr.String))
			if err == nil {
				tmp := aid
				s.AdvisorID = &tmp
			}
		}

		students = append(students, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterasi students: %w", err)
	}

	return students, total, nil
}

func (r *StudentRepositoryPostgres) GetStudentByID(id string) (*model.Student, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT
			id,
			user_id,
			student_id,
			COALESCE(program_study, ''),
			COALESCE(academic_year, ''),
			advisor_id::text,
			created_at
		FROM students
		WHERE id = $1
	`

	var s model.Student
	var advisorStr sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID,
		&s.UserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&advisorStr,
		&s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal get student by id: %w", err)
	}

	s.AdvisorID = nil
	if advisorStr.Valid && strings.TrimSpace(advisorStr.String) != "" {
		aid, err := uuid.Parse(strings.TrimSpace(advisorStr.String))
		if err == nil {
			tmp := aid
			s.AdvisorID = &tmp
		}
	}

	return &s, nil
}

func (r *StudentRepositoryPostgres) CreateStudent(req model.CreateStudentRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req.StudentID = strings.TrimSpace(req.StudentID)
	req.ProgramStudy = strings.TrimSpace(req.ProgramStudy)
	req.AcademicYear = strings.TrimSpace(req.AcademicYear)

	// advisor nil/uuid.Nil => NULL
	var advisorArg interface{}
	if req.AdvisorID != nil && *req.AdvisorID != uuid.Nil {
		advisorArg = *req.AdvisorID
	} else {
		advisorArg = nil
	}

	// program/year boleh kosong => NULL (biar rapi)
	var progArg interface{} = nil
	if req.ProgramStudy != "" {
		progArg = req.ProgramStudy
	}
	var yearArg interface{} = nil
	if req.AcademicYear != "" {
		yearArg = req.AcademicYear
	}

	query := `
		INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query, req.UserID, req.StudentID, progArg, yearArg, advisorArg).Scan(&id)
	if err != nil {
		l := strings.ToLower(err.Error())

		if strings.Contains(l, "duplicate key") || strings.Contains(l, "unique") || strings.Contains(l, "student_id_key") {
			return "", errors.New("student_id sudah digunakan")
		}
		if strings.Contains(l, "students_user_id_fkey") {
			return "", errors.New("user_id tidak valid")
		}
		if strings.Contains(l, "students_advisor_id_fkey") {
			return "", errors.New("advisor_id tidak valid")
		}
		if strings.Contains(l, "foreign key") {
			return "", errors.New("user_id atau advisor_id tidak valid")
		}

		return "", fmt.Errorf("gagal membuat student: %w", err)
	}

	return id, nil
}

func (r *StudentRepositoryPostgres) UpdateStudent(id string, req model.UpdateStudentRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var updates []string
	var args []interface{}
	argIndex := 1

	if req.StudentID != nil {
		v := strings.TrimSpace(*req.StudentID)
		if v == "" {
			return errors.New("student_id tidak boleh kosong")
		}
		updates = append(updates, fmt.Sprintf("student_id = $%d", argIndex))
		args = append(args, v)
		argIndex++
	}

	if req.ProgramStudy != nil {
		v := strings.TrimSpace(*req.ProgramStudy)
		if v == "" {
			return errors.New("program_study tidak boleh kosong")
		}
		updates = append(updates, fmt.Sprintf("program_study = $%d", argIndex))
		args = append(args, v)
		argIndex++
	}

	if req.AcademicYear != nil {
		v := strings.TrimSpace(*req.AcademicYear)
		if v == "" {
			return errors.New("academic_year tidak boleh kosong")
		}
		updates = append(updates, fmt.Sprintf("academic_year = $%d", argIndex))
		args = append(args, v)
		argIndex++
	}

	if req.AdvisorID != nil {
		var advArg interface{}
		if *req.AdvisorID == uuid.Nil {
			advArg = nil // set NULL
		} else {
			advArg = *req.AdvisorID
		}

		updates = append(updates, fmt.Sprintf("advisor_id = $%d", argIndex))
		args = append(args, advArg)
		argIndex++
	}

	if len(updates) == 0 {
		return errors.New("tidak ada field yang diupdate")
	}

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE students SET %s WHERE id = $%d`, strings.Join(updates, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		l := strings.ToLower(err.Error())
		if strings.Contains(l, "duplicate key") || strings.Contains(l, "unique") || strings.Contains(l, "student_id_key") {
			return errors.New("student_id sudah digunakan")
		}
		if strings.Contains(l, "students_advisor_id_fkey") {
			return errors.New("advisor_id tidak valid")
		}
		if strings.Contains(l, "foreign key") {
			return errors.New("advisor_id tidak valid")
		}
		return fmt.Errorf("gagal update student: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}
	if affected == 0 {
		return errors.New("student tidak ditemukan")
	}

	return nil
}

func (r *StudentRepositoryPostgres) DeleteStudent(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `DELETE FROM students WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus student: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}
	if affected == 0 {
		return errors.New("student tidak ditemukan")
	}

	return nil
}
