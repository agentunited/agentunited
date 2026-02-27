package models

import "time"

// Channel represents a chat channel
type Channel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Topic     string    `json:"topic,omitempty"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChannelMember represents a user's membership in a channel
type ChannelMember struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// ChannelWithMembers includes member information
type ChannelWithMembers struct {
	Channel
	Members      []*MemberInfo `json:"members,omitempty"`
	MemberCount  int           `json:"member_count,omitempty"`
}

// MemberInfo represents minimal user info for channel members
type MemberInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
