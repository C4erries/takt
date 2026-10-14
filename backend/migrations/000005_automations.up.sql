-- 000005_automations.up.sql

CREATE TABLE IF NOT EXISTS automations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    board_id UUID REFERENCES boards(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    trigger_type VARCHAR(64) NOT NULL,
    conditions_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    actions_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_automations_ws_trigger ON automations(workspace_id, trigger_type, is_enabled);

-- Seed default 3 preset automations for all existing workspaces
INSERT INTO automations (workspace_id, name, trigger_type, conditions_json, actions_json, is_enabled)
SELECT 
    id AS workspace_id,
    'Done & Archive',
    'task_completed',
    '{}'::jsonb,
    '{"action": "move_to_done"}'::jsonb,
    true
FROM workspaces;

INSERT INTO automations (workspace_id, name, trigger_type, conditions_json, actions_json, is_enabled)
SELECT 
    id AS workspace_id,
    'Auto-Focus',
    'task_updated',
    '{"has_due_date": true, "due_date_le_today": true}'::jsonb,
    '{"action": "set_focus"}'::jsonb,
    true
FROM workspaces;

INSERT INTO automations (workspace_id, name, trigger_type, conditions_json, actions_json, is_enabled)
SELECT 
    id AS workspace_id,
    'Inbox Zero',
    'task_updated',
    '{"project_assigned": true}'::jsonb,
    '{"action": "move_to_first_column"}'::jsonb,
    false
FROM workspaces;

