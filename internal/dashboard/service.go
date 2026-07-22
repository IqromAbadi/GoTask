package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SummaryData holds dashboard summary statistics.
type SummaryData struct {
	TotalTasks      int `json:"total_tasks"`
	Backlog         int `json:"backlog"`
	Todo            int `json:"todo"`
	InProgress      int `json:"in_progress"`
	Review          int `json:"review"`
	Done            int `json:"done"`
	Overdue         int `json:"overdue"`
	AverageProgress int `json:"average_progress"`
}

// PriorityCount holds count per priority level.
type PriorityCount struct {
	Priority string `json:"priority"`
	Count    int    `json:"count"`
}

// ProgressPoint holds progress analytics data point.
type ProgressPoint struct {
	Period      string `json:"period"`
	AvgProgress int    `json:"avg_progress"`
	TasksCount  int    `json:"tasks_count"`
}

// Service handles dashboard business logic.
type Service struct {
	db *sql.DB
}

// NewService creates a new dashboard Service.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// GetSummary returns a summary of tasks for the user.
func (s *Service) GetSummary(ctx context.Context, userID uuid.UUID, listID *uuid.UUID) (*SummaryData, error) {
	var query string
	var args []any

	if listID != nil {
		query = `
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
			WHERE tl.user_id = $1 AND t.list_id = $2`
		args = []any{userID, listID}
	} else {
		query = `
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
			WHERE tl.user_id = $1`
		args = []any{userID}
	}

	data := &SummaryData{}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&data.TotalTasks, &data.Backlog, &data.Todo, &data.InProgress,
		&data.Review, &data.Done, &data.Overdue, &data.AverageProgress,
	)
	if err != nil {
		return nil, fmt.Errorf("get summary: %w", err)
	}
	return data, nil
}

// GetPriorityDistribution returns task counts by priority.
func (s *Service) GetPriorityDistribution(ctx context.Context, userID uuid.UUID, listID *uuid.UUID) ([]PriorityCount, error) {
	var query string
	var args []any

	if listID != nil {
		query = `
			SELECT t.priority, COUNT(*) AS count
			FROM tasks t
			JOIN task_lists tl ON tl.id = t.list_id
			WHERE tl.user_id = $1 AND t.list_id = $2 AND t.deleted_at IS NULL
			GROUP BY t.priority
			ORDER BY t.priority`
		args = []any{userID, listID}
	} else {
		query = `
			SELECT t.priority, COUNT(*) AS count
			FROM tasks t
			JOIN task_lists tl ON tl.id = t.list_id
			WHERE tl.user_id = $1 AND t.deleted_at IS NULL
			GROUP BY t.priority
			ORDER BY t.priority`
		args = []any{userID}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("priority distribution: %w", err)
	}
	defer rows.Close()

	var result []PriorityCount
	for rows.Next() {
		var pc PriorityCount
		if err := rows.Scan(&pc.Priority, &pc.Count); err != nil {
			return nil, fmt.Errorf("scan priority: %w", err)
		}
		result = append(result, pc)
	}
	return result, rows.Err()
}

// GetUpcomingDeadlines returns tasks with upcoming due dates.
func (s *Service) GetUpcomingDeadlines(ctx context.Context, userID uuid.UUID, limit int) ([]map[string]any, error) {
	if limit < 1 {
		limit = 10
	}

	query := `
		SELECT t.id, t.title, t.due_date, t.status, t.priority, t.progress, tl.name AS list_name
		FROM tasks t
		JOIN task_lists tl ON tl.id = t.list_id
		WHERE tl.user_id = $1
		  AND t.due_date IS NOT NULL
		  AND t.due_date >= CURRENT_DATE
		  AND t.status != 'done'
		  AND t.deleted_at IS NULL
		ORDER BY t.due_date ASC
		LIMIT $2`

	rows, err := s.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("upcoming deadlines: %w", err)
	}
	defer rows.Close()

	var tasks []map[string]any
	for rows.Next() {
		var id, title, status, priority, listName string
		var progress int
		var dueDate time.Time
		if err := rows.Scan(&id, &title, &dueDate, &status, &priority, &progress, &listName); err != nil {
			return nil, fmt.Errorf("scan deadline: %w", err)
		}
		tasks = append(tasks, map[string]any{
			"id":        id,
			"title":     title,
			"due_date":  dueDate.Format("2006-01-02"),
			"status":    status,
			"priority":  priority,
			"progress":  progress,
			"list_name": listName,
		})
	}
	return tasks, rows.Err()
}

// GetOverdueTasks returns overdue tasks.
func (s *Service) GetOverdueTasks(ctx context.Context, userID uuid.UUID) ([]map[string]any, error) {
	query := `
		SELECT t.id, t.title, t.due_date, t.status, t.priority, t.progress, tl.name AS list_name
		FROM tasks t
		JOIN task_lists tl ON tl.id = t.list_id
		WHERE tl.user_id = $1
		  AND t.due_date < CURRENT_DATE
		  AND t.status != 'done'
		  AND t.deleted_at IS NULL
		ORDER BY t.due_date ASC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("overdue tasks: %w", err)
	}
	defer rows.Close()

	var tasks []map[string]any
	for rows.Next() {
		var id, title, status, priority, listName string
		var progress int
		var dueDate time.Time
		if err := rows.Scan(&id, &title, &dueDate, &status, &priority, &progress, &listName); err != nil {
			return nil, fmt.Errorf("scan overdue: %w", err)
		}
		tasks = append(tasks, map[string]any{
			"id":        id,
			"title":     title,
			"due_date":  dueDate.Format("2006-01-02"),
			"status":    status,
			"priority":  priority,
			"progress":  progress,
			"list_name": listName,
		})
	}
	return tasks, rows.Err()
}

// GetProgressAnalytics returns progress data for charts.
func (s *Service) GetProgressAnalytics(ctx context.Context, userID uuid.UUID, period string) ([]ProgressPoint, error) {
	validPeriods := map[string]string{
		"day":   "day",
		"week":  "week",
		"month": "month",
		"year":  "year",
	}
	p, ok := validPeriods[period]
	if !ok {
		p = "week"
	}

	query := `
		SELECT
			DATE_TRUNC($2, t.created_at) AS period,
			AVG(t.progress)::INTEGER AS avg_progress,
			COUNT(*) AS tasks_count
		FROM tasks t
		JOIN task_lists tl ON tl.id = t.list_id
		WHERE tl.user_id = $1 AND t.deleted_at IS NULL
		GROUP BY period
		ORDER BY period DESC
		LIMIT 30`

	rows, err := s.db.QueryContext(ctx, query, userID, p)
	if err != nil {
		return nil, fmt.Errorf("progress analytics: %w", err)
	}
	defer rows.Close()

	var points []ProgressPoint
	for rows.Next() {
		var pp ProgressPoint
		var periodTime time.Time
		if err := rows.Scan(&periodTime, &pp.AvgProgress, &pp.TasksCount); err != nil {
			return nil, fmt.Errorf("scan point: %w", err)
		}
		pp.Period = periodTime.Format("2006-01-02")
		points = append(points, pp)
	}
	return points, rows.Err()
}
