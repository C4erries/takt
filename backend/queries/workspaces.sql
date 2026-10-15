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

-- name: UpdateWorkspaceName :one
UPDATE workspaces
SET name = $2
WHERE id = $1 AND owner_id = $3
RETURNING *;

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

-- name: ListWorkspaceMembers :many
SELECT
    wm.id,
    wm.user_id,
    u.email,
    u.display_name,
    wm.role,
    wm.created_at
FROM workspace_members wm
JOIN users u ON u.id = wm.user_id
WHERE wm.workspace_id = $1
ORDER BY wm.created_at ASC;

-- name: DeleteWorkspaceMember :exec
DELETE FROM workspace_members
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

-- name: CreateWorkspaceInvite :one
INSERT INTO workspace_invites (
    workspace_id,
    invited_by_id,
    token,
    expires_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetWorkspaceInviteByToken :one
SELECT
    wi.id,
    wi.workspace_id,
    wi.invited_by_id,
    wi.token,
    wi.expires_at,
    wi.created_at,
    w.name AS workspace_name,
    u.display_name AS invited_by_name
FROM workspace_invites wi
JOIN workspaces w ON w.id = wi.workspace_id
JOIN users u ON u.id = wi.invited_by_id
WHERE wi.token = $1;

-- name: DeleteWorkspaceInvite :exec
DELETE FROM workspace_invites
WHERE id = $1;

-- name: DeleteWorkspaceInviteByToken :exec
DELETE FROM workspace_invites
WHERE token = $1;

-- name: GetActiveWorkspaceInvite :one
SELECT * FROM workspace_invites
WHERE workspace_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY created_at DESC
LIMIT 1;
