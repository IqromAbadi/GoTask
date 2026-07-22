package progress

import "time"

// CreateProgressRequest is the DTO for adding a progress update.
type CreateProgressRequest struct {
	Progress      int     `json:"progress" validate:"required,gte=0,lte=100"`
	Note          *string `json:"note"`
	AllowRollback bool    `json:"allow_rollback"`
}

// UpdateProgressNoteRequest is the DTO for updating a progress note.
type UpdateProgressNoteRequest struct {
	Note string `json:"note" validate:"required"`
}

// ProgressResponse is the public DTO for progress data.
type ProgressResponse struct {
	ID        string  `json:"id"`
	TaskID    string  `json:"task_id"`
	UserID    string  `json:"user_id"`
	Progress  int     `json:"progress"`
	Note      *string `json:"note"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// ToResponse converts a ProgressUpdate entity to a ProgressResponse.
func ToResponse(p *ProgressUpdate) ProgressResponse {
	return ProgressResponse{
		ID:        p.ID,
		TaskID:    p.TaskID,
		UserID:    p.UserID,
		Progress:  p.Progress,
		Note:      p.Note,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}
