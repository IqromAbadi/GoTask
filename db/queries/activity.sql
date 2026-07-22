-- Activity queries

-- name: CreateActivity :one
INSERT INTO task_activities (task_id, user_id, action, old_value, new_value, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, task_id, user_id, action, old_value, new_value, metadata, created_at;

-- name: ListTaskActivities :many
SELECT id, task_id, user_id, action, old_value, new_value, metadata, created_at
FROM task_activities
WHERE task_id = $1
ORDER BY created_at DESC;

-- name: ListUserActivities :many
SELECT a.id, a.task_id, a.user_id, a.action, a.old_value, a.new_value, a.metadata, a.created_at
FROM task_activities a
JOIN tasks t ON t.id = a.task_id
JOIN task_lists tl ON tl.id = t.list_id
WHERE tl.user_id = $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserActivities :one
SELECT COUNT(*)
FROM task_activities a
JOIN tasks t ON t.id = a.task_id
JOIN task_lists tl ON tl.id = t.list_id
WHERE tl.user_id = $1;
