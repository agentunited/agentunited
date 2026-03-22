package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRecord struct {
	WorkspaceID          string
	StripeCustomerID     string
	StripeSubscriptionID string
	Plan                 string
	Status               string
	CurrentPeriodEnd     *time.Time
	RelayTier            string
	BandwidthLimitMB     int
	EntityLimit          int
	RelayEnabled         bool
}

type SubscriptionStore interface {
	UpsertSubscription(ctx context.Context, s SubscriptionRecord) error
	GetWorkspaceEmail(ctx context.Context, workspaceID string) (string, error)
}

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	p, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{pool: p}, nil
}

func (s *PostgresStore) Close() { s.pool.Close() }

func (s *PostgresStore) UpsertSubscription(ctx context.Context, sub SubscriptionRecord) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO subscriptions (
	workspace_id, stripe_customer_id, stripe_subscription_id, plan, status, current_period_end,
	relay_tier, bandwidth_limit_mb, entity_limit, relay_enabled, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())
ON CONFLICT (workspace_id) DO UPDATE SET
	stripe_customer_id = EXCLUDED.stripe_customer_id,
	stripe_subscription_id = EXCLUDED.stripe_subscription_id,
	plan = EXCLUDED.plan,
	status = EXCLUDED.status,
	current_period_end = EXCLUDED.current_period_end,
	relay_tier = EXCLUDED.relay_tier,
	bandwidth_limit_mb = EXCLUDED.bandwidth_limit_mb,
	entity_limit = EXCLUDED.entity_limit,
	relay_enabled = EXCLUDED.relay_enabled,
	updated_at = NOW()
`, sub.WorkspaceID, sub.StripeCustomerID, sub.StripeSubscriptionID, sub.Plan, sub.Status, sub.CurrentPeriodEnd,
		sub.RelayTier, sub.BandwidthLimitMB, sub.EntityLimit, sub.RelayEnabled)
	return err
}

func (s *PostgresStore) GetWorkspaceEmail(ctx context.Context, workspaceID string) (string, error) {
	var email string
	err := s.pool.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, workspaceID).Scan(&email)
	return email, err
}
