package repository

import (
	"context"
	"testing"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewChannelRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user first
	user := &models.User{
		Email:        "channelcreator@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create channel
	channel := &models.Channel{
		Name:      "general",
		Topic:     "General discussion",
		CreatedBy: user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, channel)
	require.NoError(t, err)

	// Verify ID was assigned
	assert.NotEmpty(t, channel.ID)
	assert.False(t, channel.CreatedAt.IsZero())

	// Verify creator was added as owner
	isMember, role, err := repo.IsMember(ctx, channel.ID, user.ID)
	require.NoError(t, err)
	assert.True(t, isMember)
	assert.Equal(t, "owner", role)
}

func TestChannelRepository_Create_DuplicateName(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewChannelRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create user
	user := &models.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create first channel
	channel1 := &models.Channel{
		Name:      "general",
		Topic:     "First",
		CreatedBy: user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, channel1)
	require.NoError(t, err)

	// Try to create second channel with same name
	channel2 := &models.Channel{
		Name:      "general",
		Topic:     "Second",
		CreatedBy: user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, channel2)

	// Should return ErrChannelNameTaken
	assert.ErrorIs(t, err, models.ErrChannelNameTaken)
}

func TestChannelRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewChannelRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create user and channel
	user := &models.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	originalChannel := &models.Channel{
		Name:      "random",
		Topic:     "Random chat",
		CreatedBy: user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, originalChannel)
	require.NoError(t, err)

	// Get channel by ID
	retrievedChannel, err := repo.GetByID(ctx, originalChannel.ID)
	require.NoError(t, err)

	assert.Equal(t, originalChannel.ID, retrievedChannel.ID)
	assert.Equal(t, originalChannel.Name, retrievedChannel.Name)
	assert.Equal(t, originalChannel.Topic, retrievedChannel.Topic)
	assert.Equal(t, originalChannel.CreatedBy, retrievedChannel.CreatedBy)
}

func TestChannelRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewChannelRepository(db)
	ctx := context.Background()

	// Try to get non-existent channel
	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")

	assert.ErrorIs(t, err, models.ErrChannelNotFound)
}

func TestChannelRepository_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewChannelRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create two users
	user1 := &models.User{
		Email:        "user1@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user1)
	require.NoError(t, err)

	user2 := &models.User{
		Email:        "user2@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	// User1 creates two channels
	channel1 := &models.Channel{
		Name:      "general",
		Topic:     "General",
		CreatedBy: user1.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, channel1)
	require.NoError(t, err)

	channel2 := &models.Channel{
		Name:      "random",
		Topic:     "Random",
		CreatedBy: user1.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, channel2)
	require.NoError(t, err)

	// User2 creates one channel
	channel3 := &models.Channel{
		Name:      "off-topic",
		Topic:     "Off topic",
		CreatedBy: user2.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, channel3)
	require.NoError(t, err)

	// List channels for user1
	channels, err := repo.ListByUser(ctx, user1.ID)
	require.NoError(t, err)

	// User1 should see 2 channels (they created and auto-joined)
	assert.Len(t, channels, 2)

	// Verify member counts are included
	for _, ch := range channels {
		assert.Greater(t, ch.MemberCount, 0)
	}
}

func TestChannelRepository_GetMembers(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewChannelRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create two users
	user1 := &models.User{
		Email:        "owner@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user1)
	require.NoError(t, err)

	user2 := &models.User{
		Email:        "member@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	// Create channel (user1 is owner)
	channel := &models.Channel{
		Name:      "general",
		Topic:     "General",
		CreatedBy: user1.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, channel)
	require.NoError(t, err)

	// Add user2 as member
	err = repo.AddMember(ctx, channel.ID, user2.ID, "member")
	require.NoError(t, err)

	// Get members
	members, err := repo.GetMembers(ctx, channel.ID)
	require.NoError(t, err)

	// Should have 2 members
	assert.Len(t, members, 2)

	// Verify roles
	var ownerFound, memberFound bool
	for _, m := range members {
		if m.Email == "owner@example.com" {
			assert.Equal(t, "owner", m.Role)
			ownerFound = true
		}
		if m.Email == "member@example.com" {
			assert.Equal(t, "member", m.Role)
			memberFound = true
		}
	}
	assert.True(t, ownerFound, "Owner should be in members list")
	assert.True(t, memberFound, "Member should be in members list")
}

func TestChannelRepository_IsMember(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewChannelRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create user and channel
	user := &models.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	channel := &models.Channel{
		Name:      "general",
		Topic:     "General",
		CreatedBy: user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, channel)
	require.NoError(t, err)

	// Check membership (should be member as creator)
	isMember, role, err := repo.IsMember(ctx, channel.ID, user.ID)
	require.NoError(t, err)

	assert.True(t, isMember)
	assert.Equal(t, "owner", role)

	// Check non-member
	isMember, role, err = repo.IsMember(ctx, channel.ID, "00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	assert.False(t, isMember)
	assert.Empty(t, role)
}

func TestChannelRepository_AddMember_DuplicateMember(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewChannelRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()

	// Create user and channel
	user := &models.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	channel := &models.Channel{
		Name:      "general",
		Topic:     "General",
		CreatedBy: user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, channel)
	require.NoError(t, err)

	// Try to add user again (already added as creator)
	err = repo.AddMember(ctx, channel.ID, user.ID, "member")

	// Should return ErrAlreadyChannelMember
	assert.ErrorIs(t, err, models.ErrAlreadyChannelMember)
}
