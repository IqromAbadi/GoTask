package activity

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the data access interface for activities.
type Repository interface {
	Create(ctx context.Context, a *Activity) error
	ListByTask(ctx context.Context, taskID uuid.UUID) ([]Activity, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]Activity, int, error)
}
