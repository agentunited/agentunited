package service

import (
	"context"
	"testing"

	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMessageRepository is a mock implementation of MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, message *models.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByChannel(ctx context.Context, channelID string, limit int, before string) ([]*models.Message, bool, error) {
	args := m.Called(ctx, channelID, limit, before)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).([]*models.Message), args.Bool(1), args.Error(2)
}

func (m *MockMessageRepository) GetByID(ctx context.Context, messageID string) (*models.Message, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) Update(ctx context.Context, messageID, text string) (*models.Message, error) {
	args := m.Called(ctx, messageID, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) Delete(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockMessageRepository) Search(ctx context.Context, query string, channelID string, limit int) ([]*models.Message, error) {
	args := m.Called(ctx, query, channelID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

// Test: Send message with valid input
func TestMessageService_Send_ValidInput(t *testing.T) {
	mockMsgRepo := new(MockMessageRepository)
	mockChannelRepo := new(MockChannelRepository)
	service := NewMessageService(mockMsgRepo, mockChannelRepo)
	ctx := context.Background()

	userID := "test-user-id"
	channelID := "test-channel-id"
	text := "Hello world"

	// Mock: User is member of channel
	mockChannelRepo.On("IsMember", ctx, channelID, userID).Return(true, "member", nil)

	// Mock: Create message
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*models.Message")).
		Run(func(args mock.Arguments) {
			msg := args.Get(1).(*models.Message)
			msg.ID = "test-message-id"
		}).
		Return(nil)

	message, err := service.Send(ctx, channelID, userID, text)

	require.NoError(t, err)
	assert.Equal(t, channelID, message.ChannelID)
	assert.Equal(t, userID, message.AuthorID)
	assert.Equal(t, "user", message.AuthorType)
	assert.Equal(t, text, message.Text)
	assert.NotEmpty(t, message.ID)

	mockChannelRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

// Test: Send message when user is not a member
func TestMessageService_Send_NotMember(t *testing.T) {
	mockMsgRepo := new(MockMessageRepository)
	mockChannelRepo := new(MockChannelRepository)
	service := NewMessageService(mockMsgRepo, mockChannelRepo)
	ctx := context.Background()

	userID := "test-user-id"
	channelID := "test-channel-id"
	text := "Hello world"

	// Mock: User is NOT member
	mockChannelRepo.On("IsMember", ctx, channelID, userID).Return(false, "", nil)

	_, err := service.Send(ctx, channelID, userID, text)

	assert.ErrorIs(t, err, models.ErrNotChannelMember)
	mockChannelRepo.AssertExpectations(t)
}

// Test: Send message with invalid text (empty)
func TestMessageService_Send_EmptyText(t *testing.T) {
	mockMsgRepo := new(MockMessageRepository)
	mockChannelRepo := new(MockChannelRepository)
	service := NewMessageService(mockMsgRepo, mockChannelRepo)
	ctx := context.Background()

	_, err := service.Send(ctx, "channel-id", "user-id", "")

	assert.ErrorIs(t, err, models.ErrInvalidMessageText)
}

// Test: Send message with text too long
func TestMessageService_Send_TextTooLong(t *testing.T) {
	mockMsgRepo := new(MockMessageRepository)
	mockChannelRepo := new(MockChannelRepository)
	service := NewMessageService(mockMsgRepo, mockChannelRepo)
	ctx := context.Background()

	// Create text longer than 10,000 characters
	longText := string(make([]byte, 10001))
	for i := range longText {
		longText = string(append([]byte(longText[:i]), 'a'))
		if i >= 10000 {
			break
		}
	}

	_, err := service.Send(ctx, "channel-id", "user-id", longText)

	assert.ErrorIs(t, err, models.ErrInvalidMessageText)
}

// Test: Get messages for channel
func TestMessageService_GetMessages(t *testing.T) {
	mockMsgRepo := new(MockMessageRepository)
	mockChannelRepo := new(MockChannelRepository)
	service := NewMessageService(mockMsgRepo, mockChannelRepo)
	ctx := context.Background()

	userID := "test-user-id"
	channelID := "test-channel-id"
	limit := 50
	before := ""

	messages := []*models.Message{
		{ID: "msg1", Text: "Message 1"},
		{ID: "msg2", Text: "Message 2"},
	}

	// Mock: User is member
	mockChannelRepo.On("IsMember", ctx, channelID, userID).Return(true, "member", nil)

	// Mock: GetByChannel
	mockMsgRepo.On("GetByChannel", ctx, channelID, limit, before).
		Return(messages, false, nil)

	result, err := service.GetMessages(ctx, channelID, userID, limit, before)

	require.NoError(t, err)
	assert.Len(t, result.Messages, 2)
	assert.False(t, result.HasMore)
	assert.Equal(t, "Message 1", result.Messages[0].Text)

	mockChannelRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

// Test: Get messages when user is not a member
func TestMessageService_GetMessages_NotMember(t *testing.T) {
	mockMsgRepo := new(MockMessageRepository)
	mockChannelRepo := new(MockChannelRepository)
	service := NewMessageService(mockMsgRepo, mockChannelRepo)
	ctx := context.Background()

	userID := "test-user-id"
	channelID := "test-channel-id"

	// Mock: User is NOT member
	mockChannelRepo.On("IsMember", ctx, channelID, userID).Return(false, "", nil)

	_, err := service.GetMessages(ctx, channelID, userID, 50, "")

	assert.ErrorIs(t, err, models.ErrNotChannelMember)
	mockChannelRepo.AssertExpectations(t)
}

// Test: Get messages with pagination
func TestMessageService_GetMessages_WithPagination(t *testing.T) {
	mockMsgRepo := new(MockMessageRepository)
	mockChannelRepo := new(MockChannelRepository)
	service := NewMessageService(mockMsgRepo, mockChannelRepo)
	ctx := context.Background()

	userID := "test-user-id"
	channelID := "test-channel-id"
	limit := 2
	before := "msg2-id"

	messages := []*models.Message{
		{ID: "msg3", Text: "Message 3"},
		{ID: "msg4", Text: "Message 4"},
	}

	// Mock: User is member
	mockChannelRepo.On("IsMember", ctx, channelID, userID).Return(true, "member", nil)

	// Mock: GetByChannel with cursor
	mockMsgRepo.On("GetByChannel", ctx, channelID, limit, before).
		Return(messages, true, nil)

	result, err := service.GetMessages(ctx, channelID, userID, limit, before)

	require.NoError(t, err)
	assert.Len(t, result.Messages, 2)
	assert.True(t, result.HasMore)

	mockChannelRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}
