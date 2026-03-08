-- +goose Up
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL,
    display_name TEXT,
    stripe_customer_id TEXT UNIQUE,
    stripe_subscription_id TEXT,
    subscription_status TEXT NOT NULL DEFAULT 'free' CHECK (subscription_status IN ('free','active','past_due','canceled')),
    plan TEXT NOT NULL DEFAULT 'free' CHECK (plan IN ('free','pro','team','enterprise')),
    product_type TEXT NOT NULL DEFAULT 'self_hosted' CHECK (product_type IN ('self_hosted','managed')),
    entity_limit INTEGER NOT NULL DEFAULT 3,
    relay_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    subscription_period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS relay_tunnels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    subdomain TEXT UNIQUE NOT NULL,
    instance_url TEXT,
    connected BOOLEAN NOT NULL DEFAULT FALSE,
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL,
    stripe_event_id TEXT UNIQUE,
    payload JSONB,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customers_stripe_customer ON customers(stripe_customer_id);
CREATE INDEX IF NOT EXISTS idx_customers_stripe_subscription ON customers(stripe_subscription_id);
CREATE INDEX IF NOT EXISTS idx_relay_tunnels_customer ON relay_tunnels(customer_id);
CREATE INDEX IF NOT EXISTS idx_billing_events_customer ON billing_events(customer_id);
CREATE INDEX IF NOT EXISTS idx_billing_events_type ON billing_events(event_type);

-- +goose Down
DROP TABLE IF EXISTS billing_events;
DROP TABLE IF EXISTS relay_tunnels;
DROP TABLE IF EXISTS customers;
