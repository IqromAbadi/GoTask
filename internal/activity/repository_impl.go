package activity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL activity repository.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, a *Activity) error {
	query := `
		INSERT INTO task_activities (task_id, user_id, action, old_value, new_value, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	var metadataJSON any
	if a.Metadata != nil {
		metadataJSON = a.Metadata
	}

	err := r.db.QueryRowContext(ctx, query,
		a.TaskID, a.UserID, a.Action, a.OldValue, a.NewValue, metadataJSON,
	).Scan(&a.ID, &a.CreatedAt)

	a.CreatedAt = a.CreatedAt.UTC()

	if err != nil {
		return fmt.Errorf("create activity: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListByTask(ctx context.Context, taskID uuid.UUID) ([]Activity, error) {
	query := `
		SELECT id, task_id, user_id, action, old_value, new_value, metadata, created_at
		FROM task_activities
		WHERE task_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("list task activities: %w", err)
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var a Activity
		var oldVal, newVal sql.NullString
		var metadataBytes []byte
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Action,
			&oldVal, &newVal, &metadataBytes, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}
		if oldVal.Valid {
			a.OldValue = &oldVal.String
		}
		if newVal.Valid {
			a.NewValue = &newVal.String
		}
		if len(metadataBytes) > 0 {
			a.Metadata = json.RawMessage(metadataBytes)
		}
		a.CreatedAt = a.CreatedAt.UTC()
		activities = append(activities, a)
	}

	return activities, rows.Err()
}

func (r *postgresRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]Activity, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM task_activities a
		JOIN tasks t ON t.id = a.task_id
		JOIN task_lists tl ON tl.id = t.list_id
		WHERE tl.user_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count activities: %w", err)
	}

	query := `
		SELECT a.id, a.task_id, a.user_id, a.action, a.old_value, a.new_value, a.metadata, a.created_at
		FROM task_activities a
		JOIN tasks t ON t.id = a.task_id
		JOIN task_lists tl ON tl.id = t.list_id
		WHERE tl.user_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list user activities: %w", err)
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var a Activity
		var oldVal, newVal sql.NullString
		var metadataBytes []byte
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Action,
			&oldVal, &newVal, &metadataBytes, &a.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan activity: %w", err)
		}
		if oldVal.Valid {
			a.OldValue = &oldVal.String
		}
		if newVal.Valid {
			a.NewValue = &newVal.String
		}
		if len(metadataBytes) > 0 {
			a.Metadata = json.RawMessage(metadataBytes)
		}
		a.CreatedAt = a.CreatedAt.UTC()
		activities = append(activities, a)
	}

	return activities, total, rows.Err()
}
