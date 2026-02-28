package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
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
