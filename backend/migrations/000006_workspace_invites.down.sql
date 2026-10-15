-- 000006_workspace_invites.down.sql

DROP TABLE IF EXISTS workspace_invites;

ALTER TABLE workspaces ADD CONSTRAINT workspaces_type_check CHECK (type IN ('personal', 'shared'));

