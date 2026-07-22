package activity

import (
	"encoding/json"
	"time"
)

// Activity represents a task activity log entry.
type Activity struct {
	ID        string          `json:"id"`
	TaskID    *string         `json:"task_id"`
	UserID    string          `json:"user_id"`
	Action    string          `json:"action"`
	OldValue  *string         `json:"old_value"`
	NewValue  *string         `json:"new_value"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}

// ActivityResponse is the public DTO for activity data.
type ActivityResponse struct {
	ID        string          `json:"id"`
	TaskID    *string         `json:"task_id"`
	UserID    string          `json:"user_id"`
	Action    string          `json:"action"`
	OldValue  *string         `json:"old_value"`
	NewValue  *string         `json:"new_value"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt string          `json:"created_at"`
}

// ToResponse converts an Activity entity to an ActivityResponse.
func ToResponse(a *Activity) ActivityResponse {
	return ActivityResponse{
		ID:        a.ID,
		TaskID:    a.TaskID,
		UserID:    a.UserID,
		Action:    a.Action,
		OldValue:  a.OldValue,
		NewValue:  a.NewValue,
		Metadata:  a.Metadata,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}
}
