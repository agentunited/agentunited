package service

import (
	"context"
	"fmt"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
)

// AgentContext holds agent identity for agent-authenticated requests
type AgentContext struct {
	AgentID     string
	DisplayName string
}

// MessageService handles message business logic
type MessageService interface {
	Send(ctx context.Context, channelID, userID, text string) (*models.Message, error)
	SendAsAgent(ctx context.Context, channelID, ownerID string, agent AgentContext, text string) (*models.Message, error)
	GetMessages(ctx context.Context, channelID, userID string, limit int, before string) (*models.MessageList, error)
}

// messageService implements MessageService
type messageService struct {
	messageRepo repository.MessageRepository
	channelRepo repository.ChannelRepository
}

// NewMessageService creates a new message service
func NewMessageService(messageRepo repository.MessageRepository, channelRepo repository.ChannelRepository) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		channelRepo: channelRepo,
	}
}

// Send creates and saves a new message
func (s *messageService) Send(ctx context.Context, channelID, userID, text string) (*models.Message, error) {
	// Validate text length
	if !isValidMessageText(text) {
		return nil, models.ErrInvalidMessageText
	}

	// Check if user is a member of the channel
	isMember, _, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, models.ErrNotChannelMember
	}

	// Create message
	message := &models.Message{
		ChannelID:  channelID,
		AuthorID:   userID,
		AuthorType: "user",
		Text:       text,
		CreatedAt:  time.Now(),
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	return message, nil
}

// SendAsAgent creates a message authored by an agent
func (s *messageService) SendAsAgent(ctx context.Context, channelID, ownerID string, agent AgentContext, text string) (*models.Message, error) {
	if !isValidMessageText(text) {
		return nil, models.ErrInvalidMessageText
	}

	// Check channel membership using owner's user ID
	isMember, _, err := s.channelRepo.IsMember(ctx, channelID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, models.ErrNotChannelMember
	}

	message := &models.Message{
		ChannelID:  channelID,
		AuthorID:   agent.AgentID,
		AuthorType: "agent",
		Text:       text,
		CreatedAt:  time.Now(),
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	return message, nil
}

// GetMessages retrieves messages for a channel with pagination
func (s *messageService) GetMessages(ctx context.Context, channelID, userID string, limit int, before string) (*models.MessageList, error) {
	// Check if user is a member of the channel
	isMember, _, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, models.ErrNotChannelMember
	}

	// Enforce limit constraints (max 100)
	if limit <= 0 {
		limit = 50 // Default
	}
	if limit > 100 {
		limit = 100 // Max
	}

	// Get messages
	messages, hasMore, err := s.messageRepo.GetByChannel(ctx, channelID, limit, before)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	return &models.MessageList{
		Messages: messages,
		HasMore:  hasMore,
	}, nil
}

// isValidMessageText checks if message text is valid
// Requirements: 1-10,000 characters
func isValidMessageText(text string) bool {
	length := len(text)
	return length > 0 && length <= 10000
}
