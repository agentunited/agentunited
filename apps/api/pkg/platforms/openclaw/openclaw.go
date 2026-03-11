package openclaw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/agentunited/backend/pkg/platforms"
)

const (
	defaultTimeout = 30 * time.Second
)

// Adapter implements PlatformAdapter for OpenClaw
type Adapter struct {
	gatewayURL string
	apiKey     string
	httpClient *http.Client
}

// NewAdapter creates a new OpenClaw adapter
func NewAdapter() *Adapter {
	return &Adapter{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Platform returns the platform identifier
func (a *Adapter) Platform() string { return "openclaw" }

// Version returns the adapter version
func (a *Adapter) Version() string { return "1.0.0" }

// Initialize sets up the adapter with config
func (a *Adapter) Initialize(ctx context.Context, config map[string]interface{}) error {
	gatewayURL, ok := config["gateway_url"].(string)
	if !ok || gatewayURL == "" {
		return fmt.Errorf("gateway_url is required")
	}
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("api_key is required")
	}

	a.gatewayURL = gatewayURL
	a.apiKey = apiKey
	return nil
}

// Shutdown cleans up resources
func (a *Adapter) Shutdown(ctx context.Context) error {
	return nil
}

// Authenticate validates the API key
func (a *Adapter) Authenticate(ctx context.Context, creds platforms.Credentials) (*platforms.AuthResult, error) {
	// OpenClaw uses static API keys, validation happens via health check
	health, err := a.CheckHealth(ctx)
	if err != nil {
		return &platforms.AuthResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &platforms.AuthResult{
		Success: health.Healthy,
	}, nil
}

// GetCapabilities returns OpenClaw's supported features
func (a *Adapter) GetCapabilities() platforms.Capabilities {
	return platforms.Capabilities{
		SupportsChannels:   true,
		SupportsDMs:        true,
		SupportsReactions:  true,
		SupportsFiles:      true,
		SupportsThreading:  false,
		SupportsPresence:   true,
		SupportsHistory:    true,
		MaxFileSize:        25 * 1024 * 1024, // 25MB
		RateLimitPerMinute: 60,
	}
}

// SendMessage sends a message to an OpenClaw channel
func (a *Adapter) SendMessage(ctx context.Context, channelID string, msg platforms.NormalizedMessage) (*platforms.DeliveryResult, error) {
	url := fmt.Sprintf("%s/api/v1/channels/%s/messages", a.gatewayURL, channelID)

	payload := map[string]interface{}{
		"text": msg.Text,
	}
	if msg.ThreadID != "" {
		payload["thread_id"] = msg.ThreadID
	}
	if msg.ReplyTo != "" {
		payload["reply_to"] = msg.ReplyTo
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return &platforms.DeliveryResult{
			Success: false,
			Error:   fmt.Sprintf("API error: %s - %s", resp.Status, string(respBody)),
		}, nil
	}

	var result struct {
		MessageID string `json:"message_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Non-fatal, message was likely sent
		result.MessageID = ""
	}

	return &platforms.DeliveryResult{
		Success:   true,
		MessageID: result.MessageID,
		Timestamp: time.Now(),
	}, nil
}

// ReceiveMessage normalizes an incoming webhook payload
func (a *Adapter) ReceiveMessage(ctx context.Context, payload []byte) (*platforms.NormalizedMessage, error) {
	var raw struct {
		ID        string `json:"id"`
		ChannelID string `json:"channel_id"`
		ThreadID  string `json:"thread_id,omitempty"`
		UserID    string `json:"user_id"`
		Text      string `json:"text"`
		Timestamp int64  `json:"timestamp"`
		ReplyTo   string `json:"reply_to,omitempty"`
	}

	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	return &platforms.NormalizedMessage{
		ID:        raw.ID,
		ChannelID: raw.ChannelID,
		ThreadID:  raw.ThreadID,
		UserID:    raw.UserID,
		Text:      raw.Text,
		Timestamp: time.Unix(raw.Timestamp, 0),
		ReplyTo:   raw.ReplyTo,
	}, nil
}

// GetStatus returns the current platform status
func (a *Adapter) GetStatus(ctx context.Context) (*platforms.PlatformStatus, error) {
	return a.CheckHealth(ctx)
}

// CheckHealth verifies connectivity to the OpenClaw gateway
func (a *Adapter) CheckHealth(ctx context.Context) (*platforms.HealthStatus, error) {
	start := time.Now()

	url := fmt.Sprintf("%s/health", a.gatewayURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &platforms.HealthStatus{
			Connected: false,
			Healthy:   false,
			Error:     err.Error(),
		}, nil
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.httpClient.Do(req)
	latency := time.Since(start)

	if err != nil {
		return &platforms.HealthStatus{
			Connected: false,
			Healthy:   false,
			Error:     err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	return &platforms.HealthStatus{
		Connected: resp.StatusCode < 500,
		Healthy:   resp.StatusCode == 200,
		LastPing:  time.Now(),
		LatencyMs: int(latency.Milliseconds()),
	}, nil
}

// ValidateConfig checks if the provided config is valid
func (a *Adapter) ValidateConfig(config map[string]interface{}) (*platforms.ValidationResult, error) {
	var errors []string
	var warnings []string

	if gatewayURL, ok := config["gateway_url"].(string); !ok || gatewayURL == "" {
		errors = append(errors, "gateway_url is required")
	}

	if apiKey, ok := config["api_key"].(string); !ok || apiKey == "" {
		errors = append(errors, "api_key is required")
	}

	return &platforms.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
		Warn:   warnings,
	}, nil
}
