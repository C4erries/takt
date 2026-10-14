-- name: CreateWorkspace :one
INSERT INTO workspaces (
    owner_id,
    name,
    type
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetWorkspaceByID :one
SELECT * FROM workspaces
WHERE id = $1;

-- name: CreateWorkspaceMember :one
INSERT INTO workspace_members (
    workspace_id,
    user_id,
    role
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetWorkspaceMember :one
SELECT * FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: GetWorkspacesByUserID :many
SELECT w.*, wm.role
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = $1
ORDER BY w.created_at ASC;

-- name: GetPersonalWorkspaceByUserID :one
SELECT w.*, wm.role
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = $1 AND w.type = 'personal'
LIMIT 1;
