package progress

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL progress repository.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, p *ProgressUpdate) error {
	query := `
		INSERT INTO task_progress_updates (task_id, user_id, progress, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, p.TaskID, p.UserID, p.Progress, p.Note).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()

	if err != nil {
		return fmt.Errorf("create progress update: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListByTask(ctx context.Context, taskID uuid.UUID) ([]ProgressUpdate, error) {
	query := `
		SELECT id, task_id, user_id, progress, note, created_at, updated_at
		FROM task_progress_updates
		WHERE task_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("list progress updates: %w", err)
	}
	defer rows.Close()

	var updates []ProgressUpdate
	for rows.Next() {
		var p ProgressUpdate
		var note sql.NullString
		if err := rows.Scan(&p.ID, &p.TaskID, &p.UserID, &p.Progress, &note, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan progress: %w", err)
		}
		if note.Valid {
			p.Note = &note.String
		}
		p.CreatedAt = p.CreatedAt.UTC()
		p.UpdatedAt = p.UpdatedAt.UTC()
		updates = append(updates, p)
	}

	return updates, rows.Err()
}

func (r *postgresRepository) GetByID(ctx context.Context, id, taskID uuid.UUID) (*ProgressUpdate, error) {
	query := `
		SELECT id, task_id, user_id, progress, note, created_at, updated_at
		FROM task_progress_updates
		WHERE id = $1 AND task_id = $2`

	p := &ProgressUpdate{}
	var note sql.NullString
	err := r.db.QueryRowContext(ctx, query, id, taskID).Scan(
		&p.ID, &p.TaskID, &p.UserID, &p.Progress, &note, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get progress update: %w", err)
	}

	if note.Valid {
		p.Note = &note.String
	}
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()

	return p, nil
}

func (r *postgresRepository) UpdateNote(ctx context.Context, id, taskID uuid.UUID, note string) (*ProgressUpdate, error) {
	query := `
		UPDATE task_progress_updates
		SET note = $3, updated_at = NOW()
		WHERE id = $1 AND task_id = $2
		RETURNING id, task_id, user_id, progress, note, created_at, updated_at`

	p := &ProgressUpdate{}
	var noteScan sql.NullString
	err := r.db.QueryRowContext(ctx, query, id, taskID, note).Scan(
		&p.ID, &p.TaskID, &p.UserID, &p.Progress, &noteScan, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update progress note: %w", err)
	}

	if noteScan.Valid {
		p.Note = &noteScan.String
	}
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()

	return p, nil
}

func (r *postgresRepository) Delete(ctx context.Context, id, taskID uuid.UUID) error {
	query := `DELETE FROM task_progress_updates WHERE id = $1 AND task_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, taskID)
	if err != nil {
		return fmt.Errorf("delete progress update: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("progress update not found")
	}
	return nil
}

func (r *postgresRepository) GetLatest(ctx context.Context, taskID uuid.UUID) (*ProgressUpdate, error) {
	query := `
		SELECT id, task_id, user_id, progress, note, created_at, updated_at
		FROM task_progress_updates
		WHERE task_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	p := &ProgressUpdate{}
	var note sql.NullString
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(
		&p.ID, &p.TaskID, &p.UserID, &p.Progress, &note, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest progress: %w", err)
	}

	if note.Valid {
		p.Note = &note.String
	}
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()

	return p, nil
}
