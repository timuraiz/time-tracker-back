package models

import (
	"github.com/google/uuid"
)

// User represents the Supabase auth.users table structure
// This is read-only since Supabase manages authentication
type User struct {
	ID    uuid.UUID `json:"id" gorm:"primaryKey;column:id"`
	Name  string    `json:"name" gorm:"column:name"`
	Email string    `json:"email" gorm:"column:email"`
}

// TableName specifies the table name in the auth schema
func (User) TableName() string {
	return "auth.users"
}

type UserResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	ProfilePictureURL *string   `json:"profile_picture_url,omitempty"`
	TotalHours        float32   `json:"total_hours"`
	TotalSessions     int       `json:"total_sessions"`
	CurrentStreak     int       `json:"current_streak"`
	DayilyAvg         float32   `json:"dayily_avg"`
	Rank              int       `json:"rank"`
	Level             string    `json:"level"`
	LevelColor        string    `json:"level_color"`
	CreatedAt         string    `json:"created_at"`
}

