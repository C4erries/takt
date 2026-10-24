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
  AND is_archived = false
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
  AND parent_id IS NULL
  AND is_archived = false;

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
  AND is_archived = false
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

-- name: ListArchivedTasks :many
SELECT * FROM tasks
WHERE workspace_id = $1
  AND is_archived = true
  AND (project_id = $2 OR $2::uuid IS NULL)
ORDER BY archived_at DESC, created_at DESC;

-- name: CountArchivedTasks :one
SELECT COUNT(*)::BIGINT AS archived_tasks_count
FROM tasks
WHERE workspace_id = $1
  AND is_archived = true;

-- name: ArchiveCompletedTasksInColumn :execrows
UPDATE tasks
SET is_archived = true,
    archived_at = NOW(),
    updated_at = NOW()
WHERE column_id = $1
  AND is_completed = true
  AND is_archived = false;

-- name: UnarchiveTask :one
UPDATE tasks
SET is_archived = false,
    archived_at = NULL,
    column_id = $3,
    project_id = $4,
    position = $5,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: ArchiveCompletedTasksBatchSQL :execrows
UPDATE tasks
SET is_archived = true,
    archived_at = NOW(),
    updated_at = NOW()
WHERE workspace_id = $1
  AND column_id = $2
  AND is_completed = true
  AND is_archived = false
  AND updated_at < NOW() - ($3::text || ' hours')::interval;

