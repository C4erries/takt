-- 000012_board_positions.up.sql

ALTER TABLE boards ADD COLUMN IF NOT EXISTS position DOUBLE PRECISION NOT NULL DEFAULT 1000.0;

UPDATE boards SET position = 1000.0 WHERE position IS NULL;

CREATE INDEX IF NOT EXISTS idx_boards_project_position ON boards(project_id, position);

