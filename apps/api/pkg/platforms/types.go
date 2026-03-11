package platforms

import "time"

// Credentials holds platform-specific authentication data
type Credentials struct {
	APIKey     string            `json:"api_key"`
	Token      string            `json:"token"`
	Secret     string            `json:"secret"`
	Properties map[string]string `json:"properties"`
}

// AuthResult is returned after successful authentication
type AuthResult struct {
	Success   bool      `json:"success"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// Capabilities describes what a platform supports
type Capabilities struct {
	SupportsChannels   bool  `json:"supports_channels"`
	SupportsDMs        bool  `json:"supports_dms"`
	SupportsReactions  bool  `json:"supports_reactions"`
	SupportsFiles      bool  `json:"supports_files"`
	SupportsThreading  bool  `json:"supports_threading"`
	SupportsPresence   bool  `json:"supports_presence"`
	SupportsHistory    bool  `json:"supports_history"`
	MaxFileSize        int64 `json:"max_file_size"`
	RateLimitPerMinute int   `json:"rate_limit_per_minute"`
}

// NormalizedMessage is the universal message format across platforms
type NormalizedMessage struct {
	ID          string            `json:"id"`
	ChannelID   string            `json:"channel_id"`
	ThreadID    string            `json:"thread_id,omitempty"`
	UserID      string            `json:"user_id"`
	Text        string            `json:"text"`
	Timestamp   time.Time         `json:"timestamp"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ReplyTo     string            `json:"reply_to,omitempty"`
}

// Attachment represents a file or media attachment
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
}

// DeliveryResult is returned after sending a message
type DeliveryResult struct {
	Success   bool      `json:"success"`
	MessageID string    `json:"message_id,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// PlatformStatus describes the health of a platform connection
type PlatformStatus struct {
	Connected   bool      `json:"connected"`
	Healthy     bool      `json:"healthy"`
	LastPing    time.Time `json:"last_ping,omitempty"`
	LatencyMs   int       `json:"latency_ms,omitempty"`
	RateLimited bool      `json:"rate_limited"`
	Error       string    `json:"error,omitempty"`
}

// HealthStatus is an alias for backward compatibility
type HealthStatus = PlatformStatus

// ValidationResult is returned after validating platform config
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []string          `json:"errors,omitempty"`
	Warn   []string          `json:"warnings,omitempty"`
}
