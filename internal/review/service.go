package review

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/activity"
)

// Service handles review business logic.
type Service struct {
	repo            Repository
	db              *sql.DB
	activityService *activity.Service
}

// NewService creates a new review Service.
func NewService(repo Repository, db *sql.DB, activityService *activity.Service) *Service {
	return &Service{repo: repo, db: db, activityService: activityService}
}

// SubmitReview creates a new review and transitions task to review status.
func (s *Service) SubmitReview(ctx context.Context, taskID, userID uuid.UUID, note string) (*ReviewResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check task status and progress
	var taskStatus string
	var taskProgress int
	var taskTitle string
	err = tx.QueryRowContext(ctx,
		`SELECT status, progress, title FROM tasks WHERE id = $1 AND deleted_at IS NULL`, taskID,
	).Scan(&taskStatus, &taskProgress, &taskTitle)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task tidak ditemukan")
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	if taskStatus != "in_progress" {
		return nil, fmt.Errorf("task harus berstatus in_progress untuk submit review")
	}
	if taskProgress < 100 {
		return nil, fmt.Errorf("progress harus 100%% sebelum submit review")
	}

	// Create review record
	reviewID := uuid.New().String()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO task_reviews (id, task_id, reviewer_id, status, submission_note) VALUES ($1, $2, $3, 'pending', $4)`,
		reviewID, taskID, userID, note,
	)
	if err != nil {
		return nil, fmt.Errorf("create review: %w", err)
	}

	// Update task status to review
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET status = 'review', updated_at = NOW() WHERE id = $1`, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("update task status: %w", err)
	}

	// Log activity
	_, err = tx.ExecContext(ctx,
		`INSERT INTO task_activities (task_id, user_id, action, new_value, metadata) VALUES ($1, $2, 'task_submitted_for_review', 'review', $3)`,
		taskID, userID, fmt.Sprintf(`{"task_title":"%s"}`, taskTitle),
	)
	if err != nil {
		return nil, fmt.Errorf("create activity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &ReviewResponse{
		ID:         reviewID,
		TaskID:     taskID.String(),
		ReviewerID: userID.String(),
		Status:     "pending",
	}, nil
}

// Approve approves a review and moves task to done.
func (s *Service) Approve(ctx context.Context, reviewID, taskID, userID uuid.UUID, note string) (*ReviewResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check review exists and is pending
	var revStatus string
	var taskTitle string
	err = tx.QueryRowContext(ctx,
		`SELECT r.status, t.title FROM task_reviews r JOIN tasks t ON t.id = r.task_id WHERE r.id = $1 AND r.task_id = $2`,
		reviewID, taskID,
	).Scan(&revStatus, &taskTitle)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("review tidak ditemukan")
	}
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}
	if revStatus != "pending" {
		return nil, fmt.Errorf("review sudah diproses sebelumnya")
	}

	// Approve review
	_, err = tx.ExecContext(ctx,
		`UPDATE task_reviews SET status = 'approved', review_note = $3, reviewed_at = NOW(), updated_at = NOW() WHERE id = $1 AND task_id = $2`,
		reviewID, taskID, note,
	)
	if err != nil {
		return nil, fmt.Errorf("approve review: %w", err)
	}

	// Move task to done
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET status = 'done', progress = 100, completed_at = NOW(), updated_at = NOW() WHERE id = $1`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("update task to done: %w", err)
	}

	// Log activity
	_, err = tx.ExecContext(ctx,
		`INSERT INTO task_activities (task_id, user_id, action, metadata) VALUES ($1, $2, 'review_approved', $3)`,
		taskID, userID, fmt.Sprintf(`{"task_title":"%s"}`, taskTitle),
	)
	if err != nil {
		return nil, fmt.Errorf("create activity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &ReviewResponse{ID: reviewID.String(), TaskID: taskID.String(), Status: "approved"}, nil
}

// RequestChanges requests changes on a review.
func (s *Service) RequestChanges(ctx context.Context, reviewID, taskID, userID uuid.UUID, note string) (*ReviewResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check review exists and is pending
	var revStatus string
	var taskTitle string
	err = tx.QueryRowContext(ctx,
		`SELECT r.status, t.title FROM task_reviews r JOIN tasks t ON t.id = r.task_id WHERE r.id = $1 AND r.task_id = $2`,
		reviewID, taskID,
	).Scan(&revStatus, &taskTitle)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("review tidak ditemukan")
	}
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}
	if revStatus != "pending" {
		return nil, fmt.Errorf("review sudah diproses sebelumnya")
	}

	// Update review
	_, err = tx.ExecContext(ctx,
		`UPDATE task_reviews SET status = 'changes_requested', review_note = $3, reviewed_at = NOW(), updated_at = NOW() WHERE id = $1 AND task_id = $2`,
		reviewID, taskID, note,
	)
	if err != nil {
		return nil, fmt.Errorf("request changes: %w", err)
	}

	// Move task back to in_progress and clear completed_at
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET status = 'in_progress', completed_at = NULL, updated_at = NOW() WHERE id = $1`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	// Log activity
	_, err = tx.ExecContext(ctx,
		`INSERT INTO task_activities (task_id, user_id, action, metadata) VALUES ($1, $2, 'review_changes_requested', $3)`,
		taskID, userID, fmt.Sprintf(`{"task_title":"%s"}`, taskTitle),
	)
	if err != nil {
		return nil, fmt.Errorf("create activity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &ReviewResponse{ID: reviewID.String(), TaskID: taskID.String(), Status: "changes_requested"}, nil
}

// List returns all reviews for a task.
func (s *Service) List(ctx context.Context, taskID uuid.UUID) ([]ReviewResponse, error) {
	reviews, err := s.repo.List(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	resp := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		resp[i] = ToResponse(&r)
	}
	return resp, nil
}

// GetByID returns a specific review.
func (s *Service) GetByID(ctx context.Context, id, taskID uuid.UUID) (*ReviewResponse, error) {
	review, err := s.repo.GetByID(ctx, id, taskID)
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}
	if review == nil {
		return nil, fmt.Errorf("review tidak ditemukan")
	}
	resp := ToResponse(review)
	return &resp, nil
}
