package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/agentunited/backend/internal/models"
)

type IntegrationRepository interface {
	Create(ctx context.Context, integration *models.Integration) error
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.Integration, error)
	Delete(ctx context.Context, id, workspaceID string) error
	ListActiveByEvent(ctx context.Context, workspaceID, eventType string) ([]*models.Integration, error)
}

type integrationRepository struct{ db *DB }

func NewIntegrationRepository(db *DB) IntegrationRepository { return &integrationRepository{db: db} }

func (r *integrationRepository) Create(ctx context.Context, i *models.Integration) error {
	eventsJSON, _ := json.Marshal(i.EventSubscriptions)
	q := `INSERT INTO integrations (workspace_id, name, platform, api_key, webhook_url, event_subscriptions, active)
	      VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`
	return r.db.Pool.QueryRow(ctx, q, i.WorkspaceID, i.Name, i.Platform, i.APIKey, i.WebhookURL, eventsJSON, i.Active).Scan(&i.ID, &i.CreatedAt)
}

func (r *integrationRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.Integration, error) {
	q := `SELECT id, workspace_id, name, platform, api_key, webhook_url, event_subscriptions, active, created_at
	      FROM integrations WHERE workspace_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Pool.Query(ctx, q, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list integrations: %w", err)
	}
	defer rows.Close()

	out := []*models.Integration{}
	for rows.Next() {
		var i models.Integration
		var raw []byte
		if err := rows.Scan(&i.ID, &i.WorkspaceID, &i.Name, &i.Platform, &i.APIKey, &i.WebhookURL, &raw, &i.Active, &i.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan integration: %w", err)
		}
		if len(raw) == 0 {
			i.EventSubscriptions = []string{}
		} else {
			_ = json.Unmarshal(raw, &i.EventSubscriptions)
			if i.EventSubscriptions == nil {
				i.EventSubscriptions = []string{}
			}
		}
		out = append(out, &i)
	}
	return out, rows.Err()
}

func (r *integrationRepository) Delete(ctx context.Context, id, workspaceID string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM integrations WHERE id = $1 AND workspace_id = $2`, id, workspaceID)
	return err
}

func (r *integrationRepository) ListActiveByEvent(ctx context.Context, workspaceID, eventType string) ([]*models.Integration, error) {
	q := `SELECT id, workspace_id, name, platform, api_key, webhook_url, event_subscriptions, active, created_at
	      FROM integrations
	      WHERE workspace_id = $1 AND active = true AND event_subscriptions @> to_jsonb(ARRAY[$2]::text[])`
	rows, err := r.db.Pool.Query(ctx, q, workspaceID, eventType)
	if err != nil {
		return nil, fmt.Errorf("list active integrations: %w", err)
	}
	defer rows.Close()

	out := []*models.Integration{}
	for rows.Next() {
		var i models.Integration
		var raw []byte
		if err := rows.Scan(&i.ID, &i.WorkspaceID, &i.Name, &i.Platform, &i.APIKey, &i.WebhookURL, &raw, &i.Active, &i.CreatedAt); err != nil {
			return nil, err
		}
		if len(raw) == 0 {
			i.EventSubscriptions = []string{}
		} else {
			_ = json.Unmarshal(raw, &i.EventSubscriptions)
			if i.EventSubscriptions == nil {
				i.EventSubscriptions = []string{}
			}
		}
		out = append(out, &i)
	}
	return out, rows.Err()
}
