-- +goose Up
ALTER TABLE invites
  ADD COLUMN IF NOT EXISTS workspace_id UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_invites_workspace_id ON invites(workspace_id);

-- +goose Down
DROP INDEX IF EXISTS idx_invites_workspace_id;
ALTER TABLE invites DROP COLUMN IF EXISTS workspace_id;
