package progress

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/activity"
)

// Service handles progress update business logic.
type Service struct {
	repo            Repository
	db              *sql.DB
	activityService *activity.Service
}

// NewService creates a new progress Service.
func NewService(repo Repository, db *sql.DB, activityService *activity.Service) *Service {
	return &Service{
		repo:            repo,
		db:              db,
		activityService: activityService,
	}
}

// Create adds a progress update with transaction support.
func (s *Service) Create(ctx context.Context, taskID, userID uuid.UUID, req CreateProgressRequest) (*ProgressResponse, error) {
	if req.Progress < 0 || req.Progress > 100 {
		return nil, fmt.Errorf("progress harus antara 0 sampai 100")
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check task status (must be in_progress)
	var taskStatus string
	var currentProgress int
	err = tx.QueryRowContext(ctx,
		`SELECT status, progress FROM tasks WHERE id = $1 AND deleted_at IS NULL`, taskID,
	).Scan(&taskStatus, &currentProgress)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task tidak ditemukan")
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	if taskStatus != "in_progress" {
		return nil, fmt.Errorf("progress hanya dapat ditambahkan saat task berstatus in_progress")
	}

	// Check rollback rules
	if req.AllowRollback {
		if req.Note == nil || *req.Note == "" {
			return nil, fmt.Errorf("note wajib diisi saat rollback")
		}
	} else {
		// New progress must not be less than current
		if req.Progress < currentProgress {
			return nil, fmt.Errorf("progress baru (%d) tidak boleh lebih kecil dari progress saat ini (%d). Gunakan allow_rollback untuk rollback", req.Progress, currentProgress)
		}
	}

	// Update task progress
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET progress = $2, updated_at = NOW() WHERE id = $1`,
		taskID, req.Progress,
	)
	if err != nil {
		return nil, fmt.Errorf("update task progress: %w", err)
	}

	// Create progress update
	progressID := uuid.New().String()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO task_progress_updates (id, task_id, user_id, progress, note) VALUES ($1, $2, $3, $4, $5)`,
		progressID, taskID, userID, req.Progress, req.Note,
	)
	if err != nil {
		return nil, fmt.Errorf("create progress update: %w", err)
	}

	// Create activity log
	oldProgress := fmt.Sprintf("%d%%", currentProgress)
	newProgress := fmt.Sprintf("%d%%", req.Progress)
	action := "task_progress_updated"
	if req.AllowRollback {
		action = "task_progress_rolled_back"
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO task_activities (task_id, user_id, action, old_value, new_value) VALUES ($1, $2, $3, $4, $5)`,
		taskID, userID, action, oldProgress, newProgress,
	)
	if err != nil {
		return nil, fmt.Errorf("create activity: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	resp := &ProgressResponse{
		ID:       progressID,
		TaskID:   taskID.String(),
		UserID:   userID.String(),
		Progress: req.Progress,
		Note:     req.Note,
	}

	return resp, nil
}

// ListByTask returns all progress updates for a task.
func (s *Service) ListByTask(ctx context.Context, taskID uuid.UUID) ([]ProgressResponse, error) {
	updates, err := s.repo.ListByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list progress: %w", err)
	}

	resp := make([]ProgressResponse, len(updates))
	for i, u := range updates {
		resp[i] = ToResponse(&u)
	}
	return resp, nil
}

// GetByID returns a specific progress update.
func (s *Service) GetByID(ctx context.Context, id, taskID uuid.UUID) (*ProgressResponse, error) {
	update, err := s.repo.GetByID(ctx, id, taskID)
	if err != nil {
		return nil, fmt.Errorf("get progress: %w", err)
	}
	if update == nil {
		return nil, fmt.Errorf("progress tidak ditemukan")
	}

	resp := ToResponse(update)
	return &resp, nil
}

// UpdateNote updates the note on a progress update.
func (s *Service) UpdateNote(ctx context.Context, id, taskID uuid.UUID, note string) (*ProgressResponse, error) {
	update, err := s.repo.UpdateNote(ctx, id, taskID, note)
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	if update == nil {
		return nil, fmt.Errorf("progress tidak ditemukan")
	}

	resp := ToResponse(update)
	return &resp, nil
}

// Delete deletes a progress update.
func (s *Service) Delete(ctx context.Context, id, taskID uuid.UUID) error {
	if err := s.repo.Delete(ctx, id, taskID); err != nil {
		return fmt.Errorf("delete progress: %w", err)
	}
	return nil
}
