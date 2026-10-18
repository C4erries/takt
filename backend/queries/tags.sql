-- name: CreateTag :one
INSERT INTO tags (
    workspace_id,
    name,
    color_hex
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetTagByID :one
SELECT * FROM tags
WHERE id = $1 AND workspace_id = $2;

-- name: ListWorkspaceTags :many
SELECT * FROM tags
WHERE workspace_id = $1
ORDER BY name ASC;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE id = $1 AND workspace_id = $2;

-- name: AddTagToTask :exec
INSERT INTO task_tags (
    task_id,
    tag_id
) VALUES (
    $1, $2
) ON CONFLICT (task_id, tag_id) DO NOTHING;

-- name: RemoveTagFromTask :exec
DELETE FROM task_tags
WHERE task_id = $1 AND tag_id = $2;

-- name: GetTaskTags :many
SELECT t.* FROM tags t
JOIN task_tags tt ON t.id = tt.tag_id
WHERE tt.task_id = $1
ORDER BY t.name ASC;

-- name: ClearTaskTags :exec
DELETE FROM task_tags
WHERE task_id = $1;

-- name: GetTagByName :one
SELECT * FROM tags
WHERE workspace_id = $1 AND LOWER(name) = LOWER($2)
LIMIT 1;

-- name: ListTagsForTasks :many
SELECT tt.task_id, t.id, t.workspace_id, t.name, t.color_hex, t.created_at
FROM tags t
JOIN task_tags tt ON t.id = tt.tag_id
WHERE tt.task_id = ANY($1::uuid[])
ORDER BY t.name ASC;

