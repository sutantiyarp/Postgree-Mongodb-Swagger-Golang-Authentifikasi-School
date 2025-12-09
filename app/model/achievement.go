package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	AchievementStatusDraft     = "draft"
	AchievementStatusSubmitted = "submitted"
	AchievementStatusVerified  = "verified"
	AchievementStatusRejected  = "rejected"
	AchievementStatusDeleted   = "deleted"
)

type AchievementReference struct {
	ID                 uuid.UUID `db:"id" json:"id"`
	StudentID          uuid.UUID `db:"student_id" json:"student_id"`
	MongoAchievementID string    `db:"mongo_achievement_id" json:"mongo_achievement_id"`
	Status             string    `db:"status" json:"status"` // draft, submitted, verified, rejected, deleted
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
	StudentID        string                `bson:"studentId" json:"student_id"`
	AchievementType  string                `bson:"achievementType" json:"achievement_type"` // academic, competition, organization, publication, certification, other
	Title            string                `bson:"title" json:"title"`
	Description      string                `bson:"description" json:"description"`
	Details          map[string]interface{} `bson:"details" json:"details"`
	Attachments      []Attachment          `bson:"attachments,omitempty" json:"attachments,omitempty"`
	Tags             []string              `bson:"tags,omitempty" json:"tags,omitempty"`
	Points           *float64              `bson:"points,omitempty" json:"points,omitempty"`
	CreatedAt        time.Time             `bson:"createdAt" json:"created_at"`
	UpdatedAt        time.Time             `bson:"updatedAt" json:"updated_at"`
}

type Attachment struct {
	FileName   string    `bson:"fileName" json:"file_name"`
	FileURL    string    `bson:"fileUrl" json:"file_url"`
	FileType   string    `bson:"fileType" json:"file_type"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploaded_at"`
}

type CreateAchievementRequest struct {
	AchievementType string                 `json:"achievement_type" validate:"required,oneof=academic competition organization publication certification other"`
	Title           string                 `json:"title" validate:"required"`
	Description     string                 `json:"description" validate:"required"`
	Details         map[string]interface{} `json:"details" swaggertype:"object" validate:"required"`
	Tags            []string               `json:"tags"`
	Points          *float64               `json:"points"`
}

type SubmitAchievementRequest struct {
	// Empty, hanya trigger submit
}

type VerifyAchievementRequest struct {
	Approved bool   `json:"approved" validate:"required"`
	Note     string `json:"note"`
}

type UpdateAchievementStatusRequest struct {
	Status        string  `json:"status" validate:"required,oneof=verified rejected" example:"verified/rejected"`
	RejectionNote *string `json:"rejection_note" example:"string"`
}

// Detail structs untuk dokumentasi Swagger (oneOf)
type CompetitionDetails struct {
	CompetitionName  string `json:"competitionName" example:"ICPC National"`
	CompetitionLevel string `json:"competitionLevel" example:"national"` // international, national, regional, local
	Rank             int    `json:"rank" example:"1"`
	MedalType        string `json:"medalType" example:"gold"`
	EventDate        string `json:"eventDate,omitempty" example:"2025-11-01"`
	Location         string `json:"location,omitempty" example:"Jakarta"`
	Organizer        string `json:"organizer,omitempty" example:"ACM"`
	Score            string `json:"score,omitempty" example:"95.5"`
}

type PublicationDetails struct {
	PublicationType  string   `json:"publicationType" example:"journal"` // journal, conference, book
	PublicationTitle string   `json:"publicationTitle" example:"Efficient Algorithms"`
	Authors          []string `json:"authors" example:"[\"Alice\",\"Bob\"]"`
	Publisher        string   `json:"publisher" example:"Springer"`
	Issn             string   `json:"issn" example:"1234-5678"`
}

type OrganizationDetails struct {
	OrganizationName string `json:"organizationName" example:"BEM"`
	Position         string `json:"position" example:"Ketua"`
	PeriodStart      string `json:"periodStart" example:"2025-01-01"`
	PeriodEnd        string `json:"periodEnd" example:"2025-12-31"`
}

type CertificationDetails struct {
	CertificationName   string `json:"certificationName" example:"AWS Cloud Practitioner"`
	IssuedBy            string `json:"issuedBy" example:"Amazon"`
	CertificationNumber string `json:"certificationNumber" example:"ABC-123"`
	ValidUntil          string `json:"validUntil" example:"2026-12-31"`
}

type AcademicDetails struct {
	Description string  `json:"description,omitempty" example:"IPK 3.9"`
	Score       float64 `json:"score,omitempty" example:"3.9"`
}

// AchievementDetailsDoc dipakai untuk dokumentasi oneOf di Swagger (referensi manual melalui deskripsi).
type AchievementDetailsDoc struct {
	Competition   CompetitionDetails   `json:"competition,omitempty"`
	Publication   PublicationDetails   `json:"publication,omitempty"`
	Organization  OrganizationDetails  `json:"organization,omitempty"`
	Certification CertificationDetails `json:"certification,omitempty"`
	Academic      AcademicDetails      `json:"academic,omitempty"`
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
