ALTER TABLE tasks ADD COLUMN is_archived BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE tasks ADD COLUMN archived_at TIMESTAMPTZ NULL;
CREATE INDEX idx_tasks_workspace_archived ON tasks(workspace_id, is_archived) WHERE is_archived = true;
