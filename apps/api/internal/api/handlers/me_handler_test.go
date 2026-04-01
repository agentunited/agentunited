package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mw "github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMeAuthService struct{ mock.Mock }

func (m *mockMeAuthService) Register(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockMeAuthService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}
func (m *mockMeAuthService) GetCurrentUser(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockMeAuthService) UpdateProfile(ctx context.Context, userID, displayName, avatarURL string) (*models.User, error) {
	args := m.Called(ctx, userID, displayName, avatarURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockMeAuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	args := m.Called(ctx, userID, currentPassword, newPassword)
	return args.Error(0)
}

func (m *mockMeAuthService) ForgotPassword(ctx context.Context, email string) error {
	return nil
}

func (m *mockMeAuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	return nil
}

func withUser(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), mw.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestMeHandler_GetMe_Success(t *testing.T) {
	svc := new(mockMeAuthService)
	h := NewMeHandler(svc)
	u := &models.User{ID: "u1", Email: "u1@example.com", DisplayName: "U1", AvatarURL: "https://a", CreatedAt: time.Now()}
	svc.On("GetCurrentUser", mock.Anything, "u1").Return(u, nil)

	req := withUser(httptest.NewRequest(http.MethodGet, "/api/v1/me", nil), "u1")
	rr := httptest.NewRecorder()
	h.GetMe(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var got map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	assert.Equal(t, "u1", got["id"])
	assert.Equal(t, "u1@example.com", got["email"])
}

func TestMeHandler_GetMe_AuthRequired(t *testing.T) {
	h := NewMeHandler(new(mockMeAuthService))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rr := httptest.NewRecorder()
	h.GetMe(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMeHandler_UpdateMe_Success(t *testing.T) {
	svc := new(mockMeAuthService)
	h := NewMeHandler(svc)
	u := &models.User{ID: "u1", Email: "u1@example.com", DisplayName: "Naomi", CreatedAt: time.Now()}
	svc.On("UpdateProfile", mock.Anything, "u1", "Naomi", "").Return(u, nil)

	body := []byte(`{"display_name":"Naomi"}`)
	req := withUser(httptest.NewRequest(http.MethodPut, "/api/v1/me", bytes.NewReader(body)), "u1")
	rr := httptest.NewRecorder()
	h.UpdateMe(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMeHandler_UpdateMe_AuthRequired(t *testing.T) {
	h := NewMeHandler(new(mockMeAuthService))
	req := httptest.NewRequest(http.MethodPut, "/api/v1/me", bytes.NewReader([]byte(`{"display_name":"x"}`)))
	rr := httptest.NewRecorder()
	h.UpdateMe(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMeHandler_ChangePassword_Success(t *testing.T) {
	svc := new(mockMeAuthService)
	h := NewMeHandler(svc)
	svc.On("ChangePassword", mock.Anything, "u1", "oldpass123", "newpass123").Return(nil)

	body := []byte(`{"current_password":"oldpass123","new_password":"newpass123"}`)
	req := withUser(httptest.NewRequest(http.MethodPost, "/api/v1/me/password", bytes.NewReader(body)), "u1")
	rr := httptest.NewRecorder()
	h.ChangePassword(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMeHandler_ChangePassword_AuthRequired(t *testing.T) {
	h := NewMeHandler(new(mockMeAuthService))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/password", bytes.NewReader([]byte(`{"current_password":"a","new_password":"b"}`)))
	rr := httptest.NewRecorder()
	h.ChangePassword(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
