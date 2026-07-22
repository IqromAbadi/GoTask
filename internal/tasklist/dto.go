package tasklist

import "time"

// CreateTaskListRequest is the DTO for creating a task list.
type CreateTaskListRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
}

// UpdateTaskListRequest is the DTO for updating a task list.
type UpdateTaskListRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
}

// TaskListResponse is the public DTO for task list data.
type TaskListResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsArchived  bool    `json:"is_archived"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ToResponse converts a TaskList entity to a TaskListResponse.
func ToResponse(tl *TaskList) TaskListResponse {
	return TaskListResponse{
		ID:          tl.ID,
		Name:        tl.Name,
		Description: tl.Description,
		IsArchived:  tl.IsArchived,
		CreatedAt:   tl.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   tl.UpdatedAt.Format(time.RFC3339),
	}
}
