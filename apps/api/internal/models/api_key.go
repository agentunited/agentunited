package models

import "time"

// APIKey represents an agent API key
type APIKey struct {
	ID         string     `json:"id"`
	AgentID    string     `json:"agent_id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// APIKeyWithPlaintext includes the one-time plaintext key
type APIKeyWithPlaintext struct {
	APIKey
	PlaintextKey string `json:"plaintext_key"`
}

// CreateAPIKeyRequest represents API key creation
type CreateAPIKeyRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
