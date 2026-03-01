package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/lib/pq"
)

// WebhookRepository handles webhook data operations
type WebhookRepository interface {
	Create(ctx context.Context, webhook *models.Webhook, secret string) error
	ListByAgent(ctx context.Context, agentID string) ([]*models.Webhook, error)
	Delete(ctx context.Context, id string) error
	ListDeliveries(ctx context.Context, webhookID string, limit int) ([]*models.WebhookDelivery, error)
	ListByChannel(ctx context.Context, channelID, eventType string) ([]*models.WebhookWithSecret, error)
	CreateDelivery(ctx context.Context, delivery *models.WebhookDelivery) error
	UpdateDelivery(ctx context.Context, delivery *models.WebhookDelivery) error
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

func (r *webhookRepository) ListByChannel(ctx context.Context, channelID, eventType string) ([]*models.WebhookWithSecret, error) {
	query := `
		SELECT DISTINCT w.id, w.agent_id, w.url, w.secret, w.events, w.active, w.created_at, w.updated_at
		FROM webhooks w
		INNER JOIN agents a ON w.agent_id = a.id
		INNER JOIN channel_members cm ON a.owner_id = cm.user_id
		WHERE cm.channel_id = $1 
		  AND w.active = true
		  AND $2 = ANY(w.events)
		ORDER BY w.created_at
	`
	rows, err := r.db.Pool.Query(ctx, query, channelID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*models.WebhookWithSecret
	for rows.Next() {
		var id, agentID, url, secret string
		var eventsStr string
		var active bool
		var createdAt, updatedAt time.Time
		
		err := rows.Scan(&id, &agentID, &url, &secret, &eventsStr,
			&active, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		
		// Parse PostgreSQL array string manually (e.g., "{message.created}" -> ["message.created"])
		events := parsePostgresArray(eventsStr)
		
		webhook := &models.WebhookWithSecret{
			Webhook: models.Webhook{
				ID:        id,
				AgentID:   agentID,
				URL:       url,
				Events:    events,
				Active:    active,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			Secret: secret,
		}
		webhooks = append(webhooks, webhook)
	}
	return webhooks, nil
}

func (r *webhookRepository) CreateDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	payloadJSON, _ := json.Marshal(delivery.Payload)
	query := `
		INSERT INTO webhook_deliveries (webhook_id, event_type, payload, status, attempt_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.Pool.QueryRow(ctx, query,
		delivery.WebhookID, delivery.EventType, payloadJSON, delivery.Status, delivery.AttemptCount,
	).Scan(&delivery.ID, &delivery.CreatedAt)
}

func (r *webhookRepository) UpdateDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	query := `
		UPDATE webhook_deliveries 
		SET status = $1, response_code = $2, response_body = $3, attempt_count = $4, delivered_at = $5
		WHERE id = $6
	`
	_, err := r.db.Pool.Exec(ctx, query,
		delivery.Status, delivery.ResponseCode, delivery.ResponseBody, delivery.AttemptCount, delivery.DeliveredAt, delivery.ID,
	)
	return err
}

// parsePostgresArray parses a PostgreSQL array string like "{value1,value2}" into []string
func parsePostgresArray(arrayStr string) []string {
	// Remove the surrounding braces
	if len(arrayStr) < 2 || arrayStr[0] != '{' || arrayStr[len(arrayStr)-1] != '}' {
		return []string{}
	}
	
	// Remove braces and split by comma
	inner := arrayStr[1 : len(arrayStr)-1]
	if inner == "" {
		return []string{}
	}
	
	// Split by comma and trim spaces
	parts := strings.Split(inner, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	
	return result
}
