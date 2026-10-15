-- name: CreateAttachment :one
INSERT INTO attachments (
    workspace_id,
    task_id,
    file_name,
    file_size,
    content_type,
    s3_key,
    created_by_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetAttachmentByID :one
SELECT * FROM attachments
WHERE id = $1 AND workspace_id = $2;

-- name: ListAttachmentsByTaskID :many
SELECT * FROM attachments
WHERE task_id = $1 AND workspace_id = $2
ORDER BY created_at ASC;

-- name: DeleteAttachment :exec
DELETE FROM attachments
WHERE id = $1 AND workspace_id = $2;

-- name: DeleteAttachmentsByTaskID :exec
DELETE FROM attachments
WHERE task_id = $1 AND workspace_id = $2;

