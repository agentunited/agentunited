package models

import "time"

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	PasswordHash string    `json:"-"` // Never serialize password hash
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
