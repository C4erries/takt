-- name: SearchTasksInWorkspace :many
SELECT * FROM tasks
WHERE workspace_id = $1
  AND (title ILIKE '%' || sqlc.arg('query')::text || '%' OR description_markdown ILIKE '%' || sqlc.arg('query')::text || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count')::int;

-- name: SearchProjectsInWorkspace :many
SELECT * FROM projects
WHERE workspace_id = $1
  AND (name ILIKE '%' || sqlc.arg('query')::text || '%' OR description ILIKE '%' || sqlc.arg('query')::text || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count')::int;
