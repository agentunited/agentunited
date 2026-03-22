-- +goose Up
-- Ensure Team plan is accepted anywhere a plan CHECK constraint exists.

-- subscriptions
ALTER TABLE IF EXISTS subscriptions
    DROP CONSTRAINT IF EXISTS subscriptions_plan_check;
ALTER TABLE IF EXISTS subscriptions
    ADD CONSTRAINT subscriptions_plan_check
    CHECK (plan IN ('free', 'pro', 'team', 'enterprise'));

-- workspaces (if present in this deployment)
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'workspaces' AND column_name = 'plan'
    ) THEN
        EXECUTE 'ALTER TABLE workspaces DROP CONSTRAINT IF EXISTS workspaces_plan_check';
        EXECUTE 'ALTER TABLE workspaces ADD CONSTRAINT workspaces_plan_check CHECK (plan IN (''free'', ''pro'', ''team'', ''enterprise''))';
    END IF;
END $$;
-- +goose StatementEnd

-- relay_token (if present in this deployment)
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'relay_token' AND column_name = 'plan'
    ) THEN
        EXECUTE 'ALTER TABLE relay_token DROP CONSTRAINT IF EXISTS relay_token_plan_check';
        EXECUTE 'ALTER TABLE relay_token ADD CONSTRAINT relay_token_plan_check CHECK (plan IN (''free'', ''pro'', ''team'', ''enterprise''))';
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE IF EXISTS subscriptions
    DROP CONSTRAINT IF EXISTS subscriptions_plan_check;
ALTER TABLE IF EXISTS subscriptions
    ADD CONSTRAINT subscriptions_plan_check
    CHECK (plan IN ('free', 'pro', 'enterprise'));

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'workspaces' AND column_name = 'plan'
    ) THEN
        EXECUTE 'ALTER TABLE workspaces DROP CONSTRAINT IF EXISTS workspaces_plan_check';
        EXECUTE 'ALTER TABLE workspaces ADD CONSTRAINT workspaces_plan_check CHECK (plan IN (''free'', ''pro'', ''enterprise''))';
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'relay_token' AND column_name = 'plan'
    ) THEN
        EXECUTE 'ALTER TABLE relay_token DROP CONSTRAINT IF EXISTS relay_token_plan_check';
        EXECUTE 'ALTER TABLE relay_token ADD CONSTRAINT relay_token_plan_check CHECK (plan IN (''free'', ''pro'', ''enterprise''))';
    END IF;
END $$;
-- +goose StatementEnd
