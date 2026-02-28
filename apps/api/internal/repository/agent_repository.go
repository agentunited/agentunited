package repository

import (
	"context"
	"encoding/json"

	"github.com/agentunited/backend/internal/models"
)

// AgentRepository handles agent data operations
type AgentRepository interface {
	Create(ctx context.Context, agent *models.Agent) error
	Get(ctx context.Context, id string) (*models.Agent, error)
	ListByOwner(ctx context.Context, ownerID string) ([]*models.Agent, error)
	Update(ctx context.Context, agent *models.Agent) error
	Delete(ctx context.Context, id string) error
}

type agentRepository struct {
	db *DB
}

// NewAgentRepository creates a new agent repository
func NewAgentRepository(db *DB) AgentRepository {
	return &agentRepository{db: db}
}

func (r *agentRepository) Create(ctx context.Context, agent *models.Agent) error {
	metadataJSON, _ := json.Marshal(agent.Metadata)
	
	var query string
	var err error

	if agent.ID != "" {
		// Use provided ID (for bootstrap scenarios)
		query = `
			INSERT INTO agents (id, owner_id, name, display_name, description, avatar_url, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING created_at, updated_at
		`
		err = r.db.Pool.QueryRow(ctx, query,
			agent.ID, agent.OwnerID, agent.Name, agent.DisplayName, agent.Description, agent.AvatarURL, metadataJSON,
		).Scan(&agent.CreatedAt, &agent.UpdatedAt)
	} else {
		// Let database generate ID
		query = `
			INSERT INTO agents (owner_id, name, display_name, description, avatar_url, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at, updated_at
		`
		err = r.db.Pool.QueryRow(ctx, query,
			agent.OwnerID, agent.Name, agent.DisplayName, agent.Description, agent.AvatarURL, metadataJSON,
		).Scan(&agent.ID, &agent.CreatedAt, &agent.UpdatedAt)
	}
	
	return err
}

func (r *agentRepository) Get(ctx context.Context, id string) (*models.Agent, error) {
	agent := &models.Agent{}
	var metadataJSON []byte
	query := `
		SELECT id, owner_id, name, display_name, description, avatar_url, metadata, created_at, updated_at
		FROM agents WHERE id = $1
	`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&agent.ID, &agent.OwnerID, &agent.Name, &agent.DisplayName, &agent.Description,
		&agent.AvatarURL, &metadataJSON, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metadataJSON, &agent.Metadata)
	return agent, nil
}

func (r *agentRepository) ListByOwner(ctx context.Context, ownerID string) ([]*models.Agent, error) {
	query := `
		SELECT id, owner_id, name, display_name, description, avatar_url, metadata, created_at, updated_at
		FROM agents WHERE owner_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*models.Agent
	for rows.Next() {
		agent := &models.Agent{}
		var metadataJSON []byte
		err := rows.Scan(
			&agent.ID, &agent.OwnerID, &agent.Name, &agent.DisplayName, &agent.Description,
			&agent.AvatarURL, &metadataJSON, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(metadataJSON, &agent.Metadata)
		agents = append(agents, agent)
	}
	return agents, nil
}

func (r *agentRepository) Update(ctx context.Context, agent *models.Agent) error {
	metadataJSON, _ := json.Marshal(agent.Metadata)
	query := `
		UPDATE agents 
		SET display_name = $1, description = $2, avatar_url = $3, metadata = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	return r.db.Pool.QueryRow(ctx, query,
		agent.DisplayName, agent.Description, agent.AvatarURL, metadataJSON, agent.ID,
	).Scan(&agent.UpdatedAt)
}

func (r *agentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM agents WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}
