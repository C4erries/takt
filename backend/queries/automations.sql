-- name: ListWorkspaceAutomations :many
SELECT id, workspace_id, board_id, name, trigger_type, conditions_json, actions_json, is_enabled, created_at, updated_at
FROM automations
WHERE workspace_id = $1
ORDER BY created_at ASC;

-- name: GetAutomationByID :one
SELECT id, workspace_id, board_id, name, trigger_type, conditions_json, actions_json, is_enabled, created_at, updated_at
FROM automations
WHERE id = $1 AND workspace_id = $2;

-- name: ToggleAutomation :one
UPDATE automations
SET is_enabled = $3,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING id, workspace_id, board_id, name, trigger_type, conditions_json, actions_json, is_enabled, created_at, updated_at;

-- name: CreateAutomation :one
INSERT INTO automations (
    workspace_id,
    board_id,
    name,
    trigger_type,
    conditions_json,
    actions_json,
    is_enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, workspace_id, board_id, name, trigger_type, conditions_json, actions_json, is_enabled, created_at, updated_at;

-- name: ListEnabledAutomationsByTrigger :many
SELECT id, workspace_id, board_id, name, trigger_type, conditions_json, actions_json, is_enabled, created_at, updated_at
FROM automations
WHERE workspace_id = $1 AND trigger_type = $2 AND is_enabled = true
ORDER BY created_at ASC;

