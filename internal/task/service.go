package task

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/activity"
)

// Service handles task business logic.
type Service struct {
	repo            Repository
	activityService *activity.Service
	db              *sql.DB
}

// NewService creates a new task Service.
func NewService(repo Repository, activityService *activity.Service, db *sql.DB) *Service {
	return &Service{
		repo:            repo,
		activityService: activityService,
		db:              db,
	}
}

// Create creates a new task with defaults.
func (s *Service) Create(ctx context.Context, listID, userID uuid.UUID, req CreateTaskRequest) (*TaskResponse, error) {
	status := req.Status
	if status == "" {
		status = "backlog"
	}
	if !IsValidStatus(status) {
		return nil, fmt.Errorf("status tidak valid")
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	if !IsValidPriority(priority) {
		return nil, fmt.Errorf("prioritas tidak valid")
	}

	t := &Task{
		ListID:       listID.String(),
		CreatedBy:    userID.String(),
		Title:        req.Title,
		Description:  req.Description,
		Status:       status,
		Priority:     priority,
		DueDate:      parseDate(req.DueDate),
		EstimatedMin: req.EstimatedMin,
		Progress:     0,
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	// Log activity
	if s.activityService != nil {
		taskUID, _ := uuid.Parse(t.ID)
		_ = s.activityService.Log(ctx, &taskUID, userID, "task_created", nil, nil,
			map[string]any{"task_title": t.Title})
	}

	resp := ToResponse(t)
	return &resp, nil
}

// GetByID retrieves a task by ID (ownership checked via JOIN in repo).
func (s *Service) GetByID(ctx context.Context, id, userID uuid.UUID) (*TaskResponse, error) {
	task, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task tidak ditemukan")
	}

	resp := ToResponse(task)
	return &resp, nil
}

// Update updates a task's fields.
func (s *Service) Update(ctx context.Context, id, listID uuid.UUID, req UpdateTaskRequest) (*TaskResponse, error) {
	if req.Priority != "" && !IsValidPriority(req.Priority) {
		return nil, fmt.Errorf("prioritas tidak valid")
	}

	t := &Task{
		ID:           id.String(),
		ListID:       listID.String(),
		Title:        req.Title,
		Description:  req.Description,
		Priority:     req.Priority,
		DueDate:      parseDate(req.DueDate),
		EstimatedMin: req.EstimatedMin,
	}

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	resp := ToResponse(t)
	return &resp, nil
}

// Delete soft-deletes a task.
func (s *Service) Delete(ctx context.Context, id, listID uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, id, listID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

// List returns filtered, sorted, paginated tasks.
func (s *Service) List(ctx context.Context, listID, userID uuid.UUID, filter TaskFilter) ([]TaskResponse, int, error) {
	// Defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}

	tasks, total, err := s.repo.List(ctx, listID, userID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}

	resp := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i] = ToResponse(&t)
	}

	return resp, total, nil
}

// GetBoard returns tasks grouped by status.
func (s *Service) GetBoard(ctx context.Context, listID, userID uuid.UUID) (map[string][]TaskResponse, error) {
	board, err := s.repo.GetBoard(ctx, listID, userID)
	if err != nil {
		return nil, fmt.Errorf("get board: %w", err)
	}

	result := map[string][]TaskResponse{
		"backlog":     {},
		"todo":        {},
		"in_progress": {},
		"review":      {},
		"done":        {},
	}

	for status, tasks := range board {
		resp := make([]TaskResponse, len(tasks))
		for i, t := range tasks {
			resp[i] = ToResponse(&t)
		}
		result[status] = resp
	}

	return result, nil
}

// UpdateStatus changes a task's status with workflow validation.
func (s *Service) UpdateStatus(ctx context.Context, id, listID uuid.UUID, newStatus string, userID uuid.UUID) (*TaskResponse, error) {
	if !IsValidStatus(newStatus) {
		return nil, fmt.Errorf("status tidak valid")
	}

	// Get current task
	current, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("task tidak ditemukan")
	}

	if !isValidTransition(current.Status, newStatus) {
		return nil, fmt.Errorf("transisi status dari %s ke %s tidak diizinkan", current.Status, newStatus)
	}

	// Business rules for done status
	if newStatus == "done" {
		return nil, fmt.Errorf("status done hanya dapat dicapai melalui review")
	}

	// Business rules for review status
	if newStatus == "review" && current.Progress < 100 {
		return nil, fmt.Errorf("progress harus 100%% sebelum masuk review")
	}

	task, err := s.repo.UpdateStatus(ctx, id, listID, newStatus)
	if err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	// Set started_at if moving to in_progress for the first time
	if newStatus == "in_progress" && current.StartedAt == nil {
		now := time.Now().UTC()
		_, err = s.db.ExecContext(ctx,
			`UPDATE tasks SET started_at = $1 WHERE id = $2 AND started_at IS NULL`,
			now, id,
		)
		if err != nil {
			return nil, fmt.Errorf("set started_at: %w", err)
		}
		task.StartedAt = &now
	}

	// Log activity
	if s.activityService != nil {
		_ = s.activityService.Log(ctx, &id, userID, "task_status_changed",
			&current.Status, &newStatus,
			map[string]any{"task_title": current.Title})
	}

	resp := ToResponse(task)
	return &resp, nil
}

// UpdatePriority changes a task's priority.
func (s *Service) UpdatePriority(ctx context.Context, id, listID uuid.UUID, priority string, userID uuid.UUID) (*TaskResponse, error) {
	if !IsValidPriority(priority) {
		return nil, fmt.Errorf("prioritas tidak valid")
	}

	// Verify ownership
	current, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("task tidak ditemukan")
	}

	t := &Task{
		ID:       id.String(),
		ListID:   listID.String(),
		Title:    current.Title,
		Priority: priority,
	}

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("update priority: %w", err)
	}

	// Log activity
	if s.activityService != nil {
		_ = s.activityService.Log(ctx, &id, userID, "task_priority_changed",
			&current.Priority, &priority,
			map[string]any{"task_title": current.Title})
	}

	// Re-fetch to get full updated state
	updated, _ := s.repo.GetByID(ctx, id, userID)
	if updated != nil {
		resp := ToResponse(updated)
		return &resp, nil
	}
	resp := ToResponse(t)
	return &resp, nil
}

// Reopen reopens a completed task.
func (s *Service) Reopen(ctx context.Context, id, listID uuid.UUID, userID uuid.UUID) (*TaskResponse, error) {
	current, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("task tidak ditemukan")
	}
	if current.Status != "done" {
		return nil, fmt.Errorf("hanya task dengan status done yang dapat dibuka kembali")
	}

	task, err := s.repo.Reopen(ctx, id, listID)
	if err != nil {
		return nil, fmt.Errorf("reopen task: %w", err)
	}

	// Log activity
	if s.activityService != nil {
		_ = s.activityService.Log(ctx, &id, userID, "task_reopened", nil, nil,
			map[string]any{"task_title": current.Title})
	}

	resp := ToResponse(task)
	return &resp, nil
}

// Valid transitions map
var validTransitions = map[string][]string{
	"backlog":     {"todo"},
	"todo":        {"backlog", "in_progress"},
	"in_progress": {"todo", "review"},
	"review":      {"in_progress", "done"},
	"done":        {"in_progress"},
}

func isValidTransition(from, to string) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}

// parseDate parses a date string in YYYY-MM-DD format to time.Time.
func parseDate(dateStr *string) *time.Time {
	if dateStr == nil || *dateStr == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return nil
	}
	return &t
}
