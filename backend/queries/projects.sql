-- name: CreateProject :one
INSERT INTO projects (
    workspace_id,
    name,
    description,
    color
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects
WHERE id = $1 AND workspace_id = $2;

-- name: ListProjectsByWorkspace :many
SELECT * FROM projects
WHERE workspace_id = $1
ORDER BY created_at ASC;

-- name: UpdateProject :one
UPDATE projects
SET name = $3,
    description = $4,
    color = $5,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1 AND workspace_id = $2;

-- name: CreateBoard :one
INSERT INTO boards (
    project_id,
    name,
    position
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetBoardByID :one
SELECT * FROM boards
WHERE id = $1;

-- name: ListBoardsByProject :many
SELECT * FROM boards
WHERE project_id = $1
ORDER BY position ASC, created_at ASC;

-- name: UpdateBoard :one
UPDATE boards
SET name = COALESCE($2, name),
    position = COALESCE($3, position),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteBoard :exec
DELETE FROM boards
WHERE id = $1;

-- name: CreateColumn :one
INSERT INTO columns (
    board_id,
    name,
    position,
    status_category,
    is_inbox
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetColumnByID :one
SELECT * FROM columns
WHERE id = $1;

-- name: ListColumnsByBoard :many
SELECT * FROM columns
WHERE board_id = $1
ORDER BY position ASC, created_at ASC;

-- name: UpdateColumn :one
UPDATE columns
SET name = $2,
    position = $3,
    status_category = $4,
    is_inbox = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ResetBoardInboxColumns :exec
UPDATE columns
SET is_inbox = false,
    updated_at = NOW()
WHERE board_id = $1 AND is_inbox = true;

-- name: MoveColumnToBoard :one
UPDATE columns
SET board_id = $2,
    position = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteColumn :exec
DELETE FROM columns
WHERE id = $1;

-- name: CountTasksInColumn :one
SELECT COUNT(*)::BIGINT AS task_count
FROM tasks
WHERE column_id = $1;

-- name: CountActiveTasksByBoard :one
SELECT COUNT(*) FILTER (WHERE is_completed = false AND parent_id IS NULL)::BIGINT AS active_tasks_count
FROM tasks
WHERE column_id IN (SELECT id FROM columns WHERE board_id = $1);

-- name: MoveTaskToColumnAndPosition :one
UPDATE tasks
SET column_id = $1,
    project_id = $2,
    position = $3,
    updated_at = NOW()
WHERE id = $4 AND workspace_id = $5
RETURNING *;

-- name: ResetProjectTasksToInbox :exec
UPDATE tasks
SET project_id = NULL,
    column_id = NULL,
    updated_at = NOW()
WHERE project_id = $1 AND workspace_id = $2;

-- name: ResetBoardTasksToInbox :exec
UPDATE tasks
SET project_id = NULL,
    column_id = NULL,
    updated_at = NOW()
WHERE column_id IN (SELECT id FROM columns WHERE board_id = $1) AND workspace_id = $2;

-- name: ListTasksByProject :many
SELECT * FROM tasks
WHERE project_id = $1 AND workspace_id = $2 AND parent_id IS NULL
ORDER BY position ASC, created_at DESC;

-- name: ListTasksByColumn :many
SELECT * FROM tasks
WHERE column_id = $1 AND workspace_id = $2 AND parent_id IS NULL
ORDER BY position ASC, created_at DESC;

