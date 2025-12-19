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
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AchievementMongoRepository interface {
	Create(ctx context.Context, studentID uuid.UUID, req model.CreateAchievementRequest) (string, error)
	GetByIDs(ctx context.Context, ids []string) ([]model.Achievement, error)
	List(ctx context.Context, page, limit int64) ([]model.Achievement, int64, error)
	Delete(ctx context.Context, id string) error
}

type AchievementReferenceRepository interface {
	CreateDraft(ctx context.Context, studentID uuid.UUID, mongoID string) (string, error)
	SubmitDraft(ctx context.Context, refID string, studentID uuid.UUID) error
	Review(ctx context.Context, refID string, status string, adminID uuid.UUID, note *string) error
	Delete(ctx context.Context, refID string, adminID uuid.UUID) error
	DeleteByStudent(ctx context.Context, refID string, studentID uuid.UUID) error
	HardDelete(ctx context.Context, refID string) error
	GetByID(ctx context.Context, id string) (*model.AchievementReference, error)
	List(ctx context.Context, page, limit int64) ([]model.AchievementReference, int64, error)
	ListByStatuses(ctx context.Context, statuses []string, studentID *uuid.UUID, advisorID *uuid.UUID, page, limit int64) ([]model.AchievementReference, int64, error)
}

type achievementMongoRepository struct {
	col *mongo.Collection
}

func NewAchievementMongoRepository(db *mongo.Database) AchievementMongoRepository {
	return &achievementMongoRepository{
		col: db.Collection("achievements"),
	}
}

func (r *achievementMongoRepository) Create(ctx context.Context, studentID uuid.UUID, req model.CreateAchievementRequest) (string, error) {
	now := time.Now()

	doc := model.Achievement{
		ID:              bson.NewObjectID(),
		StudentID:       studentID.String(),
		AchievementType: strings.ToLower(strings.TrimSpace(req.AchievementType)),
		Title:           strings.TrimSpace(req.Title),
		Description:     strings.TrimSpace(req.Description),
		Details:         req.Details,
		Attachments:     req.Attachments,
		Tags:            req.Tags,
		Points:          req.Points,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	res, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan achievement ke MongoDB: %w", err)
	}

	oid, ok := res.InsertedID.(bson.ObjectID)
	if !ok {
		return "", errors.New("InsertOne tidak mengembalikan ObjectID")
	}

	return oid.Hex(), nil
}

func (r *achievementMongoRepository) List(ctx context.Context, page, limit int64) ([]model.Achievement, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.col.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil achievements: %w", err)
	}
	defer cursor.Close(ctx)

	var list []model.Achievement
	if err := cursor.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("gagal decode achievements: %w", err)
	}

	total, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return list, 0, fmt.Errorf("gagal menghitung total achievements: %w", err)
	}

	return list, total, nil
}

func (r *achievementMongoRepository) GetByIDs(ctx context.Context, ids []string) ([]model.Achievement, error) {
	if len(ids) == 0 {
		return []model.Achievement{}, nil
	}
	var objectIDs []bson.ObjectID
	for _, id := range ids {
		if oid, err := bson.ObjectIDFromHex(id); err == nil {
			objectIDs = append(objectIDs, oid)
		}
	}
	if len(objectIDs) == 0 {
		return []model.Achievement{}, nil
	}

	cursor, err := r.col.Find(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil achievements by ids: %w", err)
	}
	defer cursor.Close(ctx)

	var list []model.Achievement
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("gagal decode achievements by ids: %w", err)
	}
	return list, nil
}

func (r *achievementMongoRepository) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid mongo achievement id: %w", err)
	}
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("gagal menghapus achievement mongo: %w", err)
	}
	if res.DeletedCount == 0 {
		return errors.New("achievement mongo tidak ditemukan")
	}
	return nil
}

type achievementReferenceRepository struct {
	db *sql.DB
}

func NewAchievementReferenceRepository(db *sql.DB) AchievementReferenceRepository {
	return &achievementReferenceRepository{db: db}
}

func (r *achievementReferenceRepository) CreateDraft(ctx context.Context, studentID uuid.UUID, mongoID string) (string, error) {
	query := `
		INSERT INTO achievement_references (student_id, mongo_achievement_id, status)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id string
	err := r.db.QueryRowContext(ctx, query, studentID, mongoID, model.AchievementStatusDraft).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("gagal membuat draft achievement reference: %w", err)
	}
	return id, nil
}

func (r *achievementReferenceRepository) SubmitDraft(ctx context.Context, refID string, studentID uuid.UUID) error {
	query := `
		UPDATE achievement_references
		SET status = $1,
			submitted_at = NOW(),
			updated_at = NOW()
		WHERE id = $2
		  AND student_id = $3
		  AND status = $4
	`
	result, err := r.db.ExecContext(ctx, query, model.AchievementStatusSubmitted, refID, studentID, model.AchievementStatusDraft)
	if err != nil {
		return fmt.Errorf("gagal submit achievement: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected submit: %w", err)
	}
	if affected == 0 {
		return errors.New("achievement tidak ditemukan atau bukan milik anda atau status bukan draft")
	}
	return nil
}

func (r *achievementReferenceRepository) Review(ctx context.Context, refID string, status string, adminID uuid.UUID, note *string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != model.AchievementStatusVerified &&
		status != model.AchievementStatusRejected {
		return errors.New("status review tidak valid")
	}

	var rejectionNote interface{}
	if status == model.AchievementStatusRejected && note != nil {
		rejectionNote = strings.TrimSpace(*note)
		if rejectionNote == "" {
			rejectionNote = nil
		}
	} else {
		rejectionNote = nil
	}

	query := `
		UPDATE achievement_references
		SET status = $1,
			verified_at = NOW(),
			verified_by = $2,
			rejection_note = $3,
			updated_at = NOW()
		WHERE id = $4
		  AND status = $5
	`
	result, err := r.db.ExecContext(ctx, query, status, adminID, rejectionNote, refID, model.AchievementStatusSubmitted)
	if err != nil {
		return fmt.Errorf("gagal review achievement: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected review: %w", err)
	}
	if affected == 0 {
		return errors.New("achievement tidak ditemukan atau status bukan submitted")
	}
	return nil
}

func (r *achievementReferenceRepository) Delete(ctx context.Context, refID string, adminID uuid.UUID) error {
	query := `
		UPDATE achievement_references
		SET status = $1,
			verified_at = NOW(),
			verified_by = $2,
			rejection_note = NULL,
			updated_at = NOW()
		WHERE id = $3
		  AND status != $1
	`
	result, err := r.db.ExecContext(ctx, query, model.AchievementStatusDeleted, adminID, refID)
	if err != nil {
		return fmt.Errorf("gagal menghapus (soft delete) achievement: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected delete: %w", err)
	}
	if affected == 0 {
		return errors.New("achievement tidak ditemukan atau sudah berstatus deleted")
	}
	return nil
}

func (r *achievementReferenceRepository) DeleteByStudent(ctx context.Context, refID string, studentID uuid.UUID) error {
	query := `
		UPDATE achievement_references
		SET status = $1,
			verified_at = NOW(),
			verified_by = NULL,
			rejection_note = NULL,
			updated_at = NOW()
		WHERE id = $2
		  AND student_id = $3
		  AND status = $4
	`
	result, err := r.db.ExecContext(ctx, query, model.AchievementStatusDeleted, refID, studentID, model.AchievementStatusDraft)
	if err != nil {
		return fmt.Errorf("gagal menghapus (soft delete) achievement oleh student: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected delete student: %w", err)
	}
	if affected == 0 {
		return errors.New("achievement tidak ditemukan atau bukan milik anda atau status bukan draft")
	}
	return nil
}

func (r *achievementReferenceRepository) HardDelete(ctx context.Context, refID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM achievement_references WHERE id = $1 AND status = $2`, refID, model.AchievementStatusDeleted)
	if err != nil {
		return fmt.Errorf("gagal hard delete achievement_reference: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected hard delete: %w", err)
	}
	if rows == 0 {
		return errors.New("achievement reference tidak ditemukan atau status bukan deleted")
	}
	return nil
}

func (r *achievementReferenceRepository) GetByID(ctx context.Context, id string) (*model.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE id = $1
	`
	var ref model.AchievementReference
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("achievement reference tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil achievement reference: %w", err)
	}
	return &ref, nil
}

func (r *achievementReferenceRepository) List(ctx context.Context, page, limit int64) ([]model.AchievementReference, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM achievement_references`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total achievement_references: %w", err)
	}

	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil achievement_references: %w", err)
	}
	defer rows.Close()

	var refs []model.AchievementReference
	for rows.Next() {
		var ref model.AchievementReference
		if err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.VerifiedAt,
			&ref.VerifiedBy,
			&ref.RejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("gagal scan achievement_reference: %w", err)
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterasi achievement_references: %w", err)
	}

	return refs, total, nil
}

func (r *achievementReferenceRepository) ListByStatuses(ctx context.Context, statuses []string, studentID *uuid.UUID, advisorID *uuid.UUID, page, limit int64) ([]model.AchievementReference, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	args := []interface{}{}
	placeholders := []string{}
	for i, s := range statuses {
		args = append(args, s)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}
	statusArray := fmt.Sprintf("ARRAY[%s]", strings.Join(placeholders, ","))

	where := fmt.Sprintf("ar.status = ANY(%s)", statusArray)
	join := ""
	if studentID != nil {
		args = append(args, *studentID)
		where += fmt.Sprintf(" AND ar.student_id = $%d", len(args))
	}
	if advisorID != nil {
		join = " JOIN students s ON ar.student_id = s.id"
		args = append(args, *advisorID)
		where += fmt.Sprintf(" AND s.advisor_id = $%d", len(args))
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM achievement_references ar%s WHERE %s`, join, where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total achievement_references: %w", err)
	}

	args = append(args, limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at
		FROM achievement_references ar%s
		WHERE %s
		ORDER BY ar.created_at DESC
		LIMIT $%d OFFSET $%d
	`, join, where, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil achievement_references: %w", err)
	}
	defer rows.Close()

	var refs []model.AchievementReference
	for rows.Next() {
		var ref model.AchievementReference
		if err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.VerifiedAt,
			&ref.VerifiedBy,
			&ref.RejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("gagal scan achievement_reference: %w", err)
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterasi achievement_references: %w", err)
	}

	return refs, total, nil
}
