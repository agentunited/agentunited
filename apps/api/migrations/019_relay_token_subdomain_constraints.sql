-- +goose Up
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS relay_token VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_subscriptions_relay_token
    ON subscriptions(relay_token);

ALTER TABLE subscriptions
    ADD CONSTRAINT subscriptions_relay_subdomain_unique UNIQUE (relay_subdomain);

-- +goose Down
ALTER TABLE subscriptions
    DROP CONSTRAINT IF EXISTS subscriptions_relay_subdomain_unique;

DROP INDEX IF EXISTS idx_subscriptions_relay_token;

ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS relay_token;
