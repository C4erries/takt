-- 000008_column_automations_and_inbox.down.sql

DROP INDEX IF EXISTS idx_automations_column;
ALTER TABLE automations DROP COLUMN IF EXISTS column_id;

DROP INDEX IF EXISTS idx_columns_board_inbox;
ALTER TABLE columns DROP COLUMN IF EXISTS is_inbox;
