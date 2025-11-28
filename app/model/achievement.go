package model

import (
	"time"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AchievementReference struct {
	ID                 uuid.UUID `db:"id" json:"id"`
	StudentID          uuid.UUID `db:"student_id" json:"student_id"`
	MongoAchievementID string    `db:"mongo_achievement_id" json:"mongo_achievement_id"`
	Status             string    `db:"status" json:"status"` // draft, submitted, verified, rejected
	SubmittedAt        *time.Time `db:"submitted_at" json:"submitted_at"`
	VerifiedAt         *time.Time `db:"verified_at" json:"verified_at"`
	VerifiedBy         *uuid.UUID `db:"verified_by" json:"verified_by"`
	RejectionNote      *string    `db:"rejection_note" json:"rejection_note"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

// MongoDB Achievement Document
type Achievement struct {
	ID               bson.ObjectID         `bson:"_id,omitempty" json:"id,omitempty"`
	StudentID        uuid.UUID             `bson:"student_id" json:"student_id"`
	AchievementType  string                `bson:"achievement_type" json:"achievement_type"` // academic, competition, organization, publication, certification, other
	Title            string                `bson:"title" json:"title"`
	Description      string                `bson:"description" json:"description"`
	Details          bson.M                `bson:"details" json:"details"`
	Attachments      []Attachment          `bson:"attachments" json:"attachments"`
	Tags             []string              `bson:"tags" json:"tags"`
	Points           *int                  `bson:"points" json:"points"`
	CreatedAt        time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time             `bson:"updated_at" json:"updated_at"`
}

type Attachment struct {
	FileName   string    `bson:"file_name" json:"file_name"`
	FileURL    string    `bson:"file_url" json:"file_url"`
	FileType   string    `bson:"file_type" json:"file_type"`
	UploadedAt time.Time `bson:"uploaded_at" json:"uploaded_at"`
}

type CreateAchievementRequest struct {
	AchievementType string                 `json:"achievement_type" validate:"required,oneof=academic competition organization publication certification other"`
	Title           string                 `json:"title" validate:"required"`
	Description     string                 `json:"description" validate:"required"`
	Details         bson.M                 `json:"details" validate:"required"`
	Tags            []string               `json:"tags"`
	Points          *int                   `json:"points"`
}

type SubmitAchievementRequest struct {
	// Empty, hanya trigger submit
}

type VerifyAchievementRequest struct {
	Approved bool   `json:"approved" validate:"required"`
	Note     string `json:"note"`
}

type AchievementWithReference struct {
	Achievement Achievement           `json:"achievement"`
	Reference   AchievementReference  `json:"reference"`
}

type AchievementStatistics struct {
	TotalByType      map[string]int `json:"total_by_type"`
	TotalByPeriod    map[string]int `json:"total_by_period"`
	TopStudents      []TopStudent   `json:"top_students"`
	CompetitionLevel map[string]int `json:"competition_level"`
}

type TopStudent struct {
	StudentID    uuid.UUID `json:"student_id"`
	StudentName  string    `json:"student_name"`
	TotalAchievements int   `json:"total_achievements"`
	TotalPoints  int       `json:"total_points"`
}
