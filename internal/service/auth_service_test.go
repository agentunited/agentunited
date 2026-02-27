package service

import (
	"context"
	"testing"

	"github.com/agentunited/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// Test: Register with valid input should create user
func TestAuthService_Register_ValidInput(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret-key")
	ctx := context.Background()

	email := "test@example.com"
	password := "password123"

	// Mock: GetByEmail returns not found (email not taken)
	mockRepo.On("GetByEmail", ctx, email).Return(nil, models.ErrUserNotFound)

	// Mock: Create succeeds and sets ID
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			user := args.Get(1).(*models.User)
			user.ID = "test-uuid-1234"
		}).
		Return(nil)

	user, err := service.Register(ctx, email, password)

	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.NotEmpty(t, user.ID)
	assert.Empty(t, user.PasswordHash, "Password hash should not be returned")

	mockRepo.AssertExpectations(t)
}

// Test: Register with invalid email format should fail
func TestAuthService_Register_InvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret-key")
	ctx := context.Background()

	invalidEmails := []string{
		"notanemail",
		"missing@domain",
		"@nodomain.com",
		"spaces in@email.com",
		"",
	}

	for _, email := range invalidEmails {
		_, err := service.Register(ctx, email, "password123")
		assert.ErrorIs(t, err, models.ErrInvalidEmail, "Email: %s should be invalid", email)
	}
}

// Test: Register with weak password should fail
func TestAuthService_Register_WeakPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret-key")
	ctx := context.Background()

	weakPasswords := []string{
		"short",           // Too short
		"12345678",        // No letters
		"abcdefgh",        // No numbers
		"",                // Empty
		"abc123",          // Too short
	}

	for _, password := range weakPasswords {
		_, err := service.Register(ctx, "test@example.com", password)
		assert.ErrorIs(t, err, models.ErrWeakPassword, "Password: %s should be weak", password)
	}
}

// Test: Register with duplicate email should fail
func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret-key")
	ctx := context.Background()

	email := "existing@example.com"

	// Mock: GetByEmail returns existing user
	existingUser := &models.User{
		ID:    "existing-id",
		Email: email,
	}
	mockRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	_, err := service.Register(ctx, email, "password123")

	assert.ErrorIs(t, err, models.ErrEmailTaken)
	mockRepo.AssertExpectations(t)
}

// Test: Register should hash password with bcrypt
func TestAuthService_Register_HashesPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret-key")
	ctx := context.Background()

	email := "test@example.com"
	password := "password123"

	// Mock: GetByEmail returns not found
	mockRepo.On("GetByEmail", ctx, email).Return(nil, models.ErrUserNotFound)

	// Mock: Create succeeds - capture the password hash and set ID
	var capturedPasswordHash string
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			user := args.Get(1).(*models.User)
			capturedPasswordHash = user.PasswordHash // Capture hash before it gets cleared
			user.ID = "test-uuid-5678"
		}).
		Return(nil)

	_, err := service.Register(ctx, email, password)
	require.NoError(t, err)

	// Verify password was hashed
	assert.NotEmpty(t, capturedPasswordHash)
	assert.NotEqual(t, password, capturedPasswordHash, "Password should be hashed")

	// Verify hash is valid bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(capturedPasswordHash), []byte(password))
	assert.NoError(t, err, "Password hash should be valid bcrypt")
}

// ============================================
// Login Tests
// ============================================

// Test: Login with valid credentials should return JWT token
func TestAuthService_Login_ValidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret-key")
	ctx := context.Background()

	email := "test@example.com"
	password := "password123"

	// Create hashed password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)

	existingUser := &models.User{
		ID:           "test-uuid-1234",
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	// Mock: GetByEmail returns existing user
	mockRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	token, err := service.Login(ctx, email, password)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	mockRepo.AssertExpectations(t)
}

// Test: Login with invalid password should fail
func TestAuthService_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret-key")
	ctx := context.Background()

	email := "test@example.com"
	correctPassword := "password123"
	wrongPassword := "wrongpassword"

	// Create hashed password for correct password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), BcryptCost)

	existingUser := &models.User{
		ID:           "test-uuid-1234",
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	// Mock: GetByEmail returns existing user
	mockRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	_, err := service.Login(ctx, email, wrongPassword)

	assert.ErrorIs(t, err, models.ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}

// Test: Login with non-existent user should fail
func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret-key")
	ctx := context.Background()

	email := "nonexistent@example.com"
	password := "password123"

	// Mock: GetByEmail returns not found
	mockRepo.On("GetByEmail", ctx, email).Return(nil, models.ErrUserNotFound)

	_, err := service.Login(ctx, email, password)

	// Should return ErrInvalidCredentials (don't leak whether user exists)
	assert.ErrorIs(t, err, models.ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}

// Test: Login JWT token should contain user ID and email
func TestAuthService_Login_TokenContainsClaims(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtSecret := "test-secret-key-for-parsing"
	service := NewAuthService(mockRepo, jwtSecret)
	ctx := context.Background()

	email := "test@example.com"
	password := "password123"
	userID := "test-uuid-1234"

	// Create hashed password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)

	existingUser := &models.User{
		ID:           userID,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	// Mock: GetByEmail returns existing user
	mockRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	tokenString, err := service.Login(ctx, email, password)
	require.NoError(t, err)

	// Parse token to verify claims
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	require.NoError(t, err)
	require.True(t, token.Valid)

	claims, ok := token.Claims.(*JWTClaims)
	require.True(t, ok)

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.False(t, claims.ExpiresAt.IsZero(), "Token should have expiry")
}
