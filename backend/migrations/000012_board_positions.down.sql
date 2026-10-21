-- 000012_board_positions.down.sql

DROP INDEX IF EXISTS idx_boards_project_position;

ALTER TABLE boards DROP COLUMN IF EXISTS position;

