package comment

import "time"

// Comment represents a task comment entity.
type Comment struct {
	ID        string     `json:"id"`
	TaskID    string     `json:"task_id"`
	UserID    string     `json:"user_id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

// CreateCommentRequest is the DTO for creating a comment.
type CreateCommentRequest struct {
	Content string `json:"content" validate:"required"`
}

// UpdateCommentRequest is the DTO for updating a comment.
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required"`
}

// CommentResponse is the public DTO for comment data.
type CommentResponse struct {
	ID        string `json:"id"`
	TaskID    string `json:"task_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ToResponse converts a Comment entity to a CommentResponse.
func ToResponse(c *Comment) CommentResponse {
	return CommentResponse{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	}
}
