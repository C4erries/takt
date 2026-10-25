-- 000016_tag_scopes.down.sql

DROP INDEX IF EXISTS idx_tags_project_id;
DROP INDEX IF EXISTS idx_tags_workspace_project_name;

ALTER TABLE tags DROP COLUMN IF EXISTS project_id;

ALTER TABLE tags ADD CONSTRAINT tags_workspace_id_name_key UNIQUE (workspace_id, name);
