-- name: ListWorkspaceAutomations :many
SELECT * FROM automations
WHERE workspace_id = $1
ORDER BY created_at ASC;

-- name: ListAutomations :many
SELECT * FROM automations
WHERE workspace_id = $1
ORDER BY created_at ASC;

-- name: GetAutomationByID :one
SELECT * FROM automations
WHERE id = $1 AND workspace_id = $2;

-- name: GetAutomation :one
SELECT * FROM automations
WHERE id = $1 AND workspace_id = $2;

-- name: ToggleAutomation :one
UPDATE automations
SET is_enabled = $3,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: UpdateAutomation :one
UPDATE automations
SET name = $3,
    trigger_type = $4,
    conditions_json = $5,
    actions_json = $6,
    is_enabled = $7,
    rule_definition = $8,
    board_id = $9,
    column_id = $10,
    task_id = $11,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: CreateAutomation :one
INSERT INTO automations (
    workspace_id,
    board_id,
    column_id,
    task_id,
    name,
    trigger_type,
    conditions_json,
    actions_json,
    is_enabled,
    rule_definition
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: CreateColumnAutomation :one
INSERT INTO automations (
    workspace_id,
    board_id,
    column_id,
    task_id,
    name,
    trigger_type,
    conditions_json,
    actions_json,
    is_enabled,
    rule_definition
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: ListAutomationsByColumn :many
SELECT * FROM automations
WHERE column_id = $1 AND workspace_id = $2
ORDER BY created_at ASC;

-- name: ListAutomationsByTask :many
SELECT * FROM automations
WHERE task_id = $1 AND workspace_id = $2
ORDER BY created_at ASC;

-- name: ListEnabledAutomationsByTrigger :many
SELECT * FROM automations
WHERE workspace_id = $1 
  AND (trigger_type = $2 OR ($2 = 'task_updated' AND trigger_type = 'task_field_changed'))
  AND is_enabled = true
ORDER BY created_at ASC;

-- name: ListAllEnabledTimerAutomations :many
SELECT * FROM automations
WHERE trigger_type IN ('column_sla_timer', 'scheduled_sweep', 'task_timer') AND is_enabled = true;

-- name: DeleteAutomation :exec
DELETE FROM automations
WHERE id = $1 AND workspace_id = $2;
