package models

import "time"

// Webhook represents a webhook subscription
type Webhook struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           string                 `json:"id"`
	WebhookID    string                 `json:"webhook_id"`
	EventType    string                 `json:"event_type"`
	Payload      map[string]interface{} `json:"payload"`
	Status       string                 `json:"status"` // pending, success, failed
	ResponseCode *int                   `json:"response_code,omitempty"`
	ResponseBody *string                `json:"response_body,omitempty"`
	AttemptCount int                    `json:"attempt_count"`
	CreatedAt    time.Time              `json:"created_at"`
	DeliveredAt  *time.Time             `json:"delivered_at,omitempty"`
}

// CreateWebhookRequest represents webhook creation
type CreateWebhookRequest struct {
	URL    string   `json:"url" validate:"required,url"`
	Events []string `json:"events" validate:"required,min=1"`
}
