package models

import "time"

// User represents a user in the system
type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	DisplayName    string    `json:"display_name,omitempty"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	UserType       string    `json:"user_type,omitempty"`
	PasswordHash   string    `json:"-"` // Never serialize password hash
	AuthType       string    `json:"auth_type,omitempty"`       // "local" or "central"
	CentralUserID  string    `json:"central_user_id,omitempty"` // sub from central JWT
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
