-- Task queries

-- name: CreateTask :one
INSERT INTO tasks (list_id, created_by, title, description, status, priority, due_date, estimated_minutes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, list_id, created_by, title, description, status, priority, progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at;

-- name: GetTaskByID :one
SELECT t.id, t.list_id, t.created_by, t.title, t.description, t.status, t.priority, t.progress, t.due_date, t.estimated_minutes, t.started_at, t.completed_at, t.created_at, t.updated_at, t.deleted_at
FROM tasks t
JOIN task_lists tl ON tl.id = t.list_id
WHERE t.id = $1 AND tl.user_id = $2 AND t.deleted_at IS NULL;

-- name: UpdateTask :one
UPDATE tasks
SET title = $3, description = $4, priority = $5, due_date = $6, estimated_minutes = $7, updated_at = NOW()
WHERE id = $1 AND list_id = $2
RETURNING id, list_id, created_by, title, description, status, priority, progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = $3, updated_at = NOW(), started_at = COALESCE(started_at, CASE WHEN $3 = 'in_progress' THEN NOW() ELSE started_at END)
WHERE id = $1 AND list_id = $2
RETURNING id, list_id, created_by, title, description, status, priority, progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at;

-- name: UpdateTaskProgress :one
UPDATE tasks
SET progress = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, list_id, created_by, title, description, status, priority, progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at;

-- name: SoftDeleteTask :exec
UPDATE tasks
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND list_id = $2;

-- name: MarkTaskDone :one
UPDATE tasks
SET status = 'done', progress = 100, completed_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING id, list_id, created_by, title, description, status, priority, progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at;

-- name: ReopenTask :one
UPDATE tasks
SET status = 'in_progress', completed_at = NULL, updated_at = NOW()
WHERE id = $1 AND list_id = $2
RETURNING id, list_id, created_by, title, description, status, priority, progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at;
