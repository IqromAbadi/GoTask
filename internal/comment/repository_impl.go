package comment

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/iqromabadi/gotask/internal/activity"
)

// Repository defines the data access interface for comments.
type Repository interface {
	Create(ctx context.Context, c *Comment) error
	GetByID(ctx context.Context, id, taskID uuid.UUID) (*Comment, error)
	List(ctx context.Context, taskID uuid.UUID) ([]Comment, error)
	Update(ctx context.Context, c *Comment) error
	SoftDelete(ctx context.Context, id, taskID, userID uuid.UUID) error
}

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL comment repository.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, c *Comment) error {
	query := `
		INSERT INTO task_comments (task_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, c.TaskID, c.UserID, c.Content).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id, taskID uuid.UUID) (*Comment, error) {
	query := `
		SELECT id, task_id, user_id, content, created_at, updated_at, deleted_at
		FROM task_comments
		WHERE id = $1 AND task_id = $2 AND deleted_at IS NULL`

	c := &Comment{}
	var deletedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id, taskID).Scan(
		&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.CreatedAt, &c.UpdatedAt, &deletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get comment: %w", err)
	}
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

func (r *postgresRepository) List(ctx context.Context, taskID uuid.UUID) ([]Comment, error) {
	query := `
		SELECT id, task_id, user_id, content, created_at, updated_at, deleted_at
		FROM task_comments
		WHERE task_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		var deletedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.CreatedAt, &c.UpdatedAt, &deletedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		c.CreatedAt = c.CreatedAt.UTC()
		c.UpdatedAt = c.UpdatedAt.UTC()
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (r *postgresRepository) Update(ctx context.Context, c *Comment) error {
	query := `
		UPDATE task_comments
		SET content = $3, updated_at = NOW()
		WHERE id = $1 AND task_id = $2 AND deleted_at IS NULL
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query, c.ID, c.TaskID, c.Content).Scan(&c.UpdatedAt)
	c.UpdatedAt = c.UpdatedAt.UTC()
	if err != nil {
		return fmt.Errorf("update comment: %w", err)
	}
	return nil
}

func (r *postgresRepository) SoftDelete(ctx context.Context, id, taskID, userID uuid.UUID) error {
	query := `
		UPDATE task_comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND task_id = $2 AND user_id = $3 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, taskID, userID)
	if err != nil {
		return fmt.Errorf("soft delete comment: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("comment not found or not owned by user")
	}
	return nil
}

// Service handles comment business logic.
type Service struct {
	repo            Repository
	activityService *activity.Service
}

// NewService creates a new comment Service.
func NewService(repo Repository, activityService *activity.Service) *Service {
	return &Service{repo: repo, activityService: activityService}
}

// Create adds a comment to a task.
func (s *Service) Create(ctx context.Context, taskID, userID uuid.UUID, req CreateCommentRequest) (*CommentResponse, error) {
	c := &Comment{
		TaskID:  taskID.String(),
		UserID:  userID.String(),
		Content: req.Content,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	_ = s.activityService.Log(ctx, &taskID, userID, "comment_created", nil, nil, nil)

	resp := ToResponse(c)
	return &resp, nil
}

// List returns all non-deleted comments for a task.
func (s *Service) List(ctx context.Context, taskID uuid.UUID) ([]CommentResponse, error) {
	comments, err := s.repo.List(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	resp := make([]CommentResponse, len(comments))
	for i, c := range comments {
		resp[i] = ToResponse(&c)
	}
	return resp, nil
}

// Update updates a comment's content (only owner).
func (s *Service) Update(ctx context.Context, commentID, taskID, userID uuid.UUID, req UpdateCommentRequest) (*CommentResponse, error) {
	// Verify ownership
	existing, err := s.repo.GetByID(ctx, commentID, taskID)
	if err != nil {
		return nil, fmt.Errorf("get comment: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("komentar tidak ditemukan")
	}
	if existing.UserID != userID.String() {
		return nil, fmt.Errorf("anda hanya dapat mengedit komentar sendiri")
	}

	c := &Comment{
		ID:      commentID.String(),
		TaskID:  taskID.String(),
		Content: req.Content,
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("update comment: %w", err)
	}

	_ = s.activityService.Log(ctx, &taskID, userID, "comment_updated", nil, nil, nil)

	resp := ToResponse(c)
	return &resp, nil
}

// Delete soft-deletes a comment (only owner).
func (s *Service) Delete(ctx context.Context, commentID, taskID, userID uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, commentID, taskID, userID); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	_ = s.activityService.Log(ctx, &taskID, userID, "comment_deleted", nil, nil, nil)
	return nil
}

// Ensure time is used
var _ = time.Now
