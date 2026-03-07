-- +goose Up
ALTER TABLE integrations
  ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';

UPDATE integrations
SET event_subscriptions = '[]'::jsonb
WHERE event_subscriptions IS NULL;

-- +goose Down
ALTER TABLE integrations
  DROP COLUMN IF EXISTS name;
