package review

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the data access interface for reviews.
type Repository interface {
	Create(ctx context.Context, r *Review) error
	GetByID(ctx context.Context, id, taskID uuid.UUID) (*Review, error)
	List(ctx context.Context, taskID uuid.UUID) ([]Review, error)
	Approve(ctx context.Context, id, taskID uuid.UUID, note string) (*Review, error)
	RequestChanges(ctx context.Context, id, taskID uuid.UUID, note string) (*Review, error)
}
