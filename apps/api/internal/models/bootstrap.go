package models

// BootstrapRequest represents the complete bootstrap payload
type BootstrapRequest struct {
	PrimaryAgent   BootstrapPrimaryAgent `json:"primary_agent" validate:"required"`
	Agents         []BootstrapAgent      `json:"agents"`
	Humans         []BootstrapHuman      `json:"humans"`
	DefaultChannel BootstrapChannel      `json:"default_channel"`
}

// BootstrapPrimaryAgent represents the admin agent user
type BootstrapPrimaryAgent struct {
	Email        string                `json:"email" validate:"required,email"`
	Password     string                `json:"password" validate:"required,min=12"`
	AgentProfile BootstrapAgentProfile `json:"agent_profile" validate:"required"`
}

// BootstrapAgent represents an additional agent to create
type BootstrapAgent struct {
	Name        string                 `json:"name" validate:"required,min=3,max=100"`
	DisplayName string                 `json:"display_name" validate:"required,min=1,max=255"`
	Description string                 `json:"description"`
	AvatarURL   string                 `json:"avatar_url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// BootstrapAgentProfile represents agent profile info
type BootstrapAgentProfile struct {
	Name        string                 `json:"name" validate:"required,min=3,max=100"`
	DisplayName string                 `json:"display_name" validate:"required,min=1,max=255"`
	Description string                 `json:"description"`
	AvatarURL   string                 `json:"avatar_url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// BootstrapHuman represents a human user to invite
type BootstrapHuman struct {
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role" validate:"oneof=observer member"`
}

// BootstrapChannel represents the default channel to create
type BootstrapChannel struct {
	Name  string `json:"name"`
	Topic string `json:"topic"`
}

// BootstrapResponse represents the complete bootstrap response
type BootstrapResponse struct {
	PrimaryAgent          BootstrapPrimaryAgentResponse `json:"primary_agent"`
	Agents                []BootstrapAgentResponse      `json:"agents"`
	Humans                []BootstrapHumanResponse      `json:"humans"`
	Channel               BootstrapChannelResponse      `json:"channel"`
	InstanceID            string                        `json:"instance_id"`
	RelayToken            string                        `json:"relay_token,omitempty"`
	RelaySubdomain        string                        `json:"relay_subdomain,omitempty"`
	RelayURL              string                        `json:"relay_url,omitempty"`
	RelayTier             string                        `json:"relay_tier,omitempty"`
	RelayBandwidthLimitMB int                           `json:"relay_bandwidth_limit_mb,omitempty"`
}

// BootstrapPrimaryAgentResponse represents primary agent credentials
type BootstrapPrimaryAgentResponse struct {
	UserID   string `json:"user_id"`
	AgentID  string `json:"agent_id"`
	Email    string `json:"email"`
	JWTToken string `json:"jwt_token"`
	APIKey   string `json:"api_key"`
	APIKeyID string `json:"api_key_id"`
}

// BootstrapAgentResponse represents agent credentials
type BootstrapAgentResponse struct {
	AgentID     string `json:"agent_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	APIKey      string `json:"api_key"`
	APIKeyID    string `json:"api_key_id"`
}

// BootstrapHumanResponse represents human invite info
type BootstrapHumanResponse struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	InviteToken string `json:"invite_token"`
	InviteURL   string `json:"invite_url"`
}

// BootstrapChannelResponse represents created channel info
type BootstrapChannelResponse struct {
	ChannelID string   `json:"channel_id"`
	Name      string   `json:"name"`
	Topic     string   `json:"topic"`
	Members   []string `json:"members"`
}
