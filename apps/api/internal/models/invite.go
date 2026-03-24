package models

import "time"

// InviteStatus represents the status of an invite
type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusConsumed InviteStatus = "consumed"
	InviteStatusExpired  InviteStatus = "expired"
)

// Invite represents a user invitation
type Invite struct {
	ID         string       `json:"id"`
	UserID     string       `json:"user_id"`
	Status     InviteStatus `json:"status"`
	ExpiresAt  time.Time    `json:"expires_at"`
	CreatedAt  time.Time    `json:"created_at"`
	ConsumedAt *time.Time   `json:"consumed_at,omitempty"`
}

// InviteWithToken includes the one-time plaintext token
type InviteWithToken struct {
	Invite
	PlaintextToken string `json:"plaintext_token"`
}

// InviteAcceptRequest represents invite acceptance payload (central identity flow)
type InviteAcceptRequest struct {
	InviteToken string `json:"invite_token" validate:"required"`
	CentralJWT  string `json:"central_jwt" validate:"required"`
}

// InviteCreateRequest represents creating a new invite for a human user.
type InviteCreateRequest struct {
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}
