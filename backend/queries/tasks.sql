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
    position,
    parent_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetInboxTasks :many
SELECT * FROM tasks
WHERE workspace_id = $1
  AND project_id IS NULL
  AND column_id IS NULL
  AND parent_id IS NULL
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
  AND column_id IS NULL
  AND parent_id IS NULL;

-- name: UpdateTask :one
UPDATE tasks
SET title = $3,
    description_markdown = $4,
    due_date = $5,
    priority = $6,
    focus_order = $7,
    project_id = $8,
    column_id = $9,
    parent_id = COALESCE($10, parent_id),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: GetTodayTasks :many
SELECT * FROM tasks
WHERE workspace_id = $1
  AND is_completed = FALSE
  AND parent_id IS NULL
  AND (due_date <= CURRENT_DATE OR focus_order IS NOT NULL)
ORDER BY focus_order ASC NULLS LAST, due_date ASC, priority DESC;

-- name: SetTaskFocus :one
UPDATE tasks
SET focus_order = $3,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: SetTaskPriority :one
UPDATE tasks
SET priority = $3,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: SetTaskDueDate :one
UPDATE tasks
SET due_date = $3,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: SetTaskCompletion :one
UPDATE tasks
SET is_completed = $3,
    completed_at = CASE WHEN $3 = true THEN NOW() ELSE NULL END,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: ListSubtasksByParent :many
SELECT * FROM tasks
WHERE parent_id = $1
ORDER BY position ASC, created_at ASC;

-- name: GetSubtaskCountsByParentIds :many
SELECT parent_id, COUNT(*)::bigint AS total_count, COUNT(*) FILTER (WHERE is_completed = true)::bigint AS completed_count
FROM tasks
WHERE parent_id = ANY($1::uuid[])
GROUP BY parent_id;

-- name: DuplicateTaskRow :one
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
    position,
    parent_id
) VALUES (
    $1, $2, $3, $4, $5, false, NULL, NULL, $6, NULL, $7, $8
)
RETURNING *;

