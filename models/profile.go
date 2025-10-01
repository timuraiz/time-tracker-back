package models

import (
	"time"

	"github.com/google/uuid"
)

// Profile represents the user profile table in public schema
type Profile struct {
	ID                uuid.UUID `json:"id" gorm:"primaryKey;column:id"`
	Name              string    `json:"name" gorm:"column:name"`
	ProfilePictureURL *string   `json:"profile_picture_url,omitempty" gorm:"column:profile_picture_url"`
	CreatedAt         time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// CreateProfileRequest represents the request body for creating a profile
type CreateProfileRequest struct {
	Name string `json:"name" binding:"required"`
}

// TableName specifies the table name in the public schema
func (Profile) TableName() string {
	return "public.profiles"
}