package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockChannelService is a mock implementation of ChannelService
type MockChannelService struct {
	mock.Mock
}

func (m *MockChannelService) Create(ctx context.Context, userID, name, topic string) (*models.Channel, error) {
	args := m.Called(ctx, userID, name, topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *MockChannelService) Get(ctx context.Context, channelID, userID string) (*models.ChannelWithMembers, error) {
	args := m.Called(ctx, channelID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChannelWithMembers), args.Error(1)
}

func (m *MockChannelService) List(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ChannelWithMembers), args.Error(1)
}

func (m *MockChannelService) IsMember(ctx context.Context, channelID, userID string) (bool, string, error) {
	args := m.Called(ctx, channelID, userID)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockChannelService) Update(ctx context.Context, channelID, userID, name, topic string) (*models.Channel, error) {
	args := m.Called(ctx, channelID, userID, name, topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *MockChannelService) Delete(ctx context.Context, channelID, userID string) error {
	args := m.Called(ctx, channelID, userID)
	return args.Error(0)
}

func (m *MockChannelService) GetMembers(ctx context.Context, channelID, userID string) ([]*models.MemberInfo, error) {
	args := m.Called(ctx, channelID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MemberInfo), args.Error(1)
}

func (m *MockChannelService) AddMember(ctx context.Context, channelID, targetUserID, requesterUserID, role string) error {
	args := m.Called(ctx, channelID, targetUserID, requesterUserID, role)
	return args.Error(0)
}

func (m *MockChannelService) RemoveMember(ctx context.Context, channelID, targetUserID, requesterUserID string) error {
	args := m.Called(ctx, channelID, targetUserID, requesterUserID)
	return args.Error(0)
}

func (m *MockChannelService) ListDMChannels(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ChannelWithMembers), args.Error(1)
}

func (m *MockChannelService) CreateOrGetDMChannel(ctx context.Context, userID, targetUserID string) (*models.Channel, error) {
	args := m.Called(ctx, userID, targetUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *MockChannelService) MarkChannelRead(ctx context.Context, channelID, userID string) error {
	args := m.Called(ctx, channelID, userID)
	return args.Error(0)
}

// Test: Create channel with valid input
func TestChannelHandler_Create_Success(t *testing.T) {
	mockService := new(MockChannelService)
	handler := NewChannelHandler(mockService)

	channel := &models.Channel{
		ID:        "test-channel-id",
		Name:      "general",
		Topic:     "General discussion",
		CreatedBy: "test-user-id",
	}

	mockService.On("Create", mock.Anything, "test-user-id", "general", "General discussion").
		Return(channel, nil)

	reqBody := map[string]string{
		"name":  "general",
		"topic": "General discussion",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Add user ID to context (simulating auth middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user-id")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "channel")
	ch := response["channel"].(map[string]interface{})
	assert.Equal(t, "general", ch["name"])

	mockService.AssertExpectations(t)
}

// Test: Create channel with missing user ID (unauthenticated)
func TestChannelHandler_Create_Unauthenticated(t *testing.T) {
	mockService := new(MockChannelService)
	handler := NewChannelHandler(mockService)

	reqBody := map[string]string{
		"name":  "general",
		"topic": "General discussion",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// Test: Create channel with invalid name
func TestChannelHandler_Create_InvalidName(t *testing.T) {
	mockService := new(MockChannelService)
	handler := NewChannelHandler(mockService)

	mockService.On("Create", mock.Anything, "test-user-id", "Invalid Name", "Topic").
		Return(nil, models.ErrInvalidChannelName)

	reqBody := map[string]string{
		"name":  "Invalid Name",
		"topic": "Topic",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user-id")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	mockService.AssertExpectations(t)
}

// Test: List channels for authenticated user
func TestChannelHandler_List_Success(t *testing.T) {
	mockService := new(MockChannelService)
	handler := NewChannelHandler(mockService)

	channels := []*models.ChannelWithMembers{
		{
			Channel: models.Channel{
				ID:   "channel-1",
				Name: "general",
			},
			MemberCount: 5,
		},
		{
			Channel: models.Channel{
				ID:   "channel-2",
				Name: "random",
			},
			MemberCount: 3,
		},
	}

	mockService.On("List", mock.Anything, "test-user-id").Return(channels, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user-id")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "channels")
	chList := response["channels"].([]interface{})
	assert.Len(t, chList, 2)

	mockService.AssertExpectations(t)
}

// Test: Get channel by ID with members
func TestChannelHandler_Get_Success(t *testing.T) {
	mockService := new(MockChannelService)
	handler := NewChannelHandler(mockService)

	channelWithMembers := &models.ChannelWithMembers{
		Channel: models.Channel{
			ID:   "test-channel-id",
			Name: "general",
		},
		Members: []*models.MemberInfo{
			{ID: "user1", Email: "user1@example.com", Role: "owner"},
			{ID: "user2", Email: "user2@example.com", Role: "member"},
		},
	}

	mockService.On("Get", mock.Anything, "test-channel-id", "test-user-id").
		Return(channelWithMembers, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/test-channel-id", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user-id")
	req = req.WithContext(ctx)

	// Setup Chi URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-channel-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Get(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "channel")
	ch := response["channel"].(map[string]interface{})
	assert.Equal(t, "general", ch["name"])

	members := ch["members"].([]interface{})
	assert.Len(t, members, 2)

	mockService.AssertExpectations(t)
}

// Test: Get channel when user is not a member
func TestChannelHandler_Get_NotMember(t *testing.T) {
	mockService := new(MockChannelService)
	handler := NewChannelHandler(mockService)

	mockService.On("Get", mock.Anything, "test-channel-id", "test-user-id").
		Return(nil, models.ErrNotChannelMember)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/test-channel-id", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user-id")
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-channel-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Get(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	mockService.AssertExpectations(t)
}
