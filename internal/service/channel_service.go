package service

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
)

// channelNameRegex validates channel names: lowercase alphanumeric + hyphens, 2-32 chars
var channelNameRegex = regexp.MustCompile(`^[a-z0-9-]{2,32}$`)

// ChannelService handles channel business logic
type ChannelService interface {
	Create(ctx context.Context, userID, name, topic string) (*models.Channel, error)
	Get(ctx context.Context, channelID, userID string) (*models.ChannelWithMembers, error)
	List(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error)
	IsMember(ctx context.Context, channelID, userID string) (bool, string, error)
}

// channelService implements ChannelService
type channelService struct {
	channelRepo repository.ChannelRepository
}

// NewChannelService creates a new channel service
func NewChannelService(channelRepo repository.ChannelRepository) ChannelService {
	return &channelService{
		channelRepo: channelRepo,
	}
}

// Create creates a new channel with the user as owner
func (s *channelService) Create(ctx context.Context, userID, name, topic string) (*models.Channel, error) {
	// Validate channel name
	if !isValidChannelName(name) {
		return nil, models.ErrInvalidChannelName
	}

	// Create channel
	channel := &models.Channel{
		Name:      name,
		Topic:     topic,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	return channel, nil
}

// Get retrieves a channel with members if the user is a member
func (s *channelService) Get(ctx context.Context, channelID, userID string) (*models.ChannelWithMembers, error) {
	// Check if user is a member
	isMember, _, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, models.ErrNotChannelMember
	}

	// Get channel
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}

	// Get members
	members, err := s.channelRepo.GetMembers(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	// Construct result
	result := &models.ChannelWithMembers{
		Channel: *channel,
		Members: members,
	}

	return result, nil
}

// List retrieves all channels where the user is a member
func (s *channelService) List(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	channels, err := s.channelRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}

	return channels, nil
}

// IsMember checks if a user is a member of a channel
func (s *channelService) IsMember(ctx context.Context, channelID, userID string) (bool, string, error) {
	isMember, role, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return false, "", fmt.Errorf("check membership: %w", err)
	}

	return isMember, role, nil
}

// isValidChannelName checks if a channel name is valid
// Requirements: lowercase alphanumeric + hyphens only, 2-32 characters
func isValidChannelName(name string) bool {
	if name == "" {
		return false
	}
	return channelNameRegex.MatchString(name)
}
