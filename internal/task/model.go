package task

import "time"

// Task represents a task entity.
type Task struct {
	ID           string     `json:"id"`
	ListID       string     `json:"list_id"`
	CreatedBy    string     `json:"created_by"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	Status       string     `json:"status"`
	Priority     string     `json:"priority"`
	Progress     int        `json:"progress"`
	DueDate      *time.Time `json:"due_date"`
	EstimatedMin *int       `json:"estimated_minutes"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

// ValidStatuses lists all valid task statuses.
var ValidStatuses = []string{"backlog", "todo", "in_progress", "review", "done"}

// ValidPriorities lists all valid task priorities.
var ValidPriorities = []string{"low", "medium", "high", "urgent"}

// ValidSortFields lists all allowed sort fields.
var ValidSortFields = map[string]bool{
	"created_at": true,
	"updated_at": true,
	"due_date":   true,
	"priority":   true,
	"progress":   true,
	"title":      true,
}

// IsValidStatus checks if a status string is valid.
func IsValidStatus(s string) bool {
	for _, vs := range ValidStatuses {
		if vs == s {
			return true
		}
	}
	return false
}

// IsValidPriority checks if a priority string is valid.
func IsValidPriority(p string) bool {
	for _, vp := range ValidPriorities {
		if vp == p {
			return true
		}
	}
	return false
}
