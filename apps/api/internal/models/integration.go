package models

import "time"

type Integration struct {
	ID                 string    `json:"id"`
	WorkspaceID        string    `json:"workspace_id"`
	Name               string    `json:"name"`
	Platform           string    `json:"platform"`
	APIKey             string    `json:"-"`
	WebhookURL         string    `json:"webhook_url"`
	EventSubscriptions []string  `json:"event_subscriptions"`
	Active             bool      `json:"active"`
	CreatedAt          time.Time `json:"created_at"`
}

type CreateIntegrationRequest struct {
	Name               string   `json:"name"`
	Platform           string   `json:"platform"`
	WebhookURL         string   `json:"webhook_url"`
	EventSubscriptions []string `json:"event_subscriptions"`
	Events             []string `json:"events"`
}
