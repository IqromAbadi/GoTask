-- Dashboard queries

-- name: DashboardSummary :one
SELECT
    COUNT(*) FILTER (WHERE t.deleted_at IS NULL) AS total_tasks,
    COUNT(*) FILTER (WHERE t.status = 'backlog' AND t.deleted_at IS NULL) AS backlog,
    COUNT(*) FILTER (WHERE t.status = 'todo' AND t.deleted_at IS NULL) AS todo,
    COUNT(*) FILTER (WHERE t.status = 'in_progress' AND t.deleted_at IS NULL) AS in_progress,
    COUNT(*) FILTER (WHERE t.status = 'review' AND t.deleted_at IS NULL) AS review,
    COUNT(*) FILTER (WHERE t.status = 'done' AND t.deleted_at IS NULL) AS done,
    COUNT(*) FILTER (WHERE t.due_date < CURRENT_DATE AND t.status != 'done' AND t.deleted_at IS NULL) AS overdue,
    COALESCE(AVG(t.progress) FILTER (WHERE t.deleted_at IS NULL), 0)::INTEGER AS average_progress
FROM tasks t
JOIN task_lists tl ON tl.id = t.list_id
WHERE tl.user_id = $1;

-- name: DashboardSummaryByList :one
SELECT
    COUNT(*) FILTER (WHERE t.deleted_at IS NULL) AS total_tasks,
    COUNT(*) FILTER (WHERE t.status = 'backlog' AND t.deleted_at IS NULL) AS backlog,
    COUNT(*) FILTER (WHERE t.status = 'todo' AND t.deleted_at IS NULL) AS todo,
    COUNT(*) FILTER (WHERE t.status = 'in_progress' AND t.deleted_at IS NULL) AS in_progress,
    COUNT(*) FILTER (WHERE t.status = 'review' AND t.deleted_at IS NULL) AS review,
    COUNT(*) FILTER (WHERE t.status = 'done' AND t.deleted_at IS NULL) AS done,
    COUNT(*) FILTER (WHERE t.due_date < CURRENT_DATE AND t.status != 'done' AND t.deleted_at IS NULL) AS overdue,
    COALESCE(AVG(t.progress) FILTER (WHERE t.deleted_at IS NULL), 0)::INTEGER AS average_progress
FROM tasks t
JOIN task_lists tl ON tl.id = t.list_id
WHERE tl.user_id = $1 AND t.list_id = $2;

-- name: PriorityDistribution :many
SELECT
    t.priority,
    COUNT(*) AS count
FROM tasks t
JOIN task_lists tl ON tl.id = t.list_id
WHERE tl.user_id = $1 AND t.deleted_at IS NULL
GROUP BY t.priority;

-- name: UpcomingDeadlines :many
SELECT t.id, t.list_id, t.created_by, t.title, t.description, t.status, t.priority,
       t.progress, t.due_date, t.estimated_minutes, t.started_at, t.completed_at,
       t.created_at, t.updated_at, t.deleted_at
FROM tasks t
JOIN task_lists tl ON tl.id = t.list_id
WHERE tl.user_id = $1
  AND t.due_date IS NOT NULL
  AND t.due_date >= CURRENT_DATE
  AND t.status != 'done'
  AND t.deleted_at IS NULL
ORDER BY t.due_date ASC
LIMIT $2;

-- name: OverdueTasks :many
SELECT t.id, t.list_id, t.created_by, t.title, t.description, t.status, t.priority,
       t.progress, t.due_date, t.estimated_minutes, t.started_at, t.completed_at,
       t.created_at, t.updated_at, t.deleted_at
FROM tasks t
JOIN task_lists tl ON tl.id = t.list_id
WHERE tl.user_id = $1
  AND t.due_date < CURRENT_DATE
  AND t.status != 'done'
  AND t.deleted_at IS NULL
ORDER BY t.due_date ASC;

-- name: ProgressAnalytics :many
SELECT
    DATE_TRUNC($2, t.created_at) AS period,
    AVG(t.progress)::INTEGER AS avg_progress,
    COUNT(*) AS tasks_count
FROM tasks t
JOIN task_lists tl ON tl.id = t.list_id
WHERE tl.user_id = $1 AND t.deleted_at IS NULL
GROUP BY period
ORDER BY period DESC
LIMIT 30;
