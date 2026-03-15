package service

import (
	"context"
	"testing"

	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) UpdateProfile(ctx context.Context, id, displayName, avatarURL string) error {
	args := m.Called(ctx, id, displayName, avatarURL)
	return args.Error(0)
}

func (m *mockUserRepository) List(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.User), args.Error(1)
}

type mockAgentRepository struct {
	mock.Mock
}

func (m *mockAgentRepository) Create(ctx context.Context, agent *models.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *mockAgentRepository) Get(ctx context.Context, id string) (*models.Agent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *mockAgentRepository) ListByOwner(ctx context.Context, ownerID string) ([]*models.Agent, error) {
	args := m.Called(ctx, ownerID)
	return args.Get(0).([]*models.Agent), args.Error(1)
}

func (m *mockAgentRepository) ListAll(ctx context.Context) ([]*models.Agent, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Agent), args.Error(1)
}

func (m *mockAgentRepository) Update(ctx context.Context, agent *models.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *mockAgentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockAPIKeyRepository struct {
	mock.Mock
}

func (m *mockAPIKeyRepository) Create(ctx context.Context, key *models.APIKey, keyHash string) error {
	args := m.Called(ctx, key, keyHash)
	return args.Error(0)
}

func (m *mockAPIKeyRepository) ListByAgent(ctx context.Context, agentID string) ([]*models.APIKey, error) {
	args := m.Called(ctx, agentID)
	return args.Get(0).([]*models.APIKey), args.Error(1)
}

func (m *mockAPIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	args := m.Called(ctx, keyHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}

func (m *mockAPIKeyRepository) Delete(ctx context.Context, keyID string) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

type mockInviteRepository struct {
	mock.Mock
}

func (m *mockInviteRepository) Create(ctx context.Context, invite *models.Invite, tokenHash string) error {
	args := m.Called(ctx, invite, tokenHash)
	return args.Error(0)
}

func (m *mockInviteRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invite), args.Error(1)
}

func (m *mockInviteRepository) ValidateToken(ctx context.Context, tokenHash string) (*models.Invite, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invite), args.Error(1)
}

func (m *mockInviteRepository) ConsumeToken(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

type mockChannelRepository struct {
	mock.Mock
}

func (m *mockChannelRepository) Create(ctx context.Context, channel *models.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *mockChannelRepository) GetByID(ctx context.Context, id string) (*models.Channel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *mockChannelRepository) GetByName(ctx context.Context, name string) (*models.Channel, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *mockChannelRepository) ListByUser(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.ChannelWithMembers), args.Error(1)
}

func (m *mockChannelRepository) AddMember(ctx context.Context, channelID, userID, role string) error {
	args := m.Called(ctx, channelID, userID, role)
	return args.Error(0)
}

func (m *mockChannelRepository) GetMembers(ctx context.Context, channelID string) ([]*models.MemberInfo, error) {
	args := m.Called(ctx, channelID)
	return args.Get(0).([]*models.MemberInfo), args.Error(1)
}

func (m *mockChannelRepository) IsMember(ctx context.Context, channelID, userID string) (bool, string, error) {
	args := m.Called(ctx, channelID, userID)
	return args.Get(0).(bool), args.Get(1).(string), args.Error(2)
}

func (m *mockChannelRepository) Update(ctx context.Context, channelID, name, topic string) (*models.Channel, error) {
	args := m.Called(ctx, channelID, name, topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *mockChannelRepository) Delete(ctx context.Context, channelID string) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *mockChannelRepository) RemoveMember(ctx context.Context, channelID, userID string) error {
	args := m.Called(ctx, channelID, userID)
	return args.Error(0)
}

func (m *mockChannelRepository) ListDMChannels(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ChannelWithMembers), args.Error(1)
}

func (m *mockChannelRepository) GetOrCreateDMChannel(ctx context.Context, user1ID, user2ID string) (*models.Channel, error) {
	args := m.Called(ctx, user1ID, user2ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *mockChannelRepository) MarkChannelRead(ctx context.Context, userID, channelID string) error {
	args := m.Called(ctx, userID, channelID)
	return args.Error(0)
}

func (m *mockChannelRepository) GetUnreadCounts(ctx context.Context, userID string) (map[string]int, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

type mockSubscriptionRepository struct {
	mock.Mock
}

func (m *mockSubscriptionRepository) GetByWorkspace(ctx context.Context, workspaceID string) (*models.Subscription, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepository) UpsertByWorkspace(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *mockSubscriptionRepository) UpsertByStripeSubscriptionID(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func TestBootstrapService_Bootstrap_InstanceAlreadyBootstrapped(t *testing.T) {
	userRepo := &mockUserRepository{}
	agentRepo := &mockAgentRepository{}
	apiKeyRepo := &mockAPIKeyRepository{}
	inviteRepo := &mockInviteRepository{}
	channelRepo := &mockChannelRepository{}

	subscriptionRepo := &mockSubscriptionRepository{}
	service := NewBootstrapService(userRepo, agentRepo, apiKeyRepo, inviteRepo, channelRepo, subscriptionRepo, "test-jwt-secret", "https://agentunited.ai", "tunnel.agentunited.ai", nil)

	ctx := context.Background()

	userRepo.On("GetByEmail", ctx, "admin@example.com").Return(nil, models.ErrUserNotFound)
	// Mock existing users (instance already bootstrapped)
	userRepo.On("Count", ctx).Return(int64(1), nil)

	req := &models.BootstrapRequest{
		PrimaryAgent: models.BootstrapPrimaryAgent{
			Email:    "admin@example.com",
			Password: "supersecurepassword123",
			AgentProfile: models.BootstrapAgentProfile{
				Name:        "coordinator",
				DisplayName: "Coordinator Agent",
			},
		},
	}

	_, err := service.Bootstrap(ctx, req)
	require.Error(t, err)
	assert.Equal(t, models.ErrInstanceAlreadyBootstrapped, err)

	userRepo.AssertExpectations(t)
}

func TestBootstrapService_Bootstrap_HappyPath(t *testing.T) {
	userRepo := &mockUserRepository{}
	agentRepo := &mockAgentRepository{}
	apiKeyRepo := &mockAPIKeyRepository{}
	inviteRepo := &mockInviteRepository{}
	channelRepo := &mockChannelRepository{}

	subscriptionRepo := &mockSubscriptionRepository{}
	service := NewBootstrapService(userRepo, agentRepo, apiKeyRepo, inviteRepo, channelRepo, subscriptionRepo, "test-jwt-secret", "https://agentunited.ai", "tunnel.agentunited.ai", nil)

	ctx := context.Background()

	userRepo.On("GetByEmail", ctx, "admin@example.com").Return(nil, models.ErrUserNotFound)
	// Mock empty instance
	userRepo.On("Count", ctx).Return(int64(0), nil)

	req := &models.BootstrapRequest{
		PrimaryAgent: models.BootstrapPrimaryAgent{
			Email:    "admin@example.com",
			Password: "supersecurepassword123",
			AgentProfile: models.BootstrapAgentProfile{
				Name:        "coordinator",
				DisplayName: "Coordinator Agent",
			},
		},
		Agents: []models.BootstrapAgent{
			{
				Name:        "worker",
				DisplayName: "Worker Agent",
			},
		},
		Humans: []models.BootstrapHuman{
			{
				Email:       "human@example.com",
				DisplayName: "Human User",
				Role:        "member",
			},
		},
		DefaultChannel: models.BootstrapChannel{
			Name:  "general",
			Topic: "General discussion",
		},
	}

	// Mock all the expected repository calls
	userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil).Times(2)   // primary + human
	agentRepo.On("Create", ctx, mock.AnythingOfType("*models.Agent")).Return(nil).Times(2) // primary + worker
	apiKeyRepo.On("Create", ctx, mock.AnythingOfType("*models.APIKey"), mock.AnythingOfType("string")).Return(nil).Times(2)
	inviteRepo.On("Create", ctx, mock.AnythingOfType("*models.Invite"), mock.AnythingOfType("string")).Return(nil)
	channelRepo.On("Create", ctx, mock.AnythingOfType("*models.Channel")).Return(nil)
	channelRepo.On("AddMember", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Times(1) // human
	subscriptionRepo.On("UpsertByWorkspace", ctx, mock.AnythingOfType("*models.Subscription")).Return(nil)

	resp, err := service.Bootstrap(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify response structure
	assert.NotEmpty(t, resp.PrimaryAgent.UserID)
	assert.NotEmpty(t, resp.PrimaryAgent.AgentID)
	assert.Equal(t, "admin@example.com", resp.PrimaryAgent.Email)
	assert.NotEmpty(t, resp.PrimaryAgent.JWTToken)
	assert.NotEmpty(t, resp.PrimaryAgent.APIKey)

	assert.Len(t, resp.Agents, 1)
	assert.Equal(t, "worker", resp.Agents[0].Name)
	assert.NotEmpty(t, resp.Agents[0].APIKey)

	assert.Len(t, resp.Humans, 1)
	assert.Equal(t, "human@example.com", resp.Humans[0].Email)
	assert.NotEmpty(t, resp.Humans[0].InviteToken)
	assert.NotEmpty(t, resp.Humans[0].InviteURL)

	assert.Equal(t, "general", resp.Channel.Name)
	assert.NotEmpty(t, resp.Channel.ChannelID)
	assert.Len(t, resp.Channel.Members, 2)
	assert.NotEmpty(t, resp.RelayURL)
	assert.Equal(t, "free", resp.RelayTier)
	assert.Equal(t, 1024, resp.RelayBandwidthLimitMB)

	userRepo.AssertExpectations(t)
	agentRepo.AssertExpectations(t)
	apiKeyRepo.AssertExpectations(t)
	inviteRepo.AssertExpectations(t)
	channelRepo.AssertExpectations(t)
	subscriptionRepo.AssertExpectations(t)
}
