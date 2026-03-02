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
	SendMessageWithAttachment(ctx context.Context, message *models.Message, agentCtx *AgentContext) (*models.Message, error)
	GetMessages(ctx context.Context, channelID, userID string, limit int, before string) (*models.MessageList, error)
	EditMessage(ctx context.Context, messageID, userID, text string) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageID, userID string) error
	SearchMessages(ctx context.Context, query, channelID, userID string, limit int) ([]*models.Message, error)
}

// messageService implements MessageService
type messageService struct {
	messageRepo repository.MessageRepository
	channelRepo repository.ChannelRepository
	agentRepo   repository.AgentRepository
}

// NewMessageService creates a new message service
func NewMessageService(messageRepo repository.MessageRepository, channelRepo repository.ChannelRepository, agentRepo ...repository.AgentRepository) MessageService {
	svc := &messageService{
		messageRepo: messageRepo,
		channelRepo: channelRepo,
	}
	if len(agentRepo) > 0 {
		svc.agentRepo = agentRepo[0]
	}
	return svc
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

// SendMessageWithAttachment sends a message with optional file attachment
func (s *messageService) SendMessageWithAttachment(ctx context.Context, message *models.Message, agentCtx *AgentContext) (*models.Message, error) {
	// Check if user is a member of the channel
	isMember, _, err := s.channelRepo.IsMember(ctx, message.ChannelID, message.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, models.ErrNotChannelMember
	}

	// Validate message text (if provided)
	if message.Text != "" && !isValidMessageText(message.Text) {
		return nil, models.ErrInvalidMessageText
	}

	// At least one of text or attachment must be provided
	if message.Text == "" && message.AttachmentURL == "" {
		return nil, models.ErrInvalidMessageText
	}

	// Set timestamps
	message.CreatedAt = time.Now()
	message.UpdatedAt = message.CreatedAt

	// Override author info if sending as agent
	if agentCtx != nil {
		message.AuthorID = agentCtx.AgentID
		message.AuthorType = "agent"
		if agentCtx.DisplayName != "" {
			message.AuthorEmail = agentCtx.DisplayName
		}
	}

	// Create the message
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

// EditMessage updates a message's text (author only)
func (s *messageService) EditMessage(ctx context.Context, messageID, userID, text string) (*models.Message, error) {
	// Validate message text
	if !isValidMessageText(text) {
		return nil, models.ErrInvalidMessageText
	}

	// Get the existing message to check authorship
	existingMessage, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("get message: %w", err)
	}

	// Check if user is the author or the agent's owner
	isAuthor := existingMessage.AuthorID == userID && existingMessage.AuthorType == "user"
	if !isAuthor && existingMessage.AuthorType == "agent" {
		isAuthor = s.isAgentOwner(ctx, existingMessage.AuthorID, userID)
	}
	if !isAuthor {
		return nil, models.ErrUnauthorizedMessageEdit
	}

	// Check if user is member of the channel
	isMember, _, err := s.channelRepo.IsMember(ctx, existingMessage.ChannelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, models.ErrNotChannelMember
	}

	// Update the message
	message, err := s.messageRepo.Update(ctx, messageID, text)
	if err != nil {
		return nil, fmt.Errorf("update message: %w", err)
	}

	// Get author email for the response
	if message.AuthorEmail == "" {
		if existingMessage.AuthorEmail != "" {
			message.AuthorEmail = existingMessage.AuthorEmail
		}
	}

	return message, nil
}

// DeleteMessage removes a message (author only)
func (s *messageService) DeleteMessage(ctx context.Context, messageID, userID string) error {
	// Get the existing message to check authorship
	existingMessage, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}

	// Check if user is the author or the agent's owner
	isAuthor := existingMessage.AuthorID == userID && existingMessage.AuthorType == "user"
	if !isAuthor && existingMessage.AuthorType == "agent" {
		isAuthor = s.isAgentOwner(ctx, existingMessage.AuthorID, userID)
	}
	if !isAuthor {
		return models.ErrUnauthorizedMessageDelete
	}

	// Check if user is member of the channel
	isMember, _, err := s.channelRepo.IsMember(ctx, existingMessage.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return models.ErrNotChannelMember
	}

	// Delete the message
	if err := s.messageRepo.Delete(ctx, messageID); err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	return nil
}

// SearchMessages performs full-text search across messages
func (s *messageService) SearchMessages(ctx context.Context, query, channelID, userID string, limit int) ([]*models.Message, error) {
	// Validate query
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Set default limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// If channel specified, check membership
	if channelID != "" {
		isMember, _, err := s.channelRepo.IsMember(ctx, channelID, userID)
		if err != nil {
			return nil, fmt.Errorf("check membership: %w", err)
		}
		if !isMember {
			return nil, models.ErrNotChannelMember
		}
	} else {
		// If no channel specified, only search channels where user is member
		// This is more complex - for now we require channelID
		// TODO: Implement cross-channel search with membership filtering
		return nil, fmt.Errorf("channel_id is required for search")
	}

	// Perform search
	messages, err := s.messageRepo.Search(ctx, query, channelID, limit)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}

	return messages, nil
}

// isValidMessageText checks if message text is valid
// Requirements: 1-10,000 characters
func isValidMessageText(text string) bool {
	length := len(text)
	return length > 0 && length <= 10000
}

// isAgentOwner checks if the given user owns the given agent
func (s *messageService) isAgentOwner(ctx context.Context, agentID, userID string) bool {
	if s.agentRepo == nil {
		return false
	}
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return false
	}
	return agent.OwnerID == userID
}
