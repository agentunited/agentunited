package models

import "time"

// Message represents a chat message
type Message struct {
	ID             string    `json:"id"`
	ChannelID      string    `json:"channel_id"`
	AuthorID       string    `json:"author_id"`
	AuthorType     string    `json:"author_type"`
	AuthorEmail    string    `json:"author_email,omitempty"` // Display-resolved author label (display_name -> email prefix fallback for users)
	Text           string    `json:"text"`
	AttachmentURL  string    `json:"attachment_url,omitempty"`
	AttachmentName string    `json:"attachment_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MessageList represents a paginated list of messages
type MessageList struct {
	Messages []*Message `json:"messages"`
	HasMore  bool       `json:"has_more"`
}
