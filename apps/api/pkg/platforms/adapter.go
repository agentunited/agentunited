package platforms

import "context"

type PlatformAdapter interface {
	Authenticate(ctx context.Context, creds Credentials) (*AuthResult, error)
	GetCapabilities() Capabilities
	SendMessage(ctx context.Context, channel string, msg NormalizedMessage) (*DeliveryResult, error)
	ReceiveMessage(ctx context.Context, payload []byte) (*NormalizedMessage, error)
	GetStatus(ctx context.Context) (*PlatformStatus, error)
	ValidateConfig(config map[string]interface{}) (*ValidationResult, error)
}
