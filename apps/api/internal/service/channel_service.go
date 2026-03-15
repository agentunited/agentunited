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
	Update(ctx context.Context, channelID, userID, name, topic string) (*models.Channel, error)
	Delete(ctx context.Context, channelID, userID string) error
	List(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error)
	GetMembers(ctx context.Context, channelID, userID string) ([]*models.MemberInfo, error)
	AddMember(ctx context.Context, channelID, targetUserID, requesterUserID, role string) error
	RemoveMember(ctx context.Context, channelID, targetUserID, requesterUserID string) error
	ListDMChannels(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error)
	CreateOrGetDMChannel(ctx context.Context, userID, targetUserID string) (*models.Channel, error)
	IsMember(ctx context.Context, channelID, userID string) (bool, string, error)
	MarkChannelRead(ctx context.Context, channelID, userID string) error
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
		Type:      "channel",
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

// List retrieves all channels where the user is a member.
// Includes both regular channels and DMs so sidebar empty-state fetches can still surface DMs.
func (s *channelService) List(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	channels, err := s.channelRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	dms, err := s.channelRepo.ListDMChannels(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list dms: %w", err)
	}
	return append(channels, dms...), nil
}

// IsMember checks if a user is a member of a channel
func (s *channelService) IsMember(ctx context.Context, channelID, userID string) (bool, string, error) {
	isMember, role, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return false, "", fmt.Errorf("check membership: %w", err)
	}

	return isMember, role, nil
}

// Update modifies a channel's name and topic (owner only)
func (s *channelService) Update(ctx context.Context, channelID, userID, name, topic string) (*models.Channel, error) {
	// Check if user is a member and get their role
	isMember, role, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, models.ErrNotChannelMember
	}

	// Only owners can update channel
	if role != "owner" {
		return nil, models.ErrInsufficientPermissions
	}

	// Validate channel name if provided
	if name != "" && !isValidChannelName(name) {
		return nil, models.ErrInvalidChannelName
	}

	// Update channel
	channel, err := s.channelRepo.Update(ctx, channelID, name, topic)
	if err != nil {
		return nil, fmt.Errorf("update channel: %w", err)
	}

	return channel, nil
}

// Delete removes a channel (owner only)
func (s *channelService) Delete(ctx context.Context, channelID, userID string) error {
	// Check if user is a member and get their role
	isMember, role, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return models.ErrNotChannelMember
	}

	// Only owners can delete channel
	if role != "owner" {
		return models.ErrInsufficientPermissions
	}

	// Delete channel
	if err := s.channelRepo.Delete(ctx, channelID); err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}

	return nil
}

// GetMembers retrieves channel members (members only)
func (s *channelService) GetMembers(ctx context.Context, channelID, userID string) ([]*models.MemberInfo, error) {
	// Check if user is a member
	isMember, _, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, models.ErrNotChannelMember
	}

	// Get members
	members, err := s.channelRepo.GetMembers(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	return members, nil
}

// AddMember adds a user to a channel (owners and admins only)
func (s *channelService) AddMember(ctx context.Context, channelID, targetUserID, requesterUserID, role string) error {
	// Check if requester is a member and get their role
	isMember, requesterRole, err := s.channelRepo.IsMember(ctx, channelID, requesterUserID)
	if err != nil {
		return fmt.Errorf("check requester membership: %w", err)
	}
	if !isMember {
		return models.ErrNotChannelMember
	}

	// Only owners and admins can add members
	if requesterRole != "owner" && requesterRole != "admin" {
		return models.ErrInsufficientPermissions
	}

	// Validate role
	validRoles := []string{"member", "admin", "owner"}
	isValidRole := false
	for _, r := range validRoles {
		if role == r {
			isValidRole = true
			break
		}
	}
	if !isValidRole {
		role = "member" // Default to member if invalid role provided
	}

	// Only owners can create other owners
	if role == "owner" && requesterRole != "owner" {
		return models.ErrInsufficientPermissions
	}

	// Add member
	if err := s.channelRepo.AddMember(ctx, channelID, targetUserID, role); err != nil {
		return fmt.Errorf("add member: %w", err)
	}

	return nil
}

// RemoveMember removes a user from a channel (owners and admins only, with restrictions)
func (s *channelService) RemoveMember(ctx context.Context, channelID, targetUserID, requesterUserID string) error {
	// Check if requester is a member and get their role
	isMember, requesterRole, err := s.channelRepo.IsMember(ctx, channelID, requesterUserID)
	if err != nil {
		return fmt.Errorf("check requester membership: %w", err)
	}
	if !isMember {
		return models.ErrNotChannelMember
	}

	// Check if target is a member and get their role
	targetIsMember, targetRole, err := s.channelRepo.IsMember(ctx, channelID, targetUserID)
	if err != nil {
		return fmt.Errorf("check target membership: %w", err)
	}
	if !targetIsMember {
		return models.ErrNotChannelMember
	}

	// Rules for removal:
	// - Owners can remove anyone except other owners (unless it's themselves leaving)
	// - Admins can remove members but not other admins or owners
	// - Members can only remove themselves
	canRemove := false

	if requesterUserID == targetUserID {
		// Users can always remove themselves
		canRemove = true
	} else if requesterRole == "owner" && targetRole != "owner" {
		// Owners can remove admins and members
		canRemove = true
	} else if requesterRole == "admin" && targetRole == "member" {
		// Admins can remove members
		canRemove = true
	}

	if !canRemove {
		return models.ErrInsufficientPermissions
	}

	// Remove member
	if err := s.channelRepo.RemoveMember(ctx, channelID, targetUserID); err != nil {
		return fmt.Errorf("remove member: %w", err)
	}

	return nil
}

// ListDMChannels retrieves all DM channels for a user
func (s *channelService) ListDMChannels(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	channels, err := s.channelRepo.ListDMChannels(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list dm channels: %w", err)
	}

	return channels, nil
}

// CreateOrGetDMChannel creates a new DM channel or returns existing one between two users
func (s *channelService) CreateOrGetDMChannel(ctx context.Context, userID, targetUserID string) (*models.Channel, error) {
	// Users can't create DM with themselves
	if userID == targetUserID {
		return nil, models.ErrInvalidDMTarget
	}

	// Get or create DM channel
	channel, err := s.channelRepo.GetOrCreateDMChannel(ctx, userID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("get or create dm channel: %w", err)
	}

	return channel, nil
}

// MarkChannelRead marks a channel as read for a user.
func (s *channelService) MarkChannelRead(ctx context.Context, channelID, userID string) error {
	isMember, _, err := s.channelRepo.IsMember(ctx, channelID, userID)
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return models.ErrNotChannelMember
	}
	if err := s.channelRepo.MarkChannelRead(ctx, userID, channelID); err != nil {
		return fmt.Errorf("mark channel read: %w", err)
	}
	return nil
}

// isValidChannelName checks if a channel name is valid
// Requirements: lowercase alphanumeric + hyphens only, 2-32 characters
func isValidChannelName(name string) bool {
	if name == "" {
		return false
	}
	return channelNameRegex.MatchString(name)
}
