package review

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL review repository.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, rev *Review) error {
	query := `
		INSERT INTO task_reviews (task_id, reviewer_id, status, submission_note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, rev.TaskID, rev.ReviewerID, rev.Status, rev.SubmissionNote).
		Scan(&rev.ID, &rev.CreatedAt, &rev.UpdatedAt)
	rev.CreatedAt = rev.CreatedAt.UTC()
	rev.UpdatedAt = rev.UpdatedAt.UTC()

	if err != nil {
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id, taskID uuid.UUID) (*Review, error) {
	query := `
		SELECT id, task_id, reviewer_id, status, submission_note, review_note, reviewed_at, created_at, updated_at
		FROM task_reviews
		WHERE id = $1 AND task_id = $2`

	rev := &Review{}
	var subNote, revNote sql.NullString
	var reviewedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id, taskID).Scan(
		&rev.ID, &rev.TaskID, &rev.ReviewerID, &rev.Status,
		&subNote, &revNote, &reviewedAt, &rev.CreatedAt, &rev.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}
	if subNote.Valid {
		rev.SubmissionNote = &subNote.String
	}
	if revNote.Valid {
		rev.ReviewNote = &revNote.String
	}
	if reviewedAt.Valid {
		t := reviewedAt.Time.UTC()
		rev.ReviewedAt = &t
	}
	rev.CreatedAt = rev.CreatedAt.UTC()
	rev.UpdatedAt = rev.UpdatedAt.UTC()
	return rev, nil
}

func (r *postgresRepository) List(ctx context.Context, taskID uuid.UUID) ([]Review, error) {
	query := `
		SELECT id, task_id, reviewer_id, status, submission_note, review_note, reviewed_at, created_at, updated_at
		FROM task_reviews
		WHERE task_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var rev Review
		var subNote, revNote sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(&rev.ID, &rev.TaskID, &rev.ReviewerID, &rev.Status,
			&subNote, &revNote, &reviewedAt, &rev.CreatedAt, &rev.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		if subNote.Valid {
			rev.SubmissionNote = &subNote.String
		}
		if revNote.Valid {
			rev.ReviewNote = &revNote.String
		}
		if reviewedAt.Valid {
			t := reviewedAt.Time.UTC()
			rev.ReviewedAt = &t
		}
		rev.CreatedAt = rev.CreatedAt.UTC()
		rev.UpdatedAt = rev.UpdatedAt.UTC()
		reviews = append(reviews, rev)
	}
	return reviews, rows.Err()
}

func (r *postgresRepository) Approve(ctx context.Context, id, taskID uuid.UUID, note string) (*Review, error) {
	query := `
		UPDATE task_reviews
		SET status = 'approved', review_note = $3, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND task_id = $2 AND status = 'pending'
		RETURNING id, task_id, reviewer_id, status, submission_note, review_note, reviewed_at, created_at, updated_at`

	rev := &Review{}
	var subNote, revNote sql.NullString
	var reviewedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id, taskID, note).Scan(
		&rev.ID, &rev.TaskID, &rev.ReviewerID, &rev.Status,
		&subNote, &revNote, &reviewedAt, &rev.CreatedAt, &rev.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("approve review: %w", err)
	}
	if subNote.Valid {
		rev.SubmissionNote = &subNote.String
	}
	if revNote.Valid {
		rev.ReviewNote = &revNote.String
	}
	if reviewedAt.Valid {
		t := reviewedAt.Time.UTC()
		rev.ReviewedAt = &t
	}
	rev.CreatedAt = rev.CreatedAt.UTC()
	rev.UpdatedAt = rev.UpdatedAt.UTC()
	return rev, nil
}

func (r *postgresRepository) RequestChanges(ctx context.Context, id, taskID uuid.UUID, note string) (*Review, error) {
	query := `
		UPDATE task_reviews
		SET status = 'changes_requested', review_note = $3, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND task_id = $2 AND status = 'pending'
		RETURNING id, task_id, reviewer_id, status, submission_note, review_note, reviewed_at, created_at, updated_at`

	rev := &Review{}
	var subNote, revNote sql.NullString
	var reviewedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id, taskID, note).Scan(
		&rev.ID, &rev.TaskID, &rev.ReviewerID, &rev.Status,
		&subNote, &revNote, &reviewedAt, &rev.CreatedAt, &rev.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("request changes: %w", err)
	}
	if subNote.Valid {
		rev.SubmissionNote = &subNote.String
	}
	if revNote.Valid {
		rev.ReviewNote = &revNote.String
	}
	if reviewedAt.Valid {
		t := reviewedAt.Time.UTC()
		rev.ReviewedAt = &t
	}
	rev.CreatedAt = rev.CreatedAt.UTC()
	rev.UpdatedAt = rev.UpdatedAt.UTC()
	return rev, nil
}
