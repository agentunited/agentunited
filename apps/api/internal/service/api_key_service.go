package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
)

var (
	ErrAPIKeyNotFound = errors.New("API key not found")
	ErrInvalidAPIKey  = errors.New("invalid API key")
)

// APIKeyService handles API key business logic
type APIKeyService interface {
	CreateKey(ctx context.Context, agentID, ownerID, name string) (*models.APIKeyWithPlaintext, error)
	ListKeys(ctx context.Context, agentID, ownerID string) ([]*models.APIKey, error)
	DeleteKey(ctx context.Context, keyID, agentID, ownerID string) error
	ValidateKey(ctx context.Context, plaintextKey string) (*models.APIKey, error)
}

type apiKeyService struct {
	keyRepo   repository.APIKeyRepository
	agentRepo repository.AgentRepository
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(keyRepo repository.APIKeyRepository, agentRepo repository.AgentRepository) APIKeyService {
	return &apiKeyService{
		keyRepo:   keyRepo,
		agentRepo: agentRepo,
	}
}

func (s *apiKeyService) CreateKey(ctx context.Context, agentID, ownerID, name string) (*models.APIKeyWithPlaintext, error) {
	// Verify ownership
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return nil, ErrAgentNotFound
	}
	if agent.OwnerID != ownerID {
		return nil, ErrNotAgentOwner
	}

	// Generate key: au_<random32bytes>
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	plaintextKey := "au_" + base64.URLEncoding.EncodeToString(randomBytes)
	keyHash := hashKey(plaintextKey)
	keyPrefix := plaintextKey[:10] + "..."

	apiKey := &models.APIKey{
		AgentID:   agentID,
		Name:      name,
		KeyPrefix: keyPrefix,
	}

	if err := s.keyRepo.Create(ctx, apiKey, keyHash); err != nil {
		return nil, err
	}

	return &models.APIKeyWithPlaintext{
		APIKey:       *apiKey,
		PlaintextKey: plaintextKey,
	}, nil
}

func (s *apiKeyService) ListKeys(ctx context.Context, agentID, ownerID string) ([]*models.APIKey, error) {
	// Verify ownership
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return nil, ErrAgentNotFound
	}
	if agent.OwnerID != ownerID {
		return nil, ErrNotAgentOwner
	}

	return s.keyRepo.ListByAgent(ctx, agentID)
}

func (s *apiKeyService) DeleteKey(ctx context.Context, keyID, agentID, ownerID string) error {
	// Verify ownership
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return ErrAgentNotFound
	}
	if agent.OwnerID != ownerID {
		return ErrNotAgentOwner
	}

	return s.keyRepo.Delete(ctx, keyID)
}

func (s *apiKeyService) ValidateKey(ctx context.Context, plaintextKey string) (*models.APIKey, error) {
	keyHash := hashKey(plaintextKey)
	key, err := s.keyRepo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, ErrInvalidAPIKey
	}
	return key, nil
}

func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}
