package task

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL task repository.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, t *Task) error {
	query := `
		INSERT INTO tasks (list_id, created_by, title, description, status, priority, due_date, estimated_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		t.ListID, t.CreatedBy, t.Title, t.Description, t.Status, t.Priority, t.DueDate, t.EstimatedMin,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)

	t.CreatedAt = t.CreatedAt.UTC()
	t.UpdatedAt = t.UpdatedAt.UTC()

	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*Task, error) {
	query := `
		SELECT t.id, t.list_id, t.created_by, t.title, t.description, t.status, t.priority, 
		       t.progress, t.due_date, t.estimated_minutes, t.started_at, t.completed_at, 
		       t.created_at, t.updated_at, t.deleted_at
		FROM tasks t
		JOIN task_lists tl ON tl.id = t.list_id
		WHERE t.id = $1 AND tl.user_id = $2 AND t.deleted_at IS NULL`

	task := &Task{}
	var desc sql.NullString
	var dueDate sql.NullTime
	var estimatedMin sql.NullInt64
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&task.ID, &task.ListID, &task.CreatedBy, &task.Title, &desc,
		&task.Status, &task.Priority, &task.Progress, &dueDate, &estimatedMin,
		&startedAt, &completedAt, &task.CreatedAt, &task.UpdatedAt, &deletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get task by id: %w", err)
	}

	if desc.Valid {
		task.Description = &desc.String
	}
	if dueDate.Valid {
		t := dueDate.Time.UTC()
		task.DueDate = &t
	}
	if estimatedMin.Valid {
		v := int(estimatedMin.Int64)
		task.EstimatedMin = &v
	}
	if startedAt.Valid {
		t := startedAt.Time.UTC()
		task.StartedAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time.UTC()
		task.CompletedAt = &t
	}

	task.CreatedAt = task.CreatedAt.UTC()
	task.UpdatedAt = task.UpdatedAt.UTC()

	return task, nil
}

func (r *postgresRepository) Update(ctx context.Context, t *Task) error {
	query := `
		UPDATE tasks
		SET title = $3, description = $4, priority = $5, due_date = $6, estimated_minutes = $7, updated_at = NOW()
		WHERE id = $1 AND list_id = $2
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		t.ID, t.ListID, t.Title, t.Description, t.Priority, t.DueDate, t.EstimatedMin,
	).Scan(&t.UpdatedAt)
	t.UpdatedAt = t.UpdatedAt.UTC()

	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateStatus(ctx context.Context, id, listID uuid.UUID, status string) (*Task, error) {
	query := `
		UPDATE tasks
		SET status = $3, updated_at = NOW(), 
		    started_at = COALESCE(started_at, CASE WHEN $3 = 'in_progress' THEN NOW() ELSE started_at END)
		WHERE id = $1 AND list_id = $2
		RETURNING id, list_id, created_by, title, description, status, priority, 
		          progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at`

	return r.scanTask(ctx, query, id, listID, status)
}

func (r *postgresRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) (*Task, error) {
	query := `
		UPDATE tasks
		SET progress = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, list_id, created_by, title, description, status, priority, 
		          progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at`

	return r.scanTaskSimple(ctx, query, id, progress)
}

func (r *postgresRepository) SoftDelete(ctx context.Context, id, listID uuid.UUID) error {
	query := `UPDATE tasks SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND list_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, listID)
	if err != nil {
		return fmt.Errorf("soft delete task: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func (r *postgresRepository) MarkDone(ctx context.Context, id uuid.UUID) (*Task, error) {
	query := `
		UPDATE tasks
		SET status = 'done', progress = 100, completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id, list_id, created_by, title, description, status, priority, 
		          progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at`

	return r.scanTaskSimple(ctx, query, id)
}

func (r *postgresRepository) Reopen(ctx context.Context, id, listID uuid.UUID) (*Task, error) {
	query := `
		UPDATE tasks
		SET status = 'in_progress', completed_at = NULL, updated_at = NOW()
		WHERE id = $1 AND list_id = $2
		RETURNING id, list_id, created_by, title, description, status, priority, 
		          progress, due_date, estimated_minutes, started_at, completed_at, created_at, updated_at, deleted_at`

	return r.scanTask(ctx, query, id, listID)
}

func (r *postgresRepository) List(ctx context.Context, listID, userID uuid.UUID, filter TaskFilter) ([]Task, int, error) {
	// Base query
	where := `WHERE t.list_id = $1 AND tl.user_id = $2 AND t.deleted_at IS NULL`
	args := []any{listID, userID}
	argIdx := 3

	if filter.Status != "" {
		where += fmt.Sprintf(` AND t.status = $%d`, argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Priority != "" {
		where += fmt.Sprintf(` AND t.priority = $%d`, argIdx)
		args = append(args, filter.Priority)
		argIdx++
	}
	if filter.Search != "" {
		where += fmt.Sprintf(` AND t.title ILIKE $%d`, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}
	if filter.DueDateFrom != nil {
		where += fmt.Sprintf(` AND t.due_date >= $%d`, argIdx)
		args = append(args, filter.DueDateFrom)
		argIdx++
	}
	if filter.DueDateTo != nil {
		where += fmt.Sprintf(` AND t.due_date <= $%d`, argIdx)
		args = append(args, filter.DueDateTo)
		argIdx++
	}
	if filter.IsOverdue {
		where += ` AND t.due_date < CURRENT_DATE AND t.status != 'done'`
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM tasks t JOIN task_lists tl ON tl.id = t.list_id ` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	// Sorting (whitelist to prevent SQL injection)
	sortBy := "t.created_at"
	if filter.SortBy != "" && ValidSortFields[filter.SortBy] {
		sortBy = "t." + filter.SortBy
	}
	sortOrder := "DESC"
	if strings.ToUpper(filter.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	dataQuery := fmt.Sprintf(`
		SELECT t.id, t.list_id, t.created_by, t.title, t.description, t.status, t.priority, 
		       t.progress, t.due_date, t.estimated_minutes, t.started_at, t.completed_at, 
		       t.created_at, t.updated_at, t.deleted_at
		FROM tasks t
		JOIN task_lists tl ON tl.id = t.list_id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`, where, sortBy, sortOrder, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		task, err := scanTaskRow(rows)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, total, rows.Err()
}

func (r *postgresRepository) GetBoard(ctx context.Context, listID, userID uuid.UUID) (map[string][]Task, error) {
	query := `
		SELECT t.id, t.list_id, t.created_by, t.title, t.description, t.status, t.priority, 
		       t.progress, t.due_date, t.estimated_minutes, t.started_at, t.completed_at, 
		       t.created_at, t.updated_at, t.deleted_at
		FROM tasks t
		JOIN task_lists tl ON tl.id = t.list_id
		WHERE t.list_id = $1 AND tl.user_id = $2 AND t.deleted_at IS NULL
		ORDER BY t.priority DESC, t.created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, listID, userID)
	if err != nil {
		return nil, fmt.Errorf("get board: %w", err)
	}
	defer rows.Close()

	board := map[string][]Task{
		"backlog":     {},
		"todo":        {},
		"in_progress": {},
		"review":      {},
		"done":        {},
	}

	for rows.Next() {
		task, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		board[task.Status] = append(board[task.Status], *task)
	}

	return board, rows.Err()
}

// scanTaskRow scans a single task row.
func scanTaskRow(rows *sql.Rows) (*Task, error) {
	t := &Task{}
	var desc sql.NullString
	var dueDate sql.NullTime
	var estimatedMin sql.NullInt64
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	var deletedAt sql.NullTime

	err := rows.Scan(
		&t.ID, &t.ListID, &t.CreatedBy, &t.Title, &desc,
		&t.Status, &t.Priority, &t.Progress, &dueDate, &estimatedMin,
		&startedAt, &completedAt, &t.CreatedAt, &t.UpdatedAt, &deletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}

	if desc.Valid {
		t.Description = &desc.String
	}
	if dueDate.Valid {
		utc := dueDate.Time.UTC()
		t.DueDate = &utc
	}
	if estimatedMin.Valid {
		v := int(estimatedMin.Int64)
		t.EstimatedMin = &v
	}
	if startedAt.Valid {
		utc := startedAt.Time.UTC()
		t.StartedAt = &utc
	}
	if completedAt.Valid {
		utc := completedAt.Time.UTC()
		t.CompletedAt = &utc
	}
	t.CreatedAt = t.CreatedAt.UTC()
	t.UpdatedAt = t.UpdatedAt.UTC()

	return t, nil
}

func (r *postgresRepository) scanTask(ctx context.Context, query string, args ...any) (*Task, error) {
	t := &Task{}
	var desc sql.NullString
	var dueDate sql.NullTime
	var estimatedMin sql.NullInt64
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&t.ID, &t.ListID, &t.CreatedBy, &t.Title, &desc,
		&t.Status, &t.Priority, &t.Progress, &dueDate, &estimatedMin,
		&startedAt, &completedAt, &t.CreatedAt, &t.UpdatedAt, &deletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}

	if desc.Valid {
		t.Description = &desc.String
	}
	if dueDate.Valid {
		utc := dueDate.Time.UTC()
		t.DueDate = &utc
	}
	if estimatedMin.Valid {
		v := int(estimatedMin.Int64)
		t.EstimatedMin = &v
	}
	if startedAt.Valid {
		utc := startedAt.Time.UTC()
		t.StartedAt = &utc
	}
	if completedAt.Valid {
		utc := completedAt.Time.UTC()
		t.CompletedAt = &utc
	}
	t.CreatedAt = t.CreatedAt.UTC()
	t.UpdatedAt = t.UpdatedAt.UTC()

	return t, nil
}

func (r *postgresRepository) scanTaskSimple(ctx context.Context, query string, args ...any) (*Task, error) {
	return r.scanTask(ctx, query, args...)
}
