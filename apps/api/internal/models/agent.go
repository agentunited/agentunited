package models

import "time"

// Agent represents an AI agent identity
type Agent struct {
	ID          string                 `json:"id"`
	OwnerID     string                 `json:"owner_id"`
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name,omitempty"`
	Description string                 `json:"description,omitempty"`
	AvatarURL   string                 `json:"avatar_url,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CreateAgentRequest represents agent creation payload
type CreateAgentRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=100"`
	DisplayName string                 `json:"display_name" validate:"max=255"`
	Description string                 `json:"description"`
	AvatarURL   string                 `json:"avatar_url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateAgentRequest represents agent update payload
type UpdateAgentRequest struct {
	DisplayName *string                 `json:"display_name"`
	Description *string                 `json:"description"`
	AvatarURL   *string                 `json:"avatar_url"`
	Metadata    *map[string]interface{} `json:"metadata"`
}
