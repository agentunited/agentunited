package platforms

import (
	"context"
)

// PlatformAdapter is the generic interface for all platform integrations
type PlatformAdapter interface {
	// Identity
	Platform() string // e.g., "openclaw", "discord", "slack"
	Version() string

	// Lifecycle
	Initialize(ctx context.Context, config map[string]interface{}) error
	Shutdown(ctx context.Context) error

	// Authentication
	Authenticate(ctx context.Context, creds Credentials) (*AuthResult, error)

	// Capabilities
	GetCapabilities() Capabilities

	// Messaging
	SendMessage(ctx context.Context, channelID string, msg NormalizedMessage) (*DeliveryResult, error)
	ReceiveMessage(ctx context.Context, payload []byte) (*NormalizedMessage, error)

	// Status
	GetStatus(ctx context.Context) (*PlatformStatus, error)
	CheckHealth(ctx context.Context) (*HealthStatus, error)

	// Validation
	ValidateConfig(config map[string]interface{}) (*ValidationResult, error)
}
