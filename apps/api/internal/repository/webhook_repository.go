package repository

import (
	"context"
	"encoding/json"

	"github.com/agentunited/backend/internal/models"
	"github.com/lib/pq"
)

// WebhookRepository handles webhook data operations
type WebhookRepository interface {
	Create(ctx context.Context, webhook *models.Webhook, secret string) error
	ListByAgent(ctx context.Context, agentID string) ([]*models.Webhook, error)
	Delete(ctx context.Context, id string) error
	ListDeliveries(ctx context.Context, webhookID string, limit int) ([]*models.WebhookDelivery, error)
}

type webhookRepository struct {
	db *DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *DB) WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) Create(ctx context.Context, webhook *models.Webhook, secret string) error {
	query := `
		INSERT INTO webhooks (agent_id, url, secret, events, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.db.Pool.QueryRow(ctx, query,
		webhook.AgentID, webhook.URL, secret, pq.Array(webhook.Events), webhook.Active,
	).Scan(&webhook.ID, &webhook.CreatedAt, &webhook.UpdatedAt)
}

func (r *webhookRepository) ListByAgent(ctx context.Context, agentID string) ([]*models.Webhook, error) {
	query := `
		SELECT id, agent_id, url, events, active, created_at, updated_at
		FROM webhooks WHERE agent_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*models.Webhook
	for rows.Next() {
		webhook := &models.Webhook{}
		err := rows.Scan(&webhook.ID, &webhook.AgentID, &webhook.URL, pq.Array(&webhook.Events),
			&webhook.Active, &webhook.CreatedAt, &webhook.UpdatedAt)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, webhook)
	}
	return webhooks, nil
}

func (r *webhookRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM webhooks WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *webhookRepository) ListDeliveries(ctx context.Context, webhookID string, limit int) ([]*models.WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, event_type, payload, status, response_code, response_body, attempt_count, created_at, delivered_at
		FROM webhook_deliveries WHERE webhook_id = $1 ORDER BY created_at DESC LIMIT $2
	`
	rows, err := r.db.Pool.Query(ctx, query, webhookID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*models.WebhookDelivery
	for rows.Next() {
		delivery := &models.WebhookDelivery{}
		var payloadJSON []byte
		err := rows.Scan(&delivery.ID, &delivery.WebhookID, &delivery.EventType, &payloadJSON,
			&delivery.Status, &delivery.ResponseCode, &delivery.ResponseBody, &delivery.AttemptCount,
			&delivery.CreatedAt, &delivery.DeliveredAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(payloadJSON, &delivery.Payload)
		deliveries = append(deliveries, delivery)
	}
	return deliveries, nil
}
