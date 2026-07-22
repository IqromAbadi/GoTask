package task

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the data access interface for tasks.
type Repository interface {
	Create(ctx context.Context, t *Task) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Task, error)
	Update(ctx context.Context, t *Task) error
	UpdateStatus(ctx context.Context, id, listID uuid.UUID, status string) (*Task, error)
	UpdateProgress(ctx context.Context, id uuid.UUID, progress int) (*Task, error)
	SoftDelete(ctx context.Context, id, listID uuid.UUID) error
	MarkDone(ctx context.Context, id uuid.UUID) (*Task, error)
	Reopen(ctx context.Context, id, listID uuid.UUID) (*Task, error)
	List(ctx context.Context, listID, userID uuid.UUID, filter TaskFilter) ([]Task, int, error)
	GetBoard(ctx context.Context, listID, userID uuid.UUID) (map[string][]Task, error)
}
