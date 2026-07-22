-- Task List queries

-- name: CreateTaskList :one
INSERT INTO task_lists (user_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, user_id, name, description, is_archived, created_at, updated_at;

-- name: GetTaskListByID :one
SELECT id, user_id, name, description, is_archived, created_at, updated_at
FROM task_lists
WHERE id = $1 AND user_id = $2;

-- name: ListTaskLists :many
SELECT id, user_id, name, description, is_archived, created_at, updated_at
FROM task_lists
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateTaskList :one
UPDATE task_lists
SET name = $3, description = $4, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, description, is_archived, created_at, updated_at;

-- name: DeleteTaskList :exec
DELETE FROM task_lists
WHERE id = $1 AND user_id = $2;

-- name: ArchiveTaskList :one
UPDATE task_lists
SET is_archived = TRUE, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, description, is_archived, created_at, updated_at;

-- name: RestoreTaskList :one
UPDATE task_lists
SET is_archived = FALSE, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, description, is_archived, created_at, updated_at;
