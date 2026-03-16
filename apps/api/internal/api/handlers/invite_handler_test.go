package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockInviteService struct {
	mock.Mock
}

func (m *mockInviteService) ValidateInvite(ctx context.Context, token string) (*models.Invite, *models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*models.Invite), args.Get(1).(*models.User), args.Error(2)
}

func (m *mockInviteService) AcceptInvite(ctx context.Context, token, password, displayName string) (string, string, error) {
	args := m.Called(ctx, token, password, displayName)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockInviteService) CreateInvite(ctx context.Context, workspaceID, email, displayName string) (string, string, error) {
	args := m.Called(ctx, workspaceID, email, displayName)
	return args.String(0), args.String(1), args.Error(2)
}

func TestInviteHandler_ValidateInvite_HappyPath(t *testing.T) {
	mockService := &mockInviteService{}
	handler := NewInviteHandler(mockService)

	invite := &models.Invite{
		ID:        "invite-123",
		UserID:    "user-456",
		Status:    models.InviteStatusPending,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	user := &models.User{
		ID:    "user-456",
		Email: "test@example.com",
	}

	mockService.On("ValidateInvite", mock.Anything, "valid-token").Return(invite, user, nil)

	req, _ := http.NewRequest("GET", "/api/v1/invite?token=valid-token", nil)
	rr := httptest.NewRecorder()

	handler.ValidateInvite(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "test@example.com", resp["email"])
	assert.Equal(t, "pending", resp["status"])
	assert.NotEmpty(t, resp["expires_at"])

	mockService.AssertExpectations(t)
}

func TestInviteHandler_ValidateInvite_InvalidToken(t *testing.T) {
	mockService := &mockInviteService{}
	handler := NewInviteHandler(mockService)

	mockService.On("ValidateInvite", mock.Anything, "invalid-token").Return(nil, nil, models.ErrInviteNotFound)

	req, _ := http.NewRequest("GET", "/api/v1/invite?token=invalid-token", nil)
	rr := httptest.NewRecorder()

	handler.ValidateInvite(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "invite not found")

	mockService.AssertExpectations(t)
}

func TestInviteHandler_ValidateInvite_ExpiredToken(t *testing.T) {
	mockService := &mockInviteService{}
	handler := NewInviteHandler(mockService)

	mockService.On("ValidateInvite", mock.Anything, "expired-token").Return(nil, nil, models.ErrInviteExpired)

	req, _ := http.NewRequest("GET", "/api/v1/invite?token=expired-token", nil)
	rr := httptest.NewRecorder()

	handler.ValidateInvite(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "invite has expired")

	mockService.AssertExpectations(t)
}

func TestInviteHandler_ValidateInvite_MissingToken(t *testing.T) {
	mockService := &mockInviteService{}
	handler := NewInviteHandler(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/invite", nil) // No token query param
	rr := httptest.NewRecorder()

	handler.ValidateInvite(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "token is required")
}

func TestInviteHandler_AcceptInvite_HappyPath(t *testing.T) {
	mockService := &mockInviteService{}
	handler := NewInviteHandler(mockService)

	mockService.On("AcceptInvite", mock.Anything, "valid-token", "securepassword123", "").Return("jwt-token", "", nil)

	reqBody := models.InviteAcceptRequest{
		Token:    "valid-token",
		Password: "securepassword123",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/invite/accept", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.AcceptInvite(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "jwt-token", resp["jwt_token"])

	mockService.AssertExpectations(t)
}

func TestInviteHandler_AcceptInvite_InvalidToken(t *testing.T) {
	mockService := &mockInviteService{}
	handler := NewInviteHandler(mockService)

	mockService.On("AcceptInvite", mock.Anything, "invalid-token", "securepassword123", "").Return("", "", models.ErrInviteNotFound)

	reqBody := models.InviteAcceptRequest{
		Token:    "invalid-token",
		Password: "securepassword123",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/invite/accept", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.AcceptInvite(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "invite not found")

	mockService.AssertExpectations(t)
}

func TestInviteHandler_AcceptInvite_WeakPassword(t *testing.T) {
	mockService := &mockInviteService{}
	handler := NewInviteHandler(mockService)

	reqBody := models.InviteAcceptRequest{
		Token:    "valid-token",
		Password: "weak", // Too short
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/invite/accept", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.AcceptInvite(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "validation")
}

func TestInviteHandler_AcceptInvite_InvalidJSON(t *testing.T) {
	mockService := &mockInviteService{}
	handler := NewInviteHandler(mockService)

	req, _ := http.NewRequest("POST", "/api/v1/invite/accept", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.AcceptInvite(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "invalid JSON")
}
