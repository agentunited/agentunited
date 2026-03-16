package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func hashTestToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

func TestInviteService_ValidateInvite_HappyPath(t *testing.T) {
	userRepo := &mockUserRepository{}
	inviteRepo := &mockInviteRepository{}
	subscriptionRepo := &mockSubscriptionRepository{}
	service := NewInviteService(userRepo, inviteRepo, subscriptionRepo, nil, nil, nil, "test-jwt-secret", "http://localhost:3001")

	ctx := context.Background()

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

	expectedHash := hashTestToken("test-token")
	inviteRepo.On("ValidateToken", ctx, expectedHash).Return(invite, nil)
	userRepo.On("GetByID", ctx, "user-456").Return(user, nil)

	resultInvite, resultUser, err := service.ValidateInvite(ctx, "test-token")
	require.NoError(t, err)

	assert.Equal(t, invite.ID, resultInvite.ID)
	assert.Equal(t, user.Email, resultUser.Email)

	inviteRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestInviteService_ValidateInvite_InvalidToken(t *testing.T) {
	userRepo := &mockUserRepository{}
	inviteRepo := &mockInviteRepository{}
	subscriptionRepo := &mockSubscriptionRepository{}
	service := NewInviteService(userRepo, inviteRepo, subscriptionRepo, nil, nil, nil, "test-jwt-secret", "http://localhost:3001")

	ctx := context.Background()

	expectedHash := hashTestToken("invalid-token")
	inviteRepo.On("ValidateToken", ctx, expectedHash).Return(nil, models.ErrInviteNotFound)

	_, _, err := service.ValidateInvite(ctx, "invalid-token")
	require.Error(t, err)
	assert.Equal(t, models.ErrInviteNotFound, err)

	inviteRepo.AssertExpectations(t)
}

func TestInviteService_AcceptInvite_HappyPath(t *testing.T) {
	userRepo := &mockUserRepository{}
	inviteRepo := &mockInviteRepository{}
	subscriptionRepo := &mockSubscriptionRepository{}
	service := NewInviteService(userRepo, inviteRepo, subscriptionRepo, nil, nil, nil, "test-jwt-secret", "http://localhost:3001")

	ctx := context.Background()

	invite := &models.Invite{
		ID:        "invite-123",
		UserID:    "user-456",
		Status:    models.InviteStatusPending,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	user := &models.User{
		ID:           "user-456",
		Email:        "test@example.com",
		PasswordHash: "", // Empty before invite acceptance
	}

	expectedHash := hashTestToken("test-token")
	inviteRepo.On("ValidateToken", ctx, expectedHash).Return(invite, nil)
	userRepo.On("GetByID", ctx, "user-456").Return(user, nil)

	// Mock password update
	userRepo.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == "user-456" && u.PasswordHash != "" && u.DisplayName == "Siinn"
	})).Return(nil)

	// Mock token consumption
	inviteRepo.On("ConsumeToken", ctx, expectedHash).Return(nil)

	jwtToken, dmChannelID, err := service.AcceptInvite(ctx, "test-token", "securepassword123", "Siinn")
	require.NoError(t, err)
	assert.NotEmpty(t, jwtToken)
	assert.Empty(t, dmChannelID)

	inviteRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestInviteService_AcceptInvite_InvalidToken(t *testing.T) {
	userRepo := &mockUserRepository{}
	inviteRepo := &mockInviteRepository{}
	subscriptionRepo := &mockSubscriptionRepository{}
	service := NewInviteService(userRepo, inviteRepo, subscriptionRepo, nil, nil, nil, "test-jwt-secret", "http://localhost:3001")

	ctx := context.Background()

	expectedHash := hashTestToken("invalid-token")
	inviteRepo.On("ValidateToken", ctx, expectedHash).Return(nil, models.ErrInviteNotFound)

	_, _, err := service.AcceptInvite(ctx, "invalid-token", "securepassword123", "")
	require.Error(t, err)
	assert.Equal(t, models.ErrInviteNotFound, err)

	inviteRepo.AssertExpectations(t)
}

func TestInviteService_AcceptInvite_WeakPassword(t *testing.T) {
	userRepo := &mockUserRepository{}
	inviteRepo := &mockInviteRepository{}
	subscriptionRepo := &mockSubscriptionRepository{}
	service := NewInviteService(userRepo, inviteRepo, subscriptionRepo, nil, nil, nil, "test-jwt-secret", "http://localhost:3001")

	ctx := context.Background()

	_, _, err := service.AcceptInvite(ctx, "test-token", "weak", "")
	require.Error(t, err)
	assert.Equal(t, models.ErrWeakPassword, err)

	// No repository calls should be made for weak password
	inviteRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}
