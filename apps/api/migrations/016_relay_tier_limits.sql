-- +goose Up
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS relay_tier TEXT NOT NULL DEFAULT 'free',
    ADD COLUMN IF NOT EXISTS relay_bandwidth_used_mb INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS relay_bandwidth_limit_mb INTEGER NOT NULL DEFAULT 1024,
    ADD COLUMN IF NOT EXISTS relay_connections_max INTEGER NOT NULL DEFAULT 3,
    ADD COLUMN IF NOT EXISTS relay_custom_subdomain BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS relay_subdomain TEXT,
    ADD COLUMN IF NOT EXISTS relay_expires_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS relay_expires_at,
    DROP COLUMN IF EXISTS relay_subdomain,
    DROP COLUMN IF EXISTS relay_custom_subdomain,
    DROP COLUMN IF EXISTS relay_connections_max,
    DROP COLUMN IF EXISTS relay_bandwidth_limit_mb,
    DROP COLUMN IF EXISTS relay_bandwidth_used_mb,
    DROP COLUMN IF EXISTS relay_tier;
