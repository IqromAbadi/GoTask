package tasklist

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the data access interface for task lists.
type Repository interface {
	Create(ctx context.Context, tl *TaskList) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*TaskList, error)
	List(ctx context.Context, userID uuid.UUID) ([]TaskList, error)
	Update(ctx context.Context, tl *TaskList) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Archive(ctx context.Context, id, userID uuid.UUID) (*TaskList, error)
	Restore(ctx context.Context, id, userID uuid.UUID) (*TaskList, error)
}
