package service

import (
	"context"
	"testing"

	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockChannelRepository is a mock implementation of ChannelRepository
type MockChannelRepository struct {
	mock.Mock
}

func (m *MockChannelRepository) Create(ctx context.Context, channel *models.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *MockChannelRepository) GetByID(ctx context.Context, id string) (*models.Channel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *MockChannelRepository) ListByUser(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ChannelWithMembers), args.Error(1)
}

func (m *MockChannelRepository) GetMembers(ctx context.Context, channelID string) ([]*models.MemberInfo, error) {
	args := m.Called(ctx, channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MemberInfo), args.Error(1)
}

func (m *MockChannelRepository) IsMember(ctx context.Context, channelID, userID string) (bool, string, error) {
	args := m.Called(ctx, channelID, userID)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockChannelRepository) AddMember(ctx context.Context, channelID, userID, role string) error {
	args := m.Called(ctx, channelID, userID, role)
	return args.Error(0)
}

func (m *MockChannelRepository) Update(ctx context.Context, channelID, name, topic string) (*models.Channel, error) {
	args := m.Called(ctx, channelID, name, topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *MockChannelRepository) Delete(ctx context.Context, channelID string) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *MockChannelRepository) RemoveMember(ctx context.Context, channelID, userID string) error {
	args := m.Called(ctx, channelID, userID)
	return args.Error(0)
}

func (m *MockChannelRepository) ListDMChannels(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ChannelWithMembers), args.Error(1)
}

func (m *MockChannelRepository) GetOrCreateDMChannel(ctx context.Context, user1ID, user2ID string) (*models.Channel, error) {
	args := m.Called(ctx, user1ID, user2ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *MockChannelRepository) MarkChannelRead(ctx context.Context, userID, channelID string) error {
	args := m.Called(ctx, userID, channelID)
	return args.Error(0)
}

func (m *MockChannelRepository) GetUnreadCounts(ctx context.Context, userID string) (map[string]int, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

// Test: Create channel with valid name
func TestChannelService_Create_ValidName(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	service := NewChannelService(mockRepo)
	ctx := context.Background()

	userID := "test-user-id"
	name := "general"
	topic := "General discussion"

	// Mock repository Create
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Channel")).
		Run(func(args mock.Arguments) {
			ch := args.Get(1).(*models.Channel)
			ch.ID = "test-channel-id"
		}).
		Return(nil)

	channel, err := service.Create(ctx, userID, name, topic)

	require.NoError(t, err)
	assert.Equal(t, name, channel.Name)
	assert.Equal(t, topic, channel.Topic)
	assert.Equal(t, userID, channel.CreatedBy)
	assert.NotEmpty(t, channel.ID)

	mockRepo.AssertExpectations(t)
}

// Test: Create channel with invalid names (not lowercase, alphanumeric, hyphens)
func TestChannelService_Create_InvalidName(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	service := NewChannelService(mockRepo)
	ctx := context.Background()

	invalidNames := []string{
		"General",                           // Uppercase
		"test channel",                      // Space
		"test_channel",                      // Underscore
		"test.channel",                      // Dot
		"test@channel",                      // Special char
		"",                                  // Empty
		"a",                                 // Too short (< 2 chars)
		"123456789012345678901234567890123", // Too long (> 32 chars)
	}

	for _, name := range invalidNames {
		_, err := service.Create(ctx, "user-id", name, "Topic")
		assert.ErrorIs(t, err, models.ErrInvalidChannelName, "Name '%s' should be invalid", name)
	}
}

// Test: Create channel with duplicate name
func TestChannelService_Create_DuplicateName(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	service := NewChannelService(mockRepo)
	ctx := context.Background()

	// Mock repository to return duplicate name error
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Channel")).
		Return(models.ErrChannelNameTaken)

	_, err := service.Create(ctx, "user-id", "general", "Topic")

	assert.ErrorIs(t, err, models.ErrChannelNameTaken)
	mockRepo.AssertExpectations(t)
}

// Test: List channels for user
func TestChannelService_List(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	service := NewChannelService(mockRepo)
	ctx := context.Background()

	userID := "test-user-id"
	expectedChannels := []*models.ChannelWithMembers{
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

	mockRepo.On("ListByUser", ctx, userID).Return(expectedChannels, nil)
	mockRepo.On("ListDMChannels", ctx, userID).Return([]*models.ChannelWithMembers{}, nil)

	channels, err := service.List(ctx, userID)

	require.NoError(t, err)
	assert.Len(t, channels, 2)
	assert.Equal(t, "general", channels[0].Name)
	assert.Equal(t, 5, channels[0].MemberCount)

	mockRepo.AssertExpectations(t)
}

// Test: Get channel by ID with members
func TestChannelService_Get(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	service := NewChannelService(mockRepo)
	ctx := context.Background()

	channelID := "test-channel-id"
	userID := "test-user-id"

	channel := &models.Channel{
		ID:   channelID,
		Name: "general",
	}

	members := []*models.MemberInfo{
		{ID: "user1", Email: "user1@example.com", Role: "owner"},
		{ID: "user2", Email: "user2@example.com", Role: "member"},
	}

	// Mock IsMember check (user is member)
	mockRepo.On("IsMember", ctx, channelID, userID).Return(true, "owner", nil)

	// Mock GetByID
	mockRepo.On("GetByID", ctx, channelID).Return(channel, nil)

	// Mock GetMembers
	mockRepo.On("GetMembers", ctx, channelID).Return(members, nil)

	result, err := service.Get(ctx, channelID, userID)

	require.NoError(t, err)
	assert.Equal(t, channelID, result.ID)
	assert.Len(t, result.Members, 2)
	assert.Equal(t, "owner", result.Members[0].Role)

	mockRepo.AssertExpectations(t)
}

// Test: Get channel when user is not a member
func TestChannelService_Get_NotMember(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	service := NewChannelService(mockRepo)
	ctx := context.Background()

	channelID := "test-channel-id"
	userID := "test-user-id"

	// Mock IsMember check (user is NOT member)
	mockRepo.On("IsMember", ctx, channelID, userID).Return(false, "", nil)

	_, err := service.Get(ctx, channelID, userID)

	assert.ErrorIs(t, err, models.ErrNotChannelMember)
	mockRepo.AssertExpectations(t)
}

// Test: Check if user is member
func TestChannelService_IsMember(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	service := NewChannelService(mockRepo)
	ctx := context.Background()

	channelID := "test-channel-id"
	userID := "test-user-id"

	mockRepo.On("IsMember", ctx, channelID, userID).Return(true, "member", nil)

	isMember, role, err := service.IsMember(ctx, channelID, userID)

	require.NoError(t, err)
	assert.True(t, isMember)
	assert.Equal(t, "member", role)

	mockRepo.AssertExpectations(t)
}
