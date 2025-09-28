package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type TimeEntry struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	ProjectName string         `json:"project_name" gorm:"not null"`
	Description string         `json:"description"`
	StartTime   time.Time      `json:"start_time" gorm:"not null"`
	EndTime     *time.Time     `json:"end_time"`
	Duration    int64          `json:"duration"` // in seconds
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Note: User relation points to auth.users table managed by Supabase
	// We skip foreign key constraints since we don't have permission to modify auth.users
	User User `json:"user,omitempty" gorm:"-:migration;foreignKey:UserID;references:ID"`
}

type TimeEntryCreateRequest struct {
	ProjectName string `json:"project_name" binding:"required"`
	Description string `json:"description"`
}

type TimeEntryUpdateRequest struct {
	ProjectName string `json:"project_name"`
	Description string `json:"description"`
	EndTime     string `json:"end_time"` // ISO format
}

type TimeEntryResponse struct {
	ID          uuid.UUID  `json:"id"`
	ProjectName string     `json:"project_name"`
	Description string     `json:"description"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int64      `json:"duration"`
	CreatedAt   time.Time  `json:"created_at"`
}

type PaginatedTimeEntriesResponse struct {
	Data       []TimeEntryResponse `json:"data"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	Total      int64               `json:"total"`
	TotalPages int                 `json:"total_pages"`
}
