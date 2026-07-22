package tasklist

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL task list repository.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, tl *TaskList) error {
	query := `
		INSERT INTO task_lists (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, tl.UserID, tl.Name, tl.Description).
		Scan(&tl.ID, &tl.CreatedAt, &tl.UpdatedAt)

	tl.CreatedAt = tl.CreatedAt.UTC()
	tl.UpdatedAt = tl.UpdatedAt.UTC()

	if err != nil {
		return fmt.Errorf("create task list: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*TaskList, error) {
	query := `
		SELECT id, user_id, name, description, is_archived, created_at, updated_at
		FROM task_lists
		WHERE id = $1 AND user_id = $2`

	tl := &TaskList{}
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&tl.ID, &tl.UserID, &tl.Name, &desc, &tl.IsArchived, &tl.CreatedAt, &tl.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get task list by id: %w", err)
	}

	if desc.Valid {
		tl.Description = &desc.String
	}
	tl.CreatedAt = tl.CreatedAt.UTC()
	tl.UpdatedAt = tl.UpdatedAt.UTC()

	return tl, nil
}

func (r *postgresRepository) List(ctx context.Context, userID uuid.UUID) ([]TaskList, error) {
	query := `
		SELECT id, user_id, name, description, is_archived, created_at, updated_at
		FROM task_lists
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list task lists: %w", err)
	}
	defer rows.Close()

	var lists []TaskList
	for rows.Next() {
		var tl TaskList
		var desc sql.NullString
		if err := rows.Scan(&tl.ID, &tl.UserID, &tl.Name, &desc, &tl.IsArchived, &tl.CreatedAt, &tl.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task list: %w", err)
		}
		if desc.Valid {
			tl.Description = &desc.String
		}
		tl.CreatedAt = tl.CreatedAt.UTC()
		tl.UpdatedAt = tl.UpdatedAt.UTC()
		lists = append(lists, tl)
	}

	return lists, rows.Err()
}

func (r *postgresRepository) Update(ctx context.Context, tl *TaskList) error {
	query := `
		UPDATE task_lists
		SET name = $3, description = $4, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query, tl.ID, tl.UserID, tl.Name, tl.Description).
		Scan(&tl.UpdatedAt)
	tl.UpdatedAt = tl.UpdatedAt.UTC()

	if err != nil {
		return fmt.Errorf("update task list: %w", err)
	}
	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM task_lists WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete task list: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task list not found")
	}
	return nil
}

func (r *postgresRepository) Archive(ctx context.Context, id, userID uuid.UUID) (*TaskList, error) {
	query := `
		UPDATE task_lists
		SET is_archived = TRUE, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, description, is_archived, created_at, updated_at`

	tl := &TaskList{}
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&tl.ID, &tl.UserID, &tl.Name, &desc, &tl.IsArchived, &tl.CreatedAt, &tl.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("archive task list: %w", err)
	}

	if desc.Valid {
		tl.Description = &desc.String
	}
	tl.CreatedAt = tl.CreatedAt.UTC()
	tl.UpdatedAt = tl.UpdatedAt.UTC()

	return tl, nil
}

func (r *postgresRepository) Restore(ctx context.Context, id, userID uuid.UUID) (*TaskList, error) {
	query := `
		UPDATE task_lists
		SET is_archived = FALSE, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, description, is_archived, created_at, updated_at`

	tl := &TaskList{}
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&tl.ID, &tl.UserID, &tl.Name, &desc, &tl.IsArchived, &tl.CreatedAt, &tl.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("restore task list: %w", err)
	}

	if desc.Valid {
		tl.Description = &desc.String
	}
	tl.CreatedAt = tl.CreatedAt.UTC()
	tl.UpdatedAt = tl.UpdatedAt.UTC()

	return tl, nil
}
