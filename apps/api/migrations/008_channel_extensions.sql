-- +goose Up

-- Add type column to channels for DM support
ALTER TABLE channels ADD COLUMN type TEXT NOT NULL DEFAULT 'channel';
ALTER TABLE channels ADD CONSTRAINT chk_channel_type CHECK (type IN ('channel', 'dm'));

-- Create index for channel type filtering  
CREATE INDEX IF NOT EXISTS idx_channels_type ON channels(type);

-- Add full-text search support for messages
ALTER TABLE messages ADD COLUMN search_vector tsvector;

-- Update existing messages with search vector
UPDATE messages SET search_vector = to_tsvector('english', text);

-- Create index for full-text search
CREATE INDEX IF NOT EXISTS idx_messages_search ON messages USING gin(search_vector);

-- Create trigger to auto-update search_vector on message insert/update
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_message_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector = to_tsvector('english', NEW.text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER update_message_search_trigger
    BEFORE INSERT OR UPDATE ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_message_search_vector();

-- Add updated_at column to messages for edit tracking
ALTER TABLE messages ADD COLUMN updated_at TIMESTAMPTZ;
UPDATE messages SET updated_at = created_at;
ALTER TABLE messages ALTER COLUMN updated_at SET NOT NULL;
ALTER TABLE messages ALTER COLUMN updated_at SET DEFAULT now();

-- Create trigger to update updated_at on message edits
CREATE TRIGGER update_messages_updated_at
    BEFORE UPDATE ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down

-- Drop triggers and functions
DROP TRIGGER IF EXISTS update_message_search_trigger ON messages;
DROP TRIGGER IF EXISTS update_messages_updated_at ON messages; 
DROP FUNCTION IF EXISTS update_message_search_vector();

-- Drop indexes
DROP INDEX IF EXISTS idx_messages_search;
DROP INDEX IF EXISTS idx_channels_type;

-- Remove columns
ALTER TABLE messages DROP COLUMN IF EXISTS search_vector;
ALTER TABLE messages DROP COLUMN IF EXISTS updated_at;
ALTER TABLE channels DROP COLUMN IF EXISTS type;