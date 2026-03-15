package service

import (
	"context"
	"testing"

	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type unreadChannelRepoMock struct{ mock.Mock }

func (m *unreadChannelRepoMock) Create(ctx context.Context, channel *models.Channel) error {
	return m.Called(ctx, channel).Error(0)
}
func (m *unreadChannelRepoMock) GetByID(ctx context.Context, id string) (*models.Channel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}
func (m *unreadChannelRepoMock) Update(ctx context.Context, channelID, name, topic string) (*models.Channel, error) {
	args := m.Called(ctx, channelID, name, topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}
func (m *unreadChannelRepoMock) Delete(ctx context.Context, channelID string) error {
	return m.Called(ctx, channelID).Error(0)
}
func (m *unreadChannelRepoMock) ListByUser(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.ChannelWithMembers), args.Error(1)
}
func (m *unreadChannelRepoMock) GetMembers(ctx context.Context, channelID string) ([]*models.MemberInfo, error) {
	args := m.Called(ctx, channelID)
	return args.Get(0).([]*models.MemberInfo), args.Error(1)
}
func (m *unreadChannelRepoMock) IsMember(ctx context.Context, channelID, userID string) (bool, string, error) {
	args := m.Called(ctx, channelID, userID)
	return args.Bool(0), args.String(1), args.Error(2)
}
func (m *unreadChannelRepoMock) AddMember(ctx context.Context, channelID, userID, role string) error {
	return m.Called(ctx, channelID, userID, role).Error(0)
}
func (m *unreadChannelRepoMock) RemoveMember(ctx context.Context, channelID, userID string) error {
	return m.Called(ctx, channelID, userID).Error(0)
}
func (m *unreadChannelRepoMock) ListDMChannels(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.ChannelWithMembers), args.Error(1)
}
func (m *unreadChannelRepoMock) GetOrCreateDMChannel(ctx context.Context, user1ID, user2ID string) (*models.Channel, error) {
	args := m.Called(ctx, user1ID, user2ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}
func (m *unreadChannelRepoMock) MarkChannelRead(ctx context.Context, userID, channelID string) error {
	return m.Called(ctx, userID, channelID).Error(0)
}
func (m *unreadChannelRepoMock) GetUnreadCounts(ctx context.Context, userID string) (map[string]int, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]int), args.Error(1)
}

func TestChannelService_MarkChannelRead_SetsTimestamp(t *testing.T) {
	repo := new(unreadChannelRepoMock)
	svc := NewChannelService(repo)
	ctx := context.Background()
	repo.On("IsMember", ctx, "c1", "u1").Return(true, "member", nil)
	repo.On("MarkChannelRead", ctx, "u1", "c1").Return(nil)

	err := svc.MarkChannelRead(ctx, "c1", "u1")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestChannelService_List_ReturnsUnreadCount(t *testing.T) {
	repo := new(unreadChannelRepoMock)
	svc := NewChannelService(repo)
	ctx := context.Background()
	repo.On("ListByUser", ctx, "u1").Return([]*models.ChannelWithMembers{{Channel: models.Channel{ID: "c1"}, UnreadCount: 3}}, nil)
	repo.On("ListDMChannels", ctx, "u1").Return([]*models.ChannelWithMembers{}, nil)

	channels, err := svc.List(ctx, "u1")
	assert.NoError(t, err)
	assert.Equal(t, 3, channels[0].UnreadCount)
}
