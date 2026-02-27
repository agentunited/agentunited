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

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

// Test: Register with valid input returns 201 with user and token
func TestAuthHandler_Register_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Create test user
	testUser := &models.User{
		ID:    "test-uuid-1234",
		Email: "test@example.com",
	}

	// Mock service to return user
	mockService.On("Register", mock.Anything, "test@example.com", "password123").
		Return(testUser, nil)

	// Mock Login to return token (Register calls Login to generate token)
	mockService.On("Login", mock.Anything, "test@example.com", "password123").
		Return("test-jwt-token", nil)

	// Create request
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Assert response structure
	assert.Contains(t, response, "user")
	assert.Contains(t, response, "token")

	user, ok := response["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-uuid-1234", user["id"])
	assert.Equal(t, "test@example.com", user["email"])

	token, ok := response["token"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, token)

	mockService.AssertExpectations(t)
}

// Test: Register with invalid JSON returns 400
func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Parse response
	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

// Test: Register with missing email returns 400
func TestAuthHandler_Register_MissingEmail(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Create request with missing email
	reqBody := map[string]string{
		"password": "password123",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// Test: Register with weak password returns 400
func TestAuthHandler_Register_WeakPassword(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Mock service to return weak password error
	mockService.On("Register", mock.Anything, "test@example.com", "weak").
		Return(nil, models.ErrWeakPassword)

	// Create request
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "weak",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	mockService.AssertExpectations(t)
}

// Test: Register with duplicate email returns 409
func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Mock service to return email taken error
	mockService.On("Register", mock.Anything, "existing@example.com", "password123").
		Return(nil, models.ErrEmailTaken)

	// Create request
	reqBody := map[string]string{
		"email":    "existing@example.com",
		"password": "password123",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusConflict, rr.Code)

	// Parse response
	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Email already taken", response["error"])

	mockService.AssertExpectations(t)
}

// ============================================
// Login Tests
// ============================================

// Test: Login with valid credentials returns 200 with token
func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Mock service to return token
	mockService.On("Login", mock.Anything, "test@example.com", "password123").
		Return("test-jwt-token", nil)

	// Create request
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Login(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Assert response structure
	assert.Contains(t, response, "token")

	token, ok := response["token"].(string)
	require.True(t, ok)
	assert.Equal(t, "test-jwt-token", token)

	mockService.AssertExpectations(t)
}

// Test: Login with invalid JSON returns 400
func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Login(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Parse response
	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

// Test: Login with invalid credentials returns 401
func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Mock service to return invalid credentials error
	mockService.On("Login", mock.Anything, "test@example.com", "wrongpassword").
		Return("", models.ErrInvalidCredentials)

	// Create request
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Login(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Parse response
	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid credentials", response["error"])

	mockService.AssertExpectations(t)
}

// Test: Login with missing fields returns 400
func TestAuthHandler_Login_MissingFields(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Create request with missing password
	reqBody := map[string]string{
		"email": "test@example.com",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Login(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
