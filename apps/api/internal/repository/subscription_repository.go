package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/agentunited/backend/internal/models"
)

type SubscriptionRepository interface {
	GetByWorkspace(ctx context.Context, workspaceID string) (*models.Subscription, error)
	UpsertByWorkspace(ctx context.Context, sub *models.Subscription) error
	UpsertByStripeSubscriptionID(ctx context.Context, sub *models.Subscription) error
}

type subscriptionRepository struct{ db *DB }

func NewSubscriptionRepository(db *DB) SubscriptionRepository { return &subscriptionRepository{db: db} }

func (r *subscriptionRepository) GetByWorkspace(ctx context.Context, workspaceID string) (*models.Subscription, error) {
	q := `SELECT id, workspace_id, COALESCE(stripe_customer_id,''), COALESCE(stripe_subscription_id,''), plan, status, current_period_end,
	             relay_tier, relay_bandwidth_used_mb, relay_bandwidth_limit_mb, relay_connections_max,
	             relay_custom_subdomain, COALESCE(relay_subdomain,''), relay_expires_at, created_at, updated_at
	      FROM subscriptions WHERE workspace_id = $1`
	var s models.Subscription
	var cpe *time.Time
	var relayExpiresAt *time.Time
	if err := r.db.Pool.QueryRow(ctx, q, workspaceID).Scan(
		&s.ID, &s.WorkspaceID, &s.StripeCustomerID, &s.StripeSubscriptionID, &s.Plan, &s.Status, &cpe,
		&s.RelayTier, &s.RelayBandwidthUsedMB, &s.RelayBandwidthLimitMB, &s.RelayConnectionsMax,
		&s.RelayCustomSubdomain, &s.RelaySubdomain, &relayExpiresAt, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, err
	}
	s.CurrentPeriodEnd = cpe
	s.RelayExpiresAt = relayExpiresAt
	return &s, nil
}

func (r *subscriptionRepository) UpsertByWorkspace(ctx context.Context, sub *models.Subscription) error {
	q := `INSERT INTO subscriptions (
		workspace_id, stripe_customer_id, stripe_subscription_id, plan, status, current_period_end,
		relay_tier, relay_bandwidth_used_mb, relay_bandwidth_limit_mb, relay_connections_max,
		relay_custom_subdomain, relay_subdomain, relay_expires_at, created_at, updated_at
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW(),NOW())
	ON CONFLICT (workspace_id)
	DO UPDATE SET stripe_customer_id = EXCLUDED.stripe_customer_id,
	              stripe_subscription_id = EXCLUDED.stripe_subscription_id,
	              plan = EXCLUDED.plan,
	              status = EXCLUDED.status,
	              current_period_end = EXCLUDED.current_period_end,
	              relay_tier = EXCLUDED.relay_tier,
	              relay_bandwidth_used_mb = EXCLUDED.relay_bandwidth_used_mb,
	              relay_bandwidth_limit_mb = EXCLUDED.relay_bandwidth_limit_mb,
	              relay_connections_max = EXCLUDED.relay_connections_max,
	              relay_custom_subdomain = EXCLUDED.relay_custom_subdomain,
	              relay_subdomain = EXCLUDED.relay_subdomain,
	              relay_expires_at = EXCLUDED.relay_expires_at,
	              updated_at = NOW()`
	_, err := r.db.Pool.Exec(ctx, q,
		sub.WorkspaceID,
		nullIfEmpty(sub.StripeCustomerID),
		nullIfEmpty(sub.StripeSubscriptionID),
		sub.Plan,
		sub.Status,
		sub.CurrentPeriodEnd,
		defaultIfEmpty(sub.RelayTier, "free"),
		sub.RelayBandwidthUsedMB,
		defaultIfZero(sub.RelayBandwidthLimitMB, 1024),
		defaultIfZero(sub.RelayConnectionsMax, 3),
		sub.RelayCustomSubdomain,
		nullIfEmpty(sub.RelaySubdomain),
		sub.RelayExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("upsert subscription by workspace: %w", err)
	}
	return nil
}

func (r *subscriptionRepository) UpsertByStripeSubscriptionID(ctx context.Context, sub *models.Subscription) error {
	q := `INSERT INTO subscriptions (
		workspace_id, stripe_customer_id, stripe_subscription_id, plan, status, current_period_end,
		relay_tier, relay_bandwidth_used_mb, relay_bandwidth_limit_mb, relay_connections_max,
		relay_custom_subdomain, relay_subdomain, relay_expires_at, created_at, updated_at
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW(),NOW())
	ON CONFLICT (stripe_subscription_id)
	DO UPDATE SET stripe_customer_id = EXCLUDED.stripe_customer_id,
	              plan = EXCLUDED.plan,
	              status = EXCLUDED.status,
	              current_period_end = EXCLUDED.current_period_end,
	              relay_tier = EXCLUDED.relay_tier,
	              relay_bandwidth_used_mb = EXCLUDED.relay_bandwidth_used_mb,
	              relay_bandwidth_limit_mb = EXCLUDED.relay_bandwidth_limit_mb,
	              relay_connections_max = EXCLUDED.relay_connections_max,
	              relay_custom_subdomain = EXCLUDED.relay_custom_subdomain,
	              relay_subdomain = EXCLUDED.relay_subdomain,
	              relay_expires_at = EXCLUDED.relay_expires_at,
	              updated_at = NOW()`
	_, err := r.db.Pool.Exec(ctx, q,
		sub.WorkspaceID,
		nullIfEmpty(sub.StripeCustomerID),
		nullIfEmpty(sub.StripeSubscriptionID),
		sub.Plan,
		sub.Status,
		sub.CurrentPeriodEnd,
		defaultIfEmpty(sub.RelayTier, "free"),
		sub.RelayBandwidthUsedMB,
		defaultIfZero(sub.RelayBandwidthLimitMB, 1024),
		defaultIfZero(sub.RelayConnectionsMax, 3),
		sub.RelayCustomSubdomain,
		nullIfEmpty(sub.RelaySubdomain),
		sub.RelayExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("upsert subscription by stripe subscription id: %w", err)
	}
	return nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func defaultIfEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func defaultIfZero(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}
