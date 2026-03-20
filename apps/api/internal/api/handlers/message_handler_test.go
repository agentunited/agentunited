package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMessageService struct{ mock.Mock }

type mockWebhookService struct{ mock.Mock }

type mockRealtime struct{ mock.Mock }

type mockBroadcaster struct{ mock.Mock }

func (m *mockMessageService) Send(ctx context.Context, channelID, userID, text string) (*models.Message, error) {
	args := m.Called(ctx, channelID, userID, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}
func (m *mockMessageService) SendAsAgent(ctx context.Context, channelID, ownerID string, agent service.AgentContext, text string) (*models.Message, error) {
	args := m.Called(ctx, channelID, ownerID, agent, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}
func (m *mockMessageService) SendMessageWithAttachment(ctx context.Context, message *models.Message, agentCtx *service.AgentContext) (*models.Message, error) {
	args := m.Called(ctx, message, agentCtx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}
func (m *mockMessageService) GetMessages(ctx context.Context, channelID, userID string, limit int, before string) (*models.MessageList, error) {
	args := m.Called(ctx, channelID, userID, limit, before)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MessageList), args.Error(1)
}
func (m *mockMessageService) EditMessage(ctx context.Context, messageID, userID, text string) (*models.Message, error) {
	args := m.Called(ctx, messageID, userID, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}
func (m *mockMessageService) DeleteMessage(ctx context.Context, messageID, userID string) error {
	return m.Called(ctx, messageID, userID).Error(0)
}
func (m *mockMessageService) SearchMessages(ctx context.Context, query, channelID, userID string, limit int) ([]*models.Message, error) {
	args := m.Called(ctx, query, channelID, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *mockWebhookService) CreateWebhook(ctx context.Context, agentID, ownerID string, req *models.CreateWebhookRequest) (*models.Webhook, error) {
	return nil, nil
}
func (m *mockWebhookService) ListWebhooks(ctx context.Context, agentID, ownerID string) ([]*models.Webhook, error) {
	return nil, nil
}
func (m *mockWebhookService) DeleteWebhook(ctx context.Context, webhookID, agentID, ownerID string) error {
	return nil
}
func (m *mockWebhookService) ListDeliveries(ctx context.Context, webhookID, agentID, ownerID string, limit int) ([]*models.WebhookDelivery, error) {
	return nil, nil
}
func (m *mockWebhookService) DispatchEvent(ctx context.Context, channelID, eventType string, payload map[string]interface{}) {
	m.Called(ctx, channelID, eventType, payload)
}

func (m *mockWebhookService) DispatchSigned(ctx context.Context, url, eventType, integrationID string, payload []byte, signature string) {
	m.Called(ctx, url, eventType, integrationID, payload, signature)
}

func (m *mockRealtime) Enabled() bool {
	return m.Called().Bool(0)
}
func (m *mockRealtime) Publish(ctx context.Context, channelID string, payload any) error {
	return m.Called(ctx, channelID, payload).Error(0)
}

func (m *mockBroadcaster) Broadcast(ctx context.Context, channelID string, message []byte) {
	m.Called(ctx, channelID, message)
}

func buildSendRequest(t *testing.T) *http.Request {
	body, _ := json.Marshal(map[string]string{"text": "hello"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/ch1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "u1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "ch1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestSend_PublishesToCentrifugo_WhenEnabled(t *testing.T) {
	ms := new(mockMessageService)
	ws := new(mockWebhookService)
	rt := new(mockRealtime)
	hb := new(mockBroadcaster)
	h := NewMessageHandler(ms, ws, hb, rt)

	msg := &models.Message{ID: "m1", ChannelID: "ch1", AuthorID: "u1", Text: "hello", CreatedAt: time.Now()}
	ms.On("SendMessageWithAttachment", mock.Anything, mock.AnythingOfType("*models.Message"), (*service.AgentContext)(nil)).Return(msg, nil)
	ws.On("DispatchEvent", mock.Anything, "ch1", "message.created", mock.Anything).Return()
	hb.On("Broadcast", mock.Anything, "ch1", mock.AnythingOfType("[]uint8")).Return()
	rt.On("Enabled").Return(true)
	rt.On("Publish", mock.Anything, "ch1", mock.Anything).Return(nil)

	rr := httptest.NewRecorder()
	h.Send(rr, buildSendRequest(t))

	assert.Equal(t, http.StatusCreated, rr.Code)
	rt.AssertCalled(t, "Publish", mock.Anything, "ch1", mock.Anything)
	hb.AssertCalled(t, "Broadcast", mock.Anything, "ch1", mock.AnythingOfType("[]uint8"))
}

func TestSend_AgentAuthoredMessage_DoesNotDispatchWebhook(t *testing.T) {
	ms := new(mockMessageService)
	ws := new(mockWebhookService)
	rt := new(mockRealtime)
	hb := new(mockBroadcaster)
	h := NewMessageHandler(ms, ws, hb, rt)

	msg := &models.Message{ID: "m1", ChannelID: "ch1", AuthorID: "a1", AuthorType: "agent", Text: "hello", CreatedAt: time.Now()}
	ms.On("SendMessageWithAttachment", mock.Anything, mock.AnythingOfType("*models.Message"), mock.AnythingOfType("*service.AgentContext")).Return(msg, nil)
	rt.On("Enabled").Return(false)
	hb.On("Broadcast", mock.Anything, "ch1", mock.AnythingOfType("[]uint8")).Return()

	req := buildSendRequest(t)
	ctx := context.WithValue(req.Context(), middleware.AgentIDKey, "a1")
	ctx = context.WithValue(ctx, middleware.AgentNameKey, "Agent One")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Send(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	ws.AssertNotCalled(t, "DispatchEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestSend_UsesHubFallback_WhenCentrifugoDisabled(t *testing.T) {
	ms := new(mockMessageService)
	ws := new(mockWebhookService)
	rt := new(mockRealtime)
	hb := new(mockBroadcaster)
	h := NewMessageHandler(ms, ws, hb, rt)

	msg := &models.Message{ID: "m1", ChannelID: "ch1", AuthorID: "u1", Text: "hello", CreatedAt: time.Now()}
	ms.On("SendMessageWithAttachment", mock.Anything, mock.AnythingOfType("*models.Message"), (*service.AgentContext)(nil)).Return(msg, nil)
	ws.On("DispatchEvent", mock.Anything, "ch1", "message.created", mock.Anything).Return()
	rt.On("Enabled").Return(false)
	hb.On("Broadcast", mock.Anything, "ch1", mock.AnythingOfType("[]uint8")).Return()

	rr := httptest.NewRecorder()
	h.Send(rr, buildSendRequest(t))

	assert.Equal(t, http.StatusCreated, rr.Code)
	hb.AssertCalled(t, "Broadcast", mock.Anything, "ch1", mock.AnythingOfType("[]uint8"))
	rt.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}
