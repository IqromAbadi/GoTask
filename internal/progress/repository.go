package progress

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the data access interface for progress updates.
type Repository interface {
	Create(ctx context.Context, p *ProgressUpdate) error
	ListByTask(ctx context.Context, taskID uuid.UUID) ([]ProgressUpdate, error)
	GetByID(ctx context.Context, id, taskID uuid.UUID) (*ProgressUpdate, error)
	UpdateNote(ctx context.Context, id, taskID uuid.UUID, note string) (*ProgressUpdate, error)
	Delete(ctx context.Context, id, taskID uuid.UUID) error
	GetLatest(ctx context.Context, taskID uuid.UUID) (*ProgressUpdate, error)
}
