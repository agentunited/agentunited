package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/rs/zerolog/log"
)

var (
	ErrWebhookNotFound = errors.New("webhook not found")
)

// WebhookService handles webhook business logic
type WebhookService interface {
	CreateWebhook(ctx context.Context, agentID, ownerID string, req *models.CreateWebhookRequest) (*models.Webhook, error)
	ListWebhooks(ctx context.Context, agentID, ownerID string) ([]*models.Webhook, error)
	DeleteWebhook(ctx context.Context, webhookID, agentID, ownerID string) error
	ListDeliveries(ctx context.Context, webhookID, agentID, ownerID string, limit int) ([]*models.WebhookDelivery, error)
	DispatchEvent(ctx context.Context, channelID, eventType string, payload map[string]interface{})
}

type webhookService struct {
	webhookRepo repository.WebhookRepository
	agentRepo   repository.AgentRepository
}

// NewWebhookService creates a new webhook service
func NewWebhookService(webhookRepo repository.WebhookRepository, agentRepo repository.AgentRepository) WebhookService {
	return &webhookService{
		webhookRepo: webhookRepo,
		agentRepo:   agentRepo,
	}
}

func (s *webhookService) CreateWebhook(ctx context.Context, agentID, ownerID string, req *models.CreateWebhookRequest) (*models.Webhook, error) {
	// Verify ownership
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return nil, ErrAgentNotFound
	}
	if agent.OwnerID != ownerID {
		return nil, ErrNotAgentOwner
	}

	// Generate secret
	secretBytes := make([]byte, 32)
	rand.Read(secretBytes)
	secret := base64.URLEncoding.EncodeToString(secretBytes)

	webhook := &models.Webhook{
		AgentID: agentID,
		URL:     req.URL,
		Events:  req.Events,
		Active:  true,
	}

	if err := s.webhookRepo.Create(ctx, webhook, secret); err != nil {
		return nil, err
	}

	return webhook, nil
}

func (s *webhookService) ListWebhooks(ctx context.Context, agentID, ownerID string) ([]*models.Webhook, error) {
	// Verify ownership
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return nil, ErrAgentNotFound
	}
	if agent.OwnerID != ownerID {
		return nil, ErrNotAgentOwner
	}

	return s.webhookRepo.ListByAgent(ctx, agentID)
}

func (s *webhookService) DeleteWebhook(ctx context.Context, webhookID, agentID, ownerID string) error {
	// Verify ownership
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return ErrAgentNotFound
	}
	if agent.OwnerID != ownerID {
		return ErrNotAgentOwner
	}

	return s.webhookRepo.Delete(ctx, webhookID)
}

func (s *webhookService) ListDeliveries(ctx context.Context, webhookID, agentID, ownerID string, limit int) ([]*models.WebhookDelivery, error) {
	// Verify ownership
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return nil, ErrAgentNotFound
	}
	if agent.OwnerID != ownerID {
		return nil, ErrNotAgentOwner
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	return s.webhookRepo.ListDeliveries(ctx, webhookID, limit)
}

// DispatchEvent dispatches an event to all relevant webhooks asynchronously
func (s *webhookService) DispatchEvent(ctx context.Context, channelID, eventType string, payload map[string]interface{}) {
	// Run asynchronously to avoid blocking the caller
	go s.dispatchEventAsync(channelID, eventType, payload)
}

// dispatchEventAsync performs the actual webhook dispatch work
func (s *webhookService) dispatchEventAsync(channelID, eventType string, payload map[string]interface{}) {
	ctx := context.Background()

	// Get all relevant webhooks for this channel and event type
	webhooks, err := s.webhookRepo.ListByChannel(ctx, channelID, eventType)
	if err != nil {
		log.Error().Err(err).Str("channel_id", channelID).Str("event_type", eventType).
			Msg("failed to get webhooks for channel")
		return
	}

	if len(webhooks) == 0 {
		log.Debug().Str("channel_id", channelID).Str("event_type", eventType).
			Msg("no webhooks found for event")
		return
	}

	log.Info().Str("channel_id", channelID).Str("event_type", eventType).
		Int("webhook_count", len(webhooks)).Msg("dispatching webhooks")

	// Dispatch to each webhook
	for _, webhook := range webhooks {
		s.dispatchToWebhook(ctx, webhook, eventType, payload)
	}
}

// dispatchToWebhook dispatches to a single webhook with retry logic
func (s *webhookService) dispatchToWebhook(ctx context.Context, webhook *models.WebhookWithSecret, eventType string, payload map[string]interface{}) {
	delivery := &models.WebhookDelivery{
		WebhookID:    webhook.ID,
		EventType:    eventType,
		Payload:      payload,
		Status:       "pending",
		AttemptCount: 0,
	}

	// Create delivery record
	if err := s.webhookRepo.CreateDelivery(ctx, delivery); err != nil {
		log.Error().Err(err).Str("webhook_id", webhook.ID).Msg("failed to create delivery record")
		return
	}

	// Attempt delivery with retries
	success := false
	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts && !success; attempt++ {
		delivery.AttemptCount = attempt

		if s.attemptDelivery(webhook, delivery, payload) {
			delivery.Status = "success"
			now := time.Now()
			delivery.DeliveredAt = &now
			success = true
		} else {
			delivery.Status = "failed"
			if attempt < maxAttempts {
				// Exponential backoff: 2^attempt seconds
				backoffSeconds := 1 << attempt
				time.Sleep(time.Duration(backoffSeconds) * time.Second)
			}
		}

		// Update delivery record
		if err := s.webhookRepo.UpdateDelivery(ctx, delivery); err != nil {
			log.Error().Err(err).Str("delivery_id", delivery.ID).Msg("failed to update delivery record")
		}
	}

	logLevel := log.Info()
	if !success {
		logLevel = log.Warn()
	}
	logLevel.Str("webhook_id", webhook.ID).Str("url", webhook.URL).
		Str("event_type", eventType).Int("attempts", delivery.AttemptCount).
		Bool("success", success).Msg("webhook delivery completed")
}

// attemptDelivery attempts a single delivery to a webhook
func (s *webhookService) attemptDelivery(webhook *models.WebhookWithSecret, delivery *models.WebhookDelivery, payload map[string]interface{}) bool {
	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Str("webhook_id", webhook.ID).Msg("failed to marshal payload")
		delivery.ResponseCode = intPtr(0)
		delivery.ResponseBody = stringPtr(fmt.Sprintf("Failed to marshal payload: %v", err))
		return false
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		log.Error().Err(err).Str("webhook_id", webhook.ID).Str("url", webhook.URL).Msg("failed to create HTTP request")
		delivery.ResponseCode = intPtr(0)
		delivery.ResponseBody = stringPtr(fmt.Sprintf("Failed to create request: %v", err))
		return false
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AgentUnited-Webhooks/1.0")
	req.Header.Set("X-AgentUnited-Event", delivery.EventType)
	
	// Generate HMAC signature
	signature := generateHMACSignature(payloadBytes, webhook.Secret)
	req.Header.Set("X-AgentUnited-Signature", signature)

	// Make request with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("webhook_id", webhook.ID).Str("url", webhook.URL).Msg("failed to send webhook")
		delivery.ResponseCode = intPtr(0)
		delivery.ResponseBody = stringPtr(fmt.Sprintf("HTTP error: %v", err))
		return false
	}
	defer resp.Body.Close()

	// Read response body (limited to 1KB to avoid memory issues)
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	delivery.ResponseCode = &resp.StatusCode
	delivery.ResponseBody = stringPtr(string(bodyBytes))

	// Check if response indicates success (2xx status codes)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Debug().Str("webhook_id", webhook.ID).Str("url", webhook.URL).
			Int("status_code", resp.StatusCode).Msg("webhook delivered successfully")
		return true
	}

	log.Warn().Str("webhook_id", webhook.ID).Str("url", webhook.URL).
		Int("status_code", resp.StatusCode).Str("response", string(bodyBytes)).
		Msg("webhook returned non-success status")
	return false
}

// generateHMACSignature generates HMAC-SHA256 signature for webhook verification
func generateHMACSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("sha256=%s", signature)
}

// Helper functions for pointer types
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
