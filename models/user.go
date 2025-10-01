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
}
