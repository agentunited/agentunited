package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockBootstrapService struct {
	mock.Mock
}

type mockBootstrapRateLimiter struct {
	mock.Mock
}

func (m *mockBootstrapRateLimiter) Allow(ctx context.Context, ip string) (bool, error) {
	args := m.Called(ctx, ip)
	return args.Bool(0), args.Error(1)
}

func (m *mockBootstrapService) Bootstrap(ctx context.Context, req *models.BootstrapRequest) (*models.BootstrapResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BootstrapResponse), args.Error(1)
}

func TestBootstrapHandler_Bootstrap_HappyPath(t *testing.T) {
	mockService := &mockBootstrapService{}
	handler := NewBootstrapHandler(mockService, nil)

	// Mock service response
	mockResp := &models.BootstrapResponse{
		PrimaryAgent: models.BootstrapPrimaryAgentResponse{
			UserID:   "user-123",
			AgentID:  "agent-123",
			Email:    "admin@example.com",
			JWTToken: "jwt-token",
			APIKey:   "au_test123",
			APIKeyID: "key-123",
		},
		Agents: []models.BootstrapAgentResponse{
			{
				AgentID:     "agent-456",
				Name:        "worker",
				DisplayName: "Worker Agent",
				APIKey:      "au_worker123",
				APIKeyID:    "key-456",
			},
		},
		Humans: []models.BootstrapHumanResponse{
			{
				UserID:      "user-789",
				Email:       "human@example.com",
				InviteToken: "inv_test123",
				InviteURL:   "http://localhost:8080/invite?token=inv_test123",
			},
		},
		Channel: models.BootstrapChannelResponse{
			ChannelID: "channel-123",
			Name:      "general",
			Topic:     "General discussion",
			Members:   []string{"user-123", "user-789"},
		},
		InstanceID:            "instance-123",
		RelayURL:              "https://abc123.tunnel.agentunited.ai",
		RelayTier:             "free",
		RelayBandwidthLimitMB: 1024,
		RelaySubdomain:        "abc123",
	}

	mockService.On("Bootstrap", mock.Anything, mock.AnythingOfType("*models.BootstrapRequest")).Return(mockResp, nil)

	reqBody := models.BootstrapRequest{
		PrimaryAgent: models.BootstrapPrimaryAgent{
			Email:    "admin@example.com",
			Password: "supersecurepassword123",
			AgentProfile: models.BootstrapAgentProfile{
				Name:        "coordinator",
				DisplayName: "Coordinator Agent",
			},
		},
		Agents: []models.BootstrapAgent{
			{
				Name:        "worker",
				DisplayName: "Worker Agent",
			},
		},
		Humans: []models.BootstrapHuman{
			{
				Email:       "human@example.com",
				DisplayName: "Human User",
				Role:        "member",
			},
		},
		DefaultChannel: models.BootstrapChannel{
			Name:  "general",
			Topic: "General discussion",
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/bootstrap", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp models.BootstrapResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "admin@example.com", resp.PrimaryAgent.Email)
	assert.Equal(t, "au_test123", resp.PrimaryAgent.APIKey)
	assert.Len(t, resp.Agents, 1)
	assert.Equal(t, "worker", resp.Agents[0].Name)
	assert.Len(t, resp.Humans, 1)
	assert.Equal(t, "human@example.com", resp.Humans[0].Email)
	assert.Equal(t, "general", resp.Channel.Name)
	assert.Equal(t, "https://abc123.tunnel.agentunited.ai", resp.RelayURL)
	assert.Equal(t, "free", resp.RelayTier)
	assert.Equal(t, 1024, resp.RelayBandwidthLimitMB)

	mockService.AssertExpectations(t)
}

func TestBootstrapHandler_Bootstrap_InstanceAlreadyBootstrapped(t *testing.T) {
	mockService := &mockBootstrapService{}
	handler := NewBootstrapHandler(mockService, nil)

	mockService.On("Bootstrap", mock.Anything, mock.AnythingOfType("*models.BootstrapRequest")).Return(nil, models.ErrInstanceAlreadyBootstrapped)

	reqBody := models.BootstrapRequest{
		PrimaryAgent: models.BootstrapPrimaryAgent{
			Email:    "admin@example.com",
			Password: "supersecurepassword123",
			AgentProfile: models.BootstrapAgentProfile{
				Name:        "coordinator",
				DisplayName: "Coordinator Agent",
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/bootstrap", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "instance has already been bootstrapped")

	mockService.AssertExpectations(t)
}

func TestBootstrapHandler_Bootstrap_InvalidJSON(t *testing.T) {
	mockService := &mockBootstrapService{}
	handler := NewBootstrapHandler(mockService, nil)

	req, _ := http.NewRequest("POST", "/api/v1/bootstrap", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "invalid JSON")
}

func TestBootstrapHandler_Bootstrap_RateLimited(t *testing.T) {
	mockService := &mockBootstrapService{}
	mockLimiter := &mockBootstrapRateLimiter{}
	handler := NewBootstrapHandlerWithLimiter(mockService, mockLimiter)

	reqBody := models.BootstrapRequest{
		PrimaryAgent: models.BootstrapPrimaryAgent{
			Email:    "admin@example.com",
			Password: "supersecurepassword123",
			AgentProfile: models.BootstrapAgentProfile{
				Name:        "coordinator",
				DisplayName: "Coordinator Agent",
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/bootstrap", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "203.0.113.10:9999"

	mockLimiter.On("Allow", mock.Anything, "203.0.113.10").Return(true, nil)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "max 3 requests per IP per day")
	mockService.AssertNotCalled(t, "Bootstrap", mock.Anything, mock.Anything)
	mockLimiter.AssertExpectations(t)
}

func TestBootstrapHandler_Bootstrap_ValidationError(t *testing.T) {
	mockService := &mockBootstrapService{}
	handler := NewBootstrapHandler(mockService, nil)

	// Missing required fields
	reqBody := models.BootstrapRequest{
		PrimaryAgent: models.BootstrapPrimaryAgent{
			// Missing email and password
			AgentProfile: models.BootstrapAgentProfile{
				Name:        "coordinator",
				DisplayName: "Coordinator Agent",
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/bootstrap", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "validation")
}
