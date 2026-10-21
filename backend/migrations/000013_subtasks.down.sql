-- 000013_subtasks.down.sql

DROP INDEX IF EXISTS idx_tasks_workspace_parent;
DROP INDEX IF EXISTS idx_tasks_parent_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS parent_id;
