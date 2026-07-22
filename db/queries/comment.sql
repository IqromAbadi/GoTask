-- Comment queries

-- name: CreateComment :one
INSERT INTO task_comments (task_id, user_id, content)
VALUES ($1, $2, $3)
RETURNING id, task_id, user_id, content, created_at, updated_at, deleted_at;

-- name: GetCommentByID :one
SELECT id, task_id, user_id, content, created_at, updated_at, deleted_at
FROM task_comments
WHERE id = $1 AND task_id = $2 AND deleted_at IS NULL;

-- name: ListComments :many
SELECT id, task_id, user_id, content, created_at, updated_at, deleted_at
FROM task_comments
WHERE task_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: UpdateComment :one
UPDATE task_comments
SET content = $3, updated_at = NOW()
WHERE id = $1 AND task_id = $2 AND deleted_at IS NULL
RETURNING id, task_id, user_id, content, created_at, updated_at, deleted_at;

-- name: SoftDeleteComment :exec
UPDATE task_comments
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND task_id = $2 AND deleted_at IS NULL;
