package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/agentunited/backend/pkg/integrations"
)

type OutboundDispatcher interface {
	DispatchSigned(ctx context.Context, url, eventType, integrationID string, payload []byte, secret string)
}

type IntegrationService interface {
	List(ctx context.Context, workspaceID string) ([]*models.Integration, error)
	Create(ctx context.Context, workspaceID string, req *models.CreateIntegrationRequest) (*models.Integration, error)
	Delete(ctx context.Context, workspaceID, id string) error
	RouteEvent(ctx context.Context, event integrations.Event) error
}

type integrationService struct {
	repo       repository.IntegrationRepository
	dispatcher OutboundDispatcher
}

func NewIntegrationService(repo repository.IntegrationRepository, dispatcher OutboundDispatcher) IntegrationService {
	return &integrationService{repo: repo, dispatcher: dispatcher}
}

func (s *integrationService) List(ctx context.Context, workspaceID string) ([]*models.Integration, error) {
	list, err := s.repo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	for _, i := range list {
		i.APIKey = ""
	}
	return list, nil
}

func (s *integrationService) Create(ctx context.Context, workspaceID string, req *models.CreateIntegrationRequest) (*models.Integration, error) {
	platform := integrations.Platform(strings.ToLower(req.Platform))
	adapter, ok := integrations.GetAdapter(platform)
	if !ok && platform != integrations.PlatformLangGraph && platform != integrations.PlatformAutoGen && platform != integrations.PlatformCustom {
		return nil, fmt.Errorf("unsupported platform")
	}
	if ok {
		if err := adapter.ValidateCredentials(map[string]string{"webhook_url": req.WebhookURL}); err != nil {
			return nil, err
		}
	}

	events := req.EventSubscriptions
	if len(events) == 0 && len(req.Events) > 0 {
		events = req.Events
	}
	if events == nil {
		events = []string{}
	}

	apiKey := generateToken("au_int_")
	i := &models.Integration{WorkspaceID: workspaceID, Name: req.Name, Platform: string(platform), APIKey: apiKey, WebhookURL: req.WebhookURL, EventSubscriptions: events, Active: true}
	if err := s.repo.Create(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *integrationService) Delete(ctx context.Context, workspaceID, id string) error {
	return s.repo.Delete(ctx, id, workspaceID)
}

func (s *integrationService) RouteEvent(ctx context.Context, event integrations.Event) error {
	list, err := s.repo.ListActiveByEvent(ctx, event.WorkspaceID, event.Type)
	if err != nil {
		return err
	}

	for _, integ := range list {
		adapter, ok := integrations.GetAdapter(integrations.Platform(integ.Platform))
		if !ok {
			continue
		}
		payload, err := adapter.FormatOutbound(event)
		if err != nil {
			continue
		}
		h := hmac.New(sha256.New, []byte(integ.APIKey))
		h.Write(payload)
		sig := "sha256=" + hex.EncodeToString(h.Sum(nil))
		s.dispatcher.DispatchSigned(ctx, integ.WebhookURL, event.Type, integ.ID, payload, sig)
	}
	return nil
}

func generateToken(prefix string) string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

func NewEvent(eventType, workspaceID, channelID string, payload map[string]interface{}) integrations.Event {
	return integrations.Event{Type: eventType, WorkspaceID: workspaceID, ChannelID: channelID, Payload: payload, OccurredAt: time.Now().UTC()}
}
