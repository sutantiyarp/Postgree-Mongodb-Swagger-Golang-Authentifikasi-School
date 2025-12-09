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

type LecturerRepository interface {
	GetAllLecturers(page, limit int64) ([]model.Lecturer, int64, error)
	GetLecturerByID(id string) (*model.Lecturer, error)
	GetLecturerByUserID(userID string) (*model.Lecturer, error)
	CreateLecturer(req model.CreateLecturerRequest) (string, error)
	UpdateLecturer(id string, req model.UpdateLecturerRequest) error
	DeleteLecturer(id string) error
}

type LecturerRepositoryPostgres struct {
	db *sql.DB
}

func NewLecturerRepositoryPostgres(db *sql.DB) *LecturerRepositoryPostgres {
	return &LecturerRepositoryPostgres{db: db}
}

func (r *LecturerRepositoryPostgres) GetAllLecturers(page, limit int64) ([]model.Lecturer, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM lecturers`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal count lecturers: %w", err)
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal query lecturers: %w", err)
	}
	defer rows.Close()

	var lecturers []model.Lecturer
	for rows.Next() {
		var l model.Lecturer
		if err := rows.Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan lecturer: %w", err)
		}
		lecturers = append(lecturers, l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterasi lecturers: %w", err)
	}

	return lecturers, total, nil
}

func (r *LecturerRepositoryPostgres) GetLecturerByID(id string) (*model.Lecturer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE id = $1
	`
	var l model.Lecturer
	err := r.db.QueryRowContext(ctx, query, id).Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("lecturer tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal get lecturer by id: %w", err)
	}

	return &l, nil
}

func (r *LecturerRepositoryPostgres) GetLecturerByUserID(userID string) (*model.Lecturer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE user_id = $1
		LIMIT 1
	`

	var l model.Lecturer
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("lecturer tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal get lecturer by user_id: %w", err)
	}
	return &l, nil
}

func (r *LecturerRepositoryPostgres) CreateLecturer(req model.CreateLecturerRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO lecturers (user_id, lecturer_id, department)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query, req.UserID, strings.TrimSpace(req.LecturerID), strings.TrimSpace(req.Department)).Scan(&id)
	if err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "duplicate key") || strings.Contains(lowerErr, "unique") {
			return "", errors.New("lecturer_id sudah digunakan")
		}
		if strings.Contains(lowerErr, "foreign key") {
			return "", errors.New("user_id tidak valid")
		}
		return "", fmt.Errorf("gagal membuat lecturer: %w", err)
	}

	return id, nil
}

func (r *LecturerRepositoryPostgres) UpdateLecturer(id string, req model.UpdateLecturerRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var updates []string
	var args []interface{}
	argIndex := 1

	if req.LecturerID != nil {
		v := strings.TrimSpace(*req.LecturerID)
		if v == "" {
			return errors.New("lecturer_id tidak boleh kosong")
		}
		updates = append(updates, fmt.Sprintf("lecturer_id = $%d", argIndex))
		args = append(args, v)
		argIndex++
	}

	if req.Department != nil {
		v := strings.TrimSpace(*req.Department)
		if v == "" {
			return errors.New("department tidak boleh kosong")
		}
		updates = append(updates, fmt.Sprintf("department = $%d", argIndex))
		args = append(args, v)
		argIndex++
	}

	if len(updates) == 0 {
		return errors.New("tidak ada field yang diupdate")
	}

	query := fmt.Sprintf(`
		UPDATE lecturers
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argIndex)

	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "duplicate key") || strings.Contains(lowerErr, "unique") {
			return errors.New("lecturer_id sudah digunakan")
		}
		return fmt.Errorf("gagal update lecturer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("lecturer tidak ditemukan")
	}

	return nil
}

func (r *LecturerRepositoryPostgres) DeleteLecturer(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `DELETE FROM lecturers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus lecturer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("lecturer tidak ditemukan")
	}

	return nil
}
