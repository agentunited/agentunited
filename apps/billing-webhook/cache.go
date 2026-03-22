package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	AlreadyProcessed(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, eventID string, ttl time.Duration) error
	UpdateRelayPlan(ctx context.Context, workspaceID string, sub SubscriptionRecord) error
}

type RedisCache struct{ client *redis.Client }

func NewRedisCache(redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return &RedisCache{client: redis.NewClient(opts)}, nil
}

func (c *RedisCache) Close() error { return c.client.Close() }

func (c *RedisCache) AlreadyProcessed(ctx context.Context, eventID string) (bool, error) {
	key := "webhook:idempotency:" + eventID
	ok, err := c.client.Exists(ctx, key).Result()
	return ok > 0, err
}

func (c *RedisCache) MarkProcessed(ctx context.Context, eventID string, ttl time.Duration) error {
	key := "webhook:idempotency:" + eventID
	return c.client.Set(ctx, key, "1", ttl).Err()
}

func (c *RedisCache) UpdateRelayPlan(ctx context.Context, workspaceID string, sub SubscriptionRecord) error {
	key := "relay:workspace:" + workspaceID
	payload := map[string]any{
		"workspace_id":        workspaceID,
		"plan":                sub.Plan,
		"status":              sub.Status,
		"relay_tier":          sub.RelayTier,
		"bandwidth_limit_mb":  sub.BandwidthLimitMB,
		"entity_limit":        sub.EntityLimit,
		"relay_enabled":       sub.RelayEnabled,
		"stripe_customer_id":  sub.StripeCustomerID,
		"stripe_subscription": sub.StripeSubscriptionID,
	}
	b, _ := json.Marshal(payload)
	return c.client.Set(ctx, key, b, 0).Err()
}
