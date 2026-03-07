package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// BootstrapService handles instance bootstrapping
type BootstrapService struct {
	userRepo    repository.UserRepository
	agentRepo   repository.AgentRepository
	apiKeyRepo  repository.APIKeyRepository
	inviteRepo  repository.InviteRepository
	channelRepo repository.ChannelRepository
	jwtSecret   string
	baseURL     string
}

// NewBootstrapService creates a new bootstrap service
func NewBootstrapService(
	userRepo repository.UserRepository,
	agentRepo repository.AgentRepository,
	apiKeyRepo repository.APIKeyRepository,
	inviteRepo repository.InviteRepository,
	channelRepo repository.ChannelRepository,
	jwtSecret string,
	baseURL string,
) *BootstrapService {
	return &BootstrapService{
		userRepo:    userRepo,
		agentRepo:   agentRepo,
		apiKeyRepo:  apiKeyRepo,
		inviteRepo:  inviteRepo,
		channelRepo: channelRepo,
		jwtSecret:   jwtSecret,
		baseURL:     baseURL,
	}
}

// Bootstrap performs atomic instance provisioning
func (s *BootstrapService) Bootstrap(ctx context.Context, req *models.BootstrapRequest) (*models.BootstrapResponse, error) {
	// Check if instance already bootstrapped
	userCount, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("check user count: %w", err)
	}
	if userCount > 0 {
		return nil, models.ErrInstanceAlreadyBootstrapped
	}

	// Validate request
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Generate UUIDs upfront for relationships
	primaryUserID := uuid.New().String()
	primaryAgentID := uuid.New().String()
	channelID := uuid.New().String()
	instanceID := uuid.New().String()

	// Create primary user
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.PrimaryAgent.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	primaryUser := &models.User{
		ID:           primaryUserID,
		Email:        req.PrimaryAgent.Email,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, primaryUser); err != nil {
		return nil, fmt.Errorf("create primary user: %w", err)
	}

	// Create primary agent
	primaryAgent := &models.Agent{
		ID:          primaryAgentID,
		OwnerID:     primaryUserID,
		Name:        req.PrimaryAgent.AgentProfile.Name,
		DisplayName: req.PrimaryAgent.AgentProfile.DisplayName,
		Description: req.PrimaryAgent.AgentProfile.Description,
		AvatarURL:   req.PrimaryAgent.AgentProfile.AvatarURL,
		Metadata:    req.PrimaryAgent.AgentProfile.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.agentRepo.Create(ctx, primaryAgent); err != nil {
		return nil, fmt.Errorf("create primary agent: %w", err)
	}

	// Generate primary agent API key
	primaryAPIKey, primaryAPIKeyPlaintext, err := s.generateAPIKey(ctx, primaryAgentID, "default")
	if err != nil {
		return nil, fmt.Errorf("create primary API key: %w", err)
	}

	// Generate JWT token for primary user
	jwtToken, err := s.generateJWTToken(primaryUserID, req.PrimaryAgent.Email)
	if err != nil {
		return nil, fmt.Errorf("generate JWT token: %w", err)
	}

	// Prepare response
	resp := &models.BootstrapResponse{
		PrimaryAgent: models.BootstrapPrimaryAgentResponse{
			UserID:   primaryUserID,
			AgentID:  primaryAgentID,
			Email:    req.PrimaryAgent.Email,
			JWTToken: jwtToken,
			APIKey:   primaryAPIKeyPlaintext,
			APIKeyID: primaryAPIKey.ID,
		},
		Agents:     []models.BootstrapAgentResponse{},
		Humans:     []models.BootstrapHumanResponse{},
		InstanceID: instanceID,
	}

	// Create additional agents
	for _, agentReq := range req.Agents {
		agentID := uuid.New().String()
		
		agent := &models.Agent{
			ID:          agentID,
			OwnerID:     primaryUserID,
			Name:        agentReq.Name,
			DisplayName: agentReq.DisplayName,
			Description: agentReq.Description,
			AvatarURL:   agentReq.AvatarURL,
			Metadata:    agentReq.Metadata,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.agentRepo.Create(ctx, agent); err != nil {
			return nil, fmt.Errorf("create agent %s: %w", agentReq.Name, err)
		}

		// Generate API key for agent
		apiKey, plaintext, err := s.generateAPIKey(ctx, agentID, "default")
		if err != nil {
			return nil, fmt.Errorf("create API key for agent %s: %w", agentReq.Name, err)
		}

		resp.Agents = append(resp.Agents, models.BootstrapAgentResponse{
			AgentID:     agentID,
			Name:        agentReq.Name,
			DisplayName: agentReq.DisplayName,
			APIKey:      plaintext,
			APIKeyID:    apiKey.ID,
		})
	}

	// Create human users and invites
	for _, humanReq := range req.Humans {
		humanUserID := uuid.New().String()
		
		// Create user stub (no password yet)
		humanUser := &models.User{
			ID:           humanUserID,
			Email:        humanReq.Email,
			PasswordHash: "", // Empty until invite consumed
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.userRepo.Create(ctx, humanUser); err != nil {
			return nil, fmt.Errorf("create human user %s: %w", humanReq.Email, err)
		}

		// Create invite
		inviteToken, inviteTokenHash, err := s.generateInviteToken()
		if err != nil {
			return nil, fmt.Errorf("generate invite token for %s: %w", humanReq.Email, err)
		}

		invite := &models.Invite{
			ID:        uuid.New().String(),
			UserID:    humanUserID,
			Status:    models.InviteStatusPending,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
			CreatedAt: time.Now(),
		}

		if err := s.inviteRepo.Create(ctx, invite, inviteTokenHash); err != nil {
			return nil, fmt.Errorf("create invite for %s: %w", humanReq.Email, err)
		}

		inviteURL := s.createInviteURL(inviteToken)

		resp.Humans = append(resp.Humans, models.BootstrapHumanResponse{
			UserID:      humanUserID,
			Email:       humanReq.Email,
			InviteToken: inviteToken,
			InviteURL:   inviteURL,
		})
	}

	// Create default channel
	channelName := req.DefaultChannel.Name
	if channelName == "" {
		channelName = "general"
	}

	channel := &models.Channel{
		ID:        channelID,
		Name:      channelName,
		Topic:     req.DefaultChannel.Topic,
		Type:      "channel",
		CreatedBy: primaryUserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	// Add members to channel (primary user is already added as owner by Channel.Create)
	members := []string{primaryUserID}

	// Add human users to channel
	for _, humanResp := range resp.Humans {
		if err := s.channelRepo.AddMember(ctx, channelID, humanResp.UserID, "member"); err != nil {
			return nil, fmt.Errorf("add human user %s to channel: %w", humanResp.Email, err)
		}
		members = append(members, humanResp.UserID)
	}

	resp.Channel = models.BootstrapChannelResponse{
		ChannelID: channelID,
		Name:      channelName,
		Topic:     req.DefaultChannel.Topic,
		Members:   members,
	}

	return resp, nil
}

// validateRequest validates the bootstrap request
func (s *BootstrapService) validateRequest(req *models.BootstrapRequest) error {
	// Basic validation would be handled by struct tags in handlers
	// Add any business logic validation here
	return nil
}

// generateAPIKey creates a new API key with secure random data
func (s *BootstrapService) generateAPIKey(ctx context.Context, agentID, name string) (*models.APIKey, string, error) {
	// Generate random 32 bytes for the key
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, "", fmt.Errorf("generate random bytes: %w", err)
	}

	// Create plaintext key with au_ prefix
	plaintextKey := "au_" + base64.URLEncoding.EncodeToString(randomBytes)

	// Hash the key for storage
	hash := sha256.Sum256([]byte(plaintextKey))
	keyHash := fmt.Sprintf("%x", hash)

	// Create key prefix for display (first 8 chars after prefix)
	keyPrefix := plaintextKey[:11] // "au_" + first 8 chars

	apiKey := &models.APIKey{
		ID:        uuid.New().String(),
		AgentID:   agentID,
		Name:      name,
		KeyPrefix: keyPrefix,
		CreatedAt: time.Now(),
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey, keyHash); err != nil {
		return nil, "", err
	}

	return apiKey, plaintextKey, nil
}

// generateInviteToken creates a new invite token
func (s *BootstrapService) generateInviteToken() (string, string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}

	token := "inv_" + base64.URLEncoding.EncodeToString(randomBytes)
	
	hash := sha256.Sum256([]byte(token))
	tokenHash := fmt.Sprintf("%x", hash)

	return token, tokenHash, nil
}

// generateJWTToken creates a JWT token for the user
func (s *BootstrapService) generateJWTToken(userID, email string) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// createInviteURL creates a full invite URL
func (s *BootstrapService) createInviteURL(token string) string {
	u, _ := url.Parse(s.baseURL)
	u.Path = "/invite"
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String()
}