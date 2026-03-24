-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS central_user_id TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_type TEXT NOT NULL DEFAULT 'local';
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_central_user_id ON users(central_user_id) WHERE central_user_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_central_user_id;
ALTER TABLE users DROP COLUMN IF EXISTS auth_type;
ALTER TABLE users DROP COLUMN IF EXISTS central_user_id;
