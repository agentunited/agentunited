package repository

import (
	"context"

	"github.com/agentunited/backend/internal/models"
)

// APIKeyRepository handles API key data operations
type APIKeyRepository interface {
	Create(ctx context.Context, key *models.APIKey, keyHash string) error
	ListByAgent(ctx context.Context, agentID string) ([]*models.APIKey, error)
	GetByHash(ctx context.Context, keyHash string) (*models.APIKey, error)
	Delete(ctx context.Context, keyID string) error
}

type apiKeyRepository struct {
	db *DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, key *models.APIKey, keyHash string) error {
	query := `
		INSERT INTO agent_api_keys (agent_id, name, key_hash, key_prefix)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.db.Pool.QueryRow(ctx, query,
		key.AgentID, key.Name, keyHash, key.KeyPrefix,
	).Scan(&key.ID, &key.CreatedAt)
}

func (r *apiKeyRepository) ListByAgent(ctx context.Context, agentID string) ([]*models.APIKey, error) {
	query := `
		SELECT id, agent_id, name, key_prefix, last_used_at, created_at
		FROM agent_api_keys WHERE agent_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		key := &models.APIKey{}
		err := rows.Scan(&key.ID, &key.AgentID, &key.Name, &key.KeyPrefix, &key.LastUsedAt, &key.CreatedAt)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (r *apiKeyRepository) GetByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	key := &models.APIKey{}
	query := `
		SELECT id, agent_id, name, key_prefix, last_used_at, created_at
		FROM agent_api_keys WHERE key_hash = $1
	`
	err := r.db.Pool.QueryRow(ctx, query, keyHash).Scan(
		&key.ID, &key.AgentID, &key.Name, &key.KeyPrefix, &key.LastUsedAt, &key.CreatedAt,
	)
	return key, err
}

func (r *apiKeyRepository) Delete(ctx context.Context, keyID string) error {
	query := `DELETE FROM agent_api_keys WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, keyID)
	return err
}
