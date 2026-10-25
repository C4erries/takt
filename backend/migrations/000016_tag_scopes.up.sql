-- 000016_tag_scopes.up.sql

ALTER TABLE tags ADD COLUMN project_id UUID NULL REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE tags DROP CONSTRAINT IF EXISTS tags_workspace_id_name_key;
DROP INDEX IF EXISTS idx_tags_workspace_name;

CREATE UNIQUE INDEX idx_tags_workspace_project_name 
  ON tags(workspace_id, COALESCE(project_id, '00000000-0000-0000-0000-000000000000'), LOWER(name));

CREATE INDEX idx_tags_project_id ON tags(project_id) WHERE project_id IS NOT NULL;
