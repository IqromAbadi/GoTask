-- Task progress queries

-- name: CreateProgressUpdate :one
INSERT INTO task_progress_updates (task_id, user_id, progress, note)
VALUES ($1, $2, $3, $4)
RETURNING id, task_id, user_id, progress, note, created_at, updated_at;

-- name: ListProgressUpdates :many
SELECT id, task_id, user_id, progress, note, created_at, updated_at
FROM task_progress_updates
WHERE task_id = $1
ORDER BY created_at DESC;

-- name: GetProgressUpdateByID :one
SELECT id, task_id, user_id, progress, note, created_at, updated_at
FROM task_progress_updates
WHERE id = $1 AND task_id = $2;

-- name: UpdateProgressNote :one
UPDATE task_progress_updates
SET note = $3, updated_at = NOW()
WHERE id = $1 AND task_id = $2
RETURNING id, task_id, user_id, progress, note, created_at, updated_at;

-- name: DeleteProgressUpdate :exec
DELETE FROM task_progress_updates
WHERE id = $1 AND task_id = $2;

-- name: GetLatestProgress :one
SELECT progress
FROM task_progress_updates
WHERE task_id = $1
ORDER BY created_at DESC
LIMIT 1;
