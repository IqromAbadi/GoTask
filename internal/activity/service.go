package activity

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Service handles activity business logic.
type Service struct {
	repo Repository
}

// NewService creates a new activity Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Log creates a new activity entry.
func (s *Service) Log(ctx context.Context, taskID *uuid.UUID, userID uuid.UUID, action string, oldValue, newValue *string, metadata map[string]any) error {
	var metaJSON json.RawMessage
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
		metaJSON = b
	}

	var tid *string
	if taskID != nil {
		s := taskID.String()
		tid = &s
	}

	a := &Activity{
		TaskID:   tid,
		UserID:   userID.String(),
		Action:   action,
		OldValue: oldValue,
		NewValue: newValue,
		Metadata: metaJSON,
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return fmt.Errorf("create activity: %w", err)
	}
	return nil
}

// ListByTask returns activities for a specific task.
func (s *Service) ListByTask(ctx context.Context, taskID uuid.UUID) ([]ActivityResponse, error) {
	activities, err := s.repo.ListByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list activities: %w", err)
	}

	resp := make([]ActivityResponse, len(activities))
	for i, a := range activities {
		resp[i] = ToResponse(&a)
	}
	return resp, nil
}

// ListByUser returns activities for a user with pagination.
func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]ActivityResponse, int, error) {
	activities, total, err := s.repo.ListByUser(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list user activities: %w", err)
	}

	resp := make([]ActivityResponse, len(activities))
	for i, a := range activities {
		resp[i] = ToResponse(&a)
	}
	return resp, total, nil
}
