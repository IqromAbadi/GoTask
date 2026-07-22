package task

import "time"

// CreateTaskRequest is the DTO for creating a task.
type CreateTaskRequest struct {
	Title        string     `json:"title" validate:"required"`
	Description  *string    `json:"description"`
	Priority     string     `json:"priority"`
	Status       string     `json:"status"`
	DueDate      *time.Time `json:"due_date"`
	EstimatedMin *int       `json:"estimated_minutes"`
}

// UpdateTaskRequest is the DTO for updating a task.
type UpdateTaskRequest struct {
	Title        string     `json:"title" validate:"required"`
	Description  *string    `json:"description"`
	Priority     string     `json:"priority"`
	DueDate      *time.Time `json:"due_date"`
	EstimatedMin *int       `json:"estimated_minutes"`
}

// UpdateStatusRequest is the DTO for changing task status.
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// UpdatePriorityRequest is the DTO for changing task priority.
type UpdatePriorityRequest struct {
	Priority string `json:"priority" validate:"required"`
}

// TaskResponse is the public DTO for task data.
type TaskResponse struct {
	ID           string  `json:"id"`
	ListID       string  `json:"list_id"`
	CreatedBy    string  `json:"created_by"`
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	Status       string  `json:"status"`
	Priority     string  `json:"priority"`
	Progress     int     `json:"progress"`
	DueDate      *string `json:"due_date"`
	EstimatedMin *int    `json:"estimated_minutes"`
	StartedAt    *string `json:"started_at"`
	CompletedAt  *string `json:"completed_at"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// TaskFilter holds filter parameters for listing tasks.
type TaskFilter struct {
	Status      string
	Priority    string
	Search      string
	DueDateFrom *time.Time
	DueDateTo   *time.Time
	IsOverdue   bool
	SortBy      string
	SortOrder   string
	Page        int
	Limit       int
}

// ToResponse converts a Task entity to a TaskResponse.
func ToResponse(t *Task) TaskResponse {
	resp := TaskResponse{
		ID:           t.ID,
		ListID:       t.ListID,
		CreatedBy:    t.CreatedBy,
		Title:        t.Title,
		Description:  t.Description,
		Status:       t.Status,
		Priority:     t.Priority,
		Progress:     t.Progress,
		EstimatedMin: t.EstimatedMin,
		CreatedAt:    t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    t.UpdatedAt.Format(time.RFC3339),
	}

	if t.DueDate != nil {
		s := t.DueDate.Format("2006-01-02")
		resp.DueDate = &s
	}
	if t.StartedAt != nil {
		s := t.StartedAt.Format(time.RFC3339)
		resp.StartedAt = &s
	}
	if t.CompletedAt != nil {
		s := t.CompletedAt.Format(time.RFC3339)
		resp.CompletedAt = &s
	}

	return resp
}
