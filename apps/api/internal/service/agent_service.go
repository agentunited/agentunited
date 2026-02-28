package service

import (
	"context"
	"errors"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
)

var (
	ErrAgentNotFound  = errors.New("agent not found")
	ErrNotAgentOwner  = errors.New("not agent owner")
	ErrAgentNameTaken = errors.New("agent name already taken")
)

// AgentService handles business logic for agents
type AgentService interface {
	CreateAgent(ctx context.Context, ownerID string, req *models.CreateAgentRequest) (*models.Agent, error)
	GetAgent(ctx context.Context, agentID string) (*models.Agent, error)
	ListAgents(ctx context.Context, ownerID string) ([]*models.Agent, error)
	UpdateAgent(ctx context.Context, agentID, userID string, req *models.UpdateAgentRequest) (*models.Agent, error)
	DeleteAgent(ctx context.Context, agentID, userID string) error
}

type agentService struct {
	repo repository.AgentRepository
}

// NewAgentService creates a new agent service
func NewAgentService(repo repository.AgentRepository) AgentService {
	return &agentService{repo: repo}
}

func (s *agentService) CreateAgent(ctx context.Context, ownerID string, req *models.CreateAgentRequest) (*models.Agent, error) {
	agent := &models.Agent{
		OwnerID:     ownerID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		AvatarURL:   req.AvatarURL,
		Metadata:    req.Metadata,
	}
	if agent.Metadata == nil {
		agent.Metadata = make(map[string]interface{})
	}
	
	if err := s.repo.Create(ctx, agent); err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			return nil, ErrAgentNameTaken
		}
		return nil, err
	}
	return agent, nil
}

func (s *agentService) GetAgent(ctx context.Context, agentID string) (*models.Agent, error) {
	agent, err := s.repo.Get(ctx, agentID)
	if err != nil {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

func (s *agentService) ListAgents(ctx context.Context, ownerID string) ([]*models.Agent, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

func (s *agentService) UpdateAgent(ctx context.Context, agentID, userID string, req *models.UpdateAgentRequest) (*models.Agent, error) {
	agent, err := s.repo.Get(ctx, agentID)
	if err != nil {
		return nil, ErrAgentNotFound
	}
	if agent.OwnerID != userID {
		return nil, ErrNotAgentOwner
	}

	// Apply updates
	if req.DisplayName != nil {
		agent.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		agent.Description = *req.Description
	}
	if req.AvatarURL != nil {
		agent.AvatarURL = *req.AvatarURL
	}
	if req.Metadata != nil {
		agent.Metadata = *req.Metadata
	}

	if err := s.repo.Update(ctx, agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *agentService) DeleteAgent(ctx context.Context, agentID, userID string) error {
	agent, err := s.repo.Get(ctx, agentID)
	if err != nil {
		return ErrAgentNotFound
	}
	if agent.OwnerID != userID {
		return ErrNotAgentOwner
	}
	return s.repo.Delete(ctx, agentID)
}

// Helper to detect unique constraint violations
func isUniqueViolation(err error) bool {
	// PostgreSQL unique violation error code is 23505
	return err != nil && err.Error() == "ERROR: duplicate key value violates unique constraint \"idx_agents_owner_name\" (SQLSTATE 23505)"
}
