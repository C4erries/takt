-- name: CreateTask :one
INSERT INTO tasks (
    workspace_id,
    project_id,
    column_id,
    title,
    description_markdown,
    is_completed,
    completed_at,
    due_date,
    priority,
    focus_order,
    position
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetInboxTasks :many
SELECT * FROM tasks
WHERE workspace_id = $1
  AND project_id IS NULL
  AND column_id IS NULL
ORDER BY position ASC, created_at DESC;

-- name: GetTaskByID :one
SELECT * FROM tasks
WHERE id = $1 AND workspace_id = $2;

-- name: ToggleTaskCompletion :one
UPDATE tasks
SET is_completed = NOT is_completed,
    completed_at = CASE WHEN is_completed = false THEN NOW() ELSE NULL END,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1 AND workspace_id = $2;

-- name: GetMaxInboxPosition :one
SELECT COALESCE(MAX(position), -1)::INT as max_position
FROM tasks
WHERE workspace_id = $1
  AND project_id IS NULL
  AND column_id IS NULL;
