-- 000013_subtasks.up.sql

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS parent_id UUID NULL REFERENCES tasks(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_workspace_parent ON tasks(workspace_id, parent_id);
