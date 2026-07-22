package tasklist

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles task list business logic.
type Service struct {
	repo Repository
}

// NewService creates a new task list Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new task list.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateTaskListRequest) (*TaskListResponse, error) {
	tl := &TaskList{
		UserID:      userID.String(),
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, tl); err != nil {
		return nil, fmt.Errorf("create task list: %w", err)
	}

	resp := ToResponse(tl)
	return &resp, nil
}

// GetByID retrieves a task list by ID (ownership checked in repo).
func (s *Service) GetByID(ctx context.Context, id, userID uuid.UUID) (*TaskListResponse, error) {
	tl, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get task list: %w", err)
	}
	if tl == nil {
		return nil, fmt.Errorf("task list tidak ditemukan")
	}

	resp := ToResponse(tl)
	return &resp, nil
}

// List retrieves all task lists for a user.
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]TaskListResponse, error) {
	lists, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list task lists: %w", err)
	}

	resp := make([]TaskListResponse, len(lists))
	for i, tl := range lists {
		resp[i] = ToResponse(&tl)
	}
	return resp, nil
}

// Update updates a task list.
func (s *Service) Update(ctx context.Context, id, userID uuid.UUID, req UpdateTaskListRequest) (*TaskListResponse, error) {
	tl := &TaskList{
		ID:          id.String(),
		UserID:      userID.String(),
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.Update(ctx, tl); err != nil {
		return nil, fmt.Errorf("update task list: %w", err)
	}

	resp := ToResponse(tl)
	return &resp, nil
}

// Delete deletes a task list.
func (s *Service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete task list: %w", err)
	}
	return nil
}

// Archive archives a task list.
func (s *Service) Archive(ctx context.Context, id, userID uuid.UUID) (*TaskListResponse, error) {
	tl, err := s.repo.Archive(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("archive task list: %w", err)
	}
	if tl == nil {
		return nil, fmt.Errorf("task list tidak ditemukan")
	}

	resp := ToResponse(tl)
	return &resp, nil
}

// Restore restores an archived task list.
func (s *Service) Restore(ctx context.Context, id, userID uuid.UUID) (*TaskListResponse, error) {
	tl, err := s.repo.Restore(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("restore task list: %w", err)
	}
	if tl == nil {
		return nil, fmt.Errorf("task list tidak ditemukan")
	}

	resp := ToResponse(tl)
	return &resp, nil
}
