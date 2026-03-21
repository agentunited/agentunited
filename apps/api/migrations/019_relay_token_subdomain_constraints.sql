-- +goose Up
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS relay_token VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_subscriptions_relay_token
    ON subscriptions(relay_token);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'subscriptions_relay_subdomain_unique'
    ) THEN
        ALTER TABLE subscriptions
            ADD CONSTRAINT subscriptions_relay_subdomain_unique UNIQUE (relay_subdomain);
    END IF;
END$$;

-- +goose Down
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'subscriptions_relay_subdomain_unique'
    ) THEN
        ALTER TABLE subscriptions
            DROP CONSTRAINT subscriptions_relay_subdomain_unique;
    END IF;
END$$;

DROP INDEX IF EXISTS idx_subscriptions_relay_token;

ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS relay_token;
