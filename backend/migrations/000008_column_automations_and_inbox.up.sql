-- 000008_column_automations_and_inbox.up.sql

ALTER TABLE columns ADD COLUMN is_inbox BOOLEAN NOT NULL DEFAULT FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_board_inbox ON columns(board_id) WHERE is_inbox = TRUE;

ALTER TABLE automations ADD COLUMN column_id UUID NULL REFERENCES columns(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_automations_column ON automations(column_id);
