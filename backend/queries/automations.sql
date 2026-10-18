-- name: ListWorkspaceAutomations :many
SELECT * FROM automations
WHERE workspace_id = $1
ORDER BY created_at ASC;

-- name: GetAutomationByID :one
SELECT * FROM automations
WHERE id = $1 AND workspace_id = $2;

-- name: ToggleAutomation :one
UPDATE automations
SET is_enabled = $3,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

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
RETURNING *;

-- name: CreateColumnAutomation :one
INSERT INTO automations (
    workspace_id,
    board_id,
    column_id,
    name,
    trigger_type,
    conditions_json,
    actions_json,
    is_enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListAutomationsByColumn :many
SELECT * FROM automations
WHERE column_id = $1 AND workspace_id = $2
ORDER BY created_at ASC;

-- name: ListEnabledAutomationsByTrigger :many
SELECT * FROM automations
WHERE workspace_id = $1 AND trigger_type = $2 AND is_enabled = true
ORDER BY created_at ASC;
