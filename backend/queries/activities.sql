-- name: CreateTaskActivity :one
INSERT INTO task_activities (
    id,
    task_id,
    user_id,
    activity_type,
    payload,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListTaskActivitiesByTaskID :many
SELECT 
    ta.id,
    ta.task_id,
    ta.user_id,
    ta.activity_type,
    ta.payload,
    ta.created_at,
    u.email AS user_email,
    u.display_name AS user_name
FROM task_activities ta
LEFT JOIN users u ON u.id = ta.user_id
WHERE ta.task_id = $1
ORDER BY ta.created_at DESC;
