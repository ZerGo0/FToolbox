-- Add new fields to tags table
ALTER TABLE tags ADD COLUMN name TEXT NOT NULL DEFAULT '';
ALTER TABLE tags ADD COLUMN is_tracked INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tags ADD COLUMN last_checked_at INTEGER;

-- Add change field to tag_history table
ALTER TABLE tag_history ADD COLUMN change INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tag_history ADD COLUMN recorded_at INTEGER NOT NULL DEFAULT (unixepoch());

-- Update existing records with sensible defaults
UPDATE tags SET name = tag WHERE name = '';

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_tags_is_tracked ON tags(is_tracked);
CREATE INDEX IF NOT EXISTS idx_tags_last_checked_at ON tags(last_checked_at);
CREATE INDEX IF NOT EXISTS idx_tag_history_recorded_at ON tag_history(recorded_at);