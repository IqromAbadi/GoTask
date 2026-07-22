package task

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockRepo implements Repository for testing.
type mockRepo struct {
	tasks map[string]*Task
}

func newMockRepo() *mockRepo {
	return &mockRepo{tasks: make(map[string]*Task)}
}

func (m *mockRepo) Create(ctx context.Context, t *Task) error {
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = time.Now().UTC()
	m.tasks[t.ID] = t
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*Task, error) {
	t, ok := m.tasks[id.String()]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (m *mockRepo) Update(ctx context.Context, t *Task) error {
	if existing, ok := m.tasks[t.ID]; ok {
		existing.Title = t.Title
		existing.Description = t.Description
		existing.Priority = t.Priority
		existing.DueDate = t.DueDate
		existing.EstimatedMin = t.EstimatedMin
		existing.UpdatedAt = time.Now().UTC()
		m.tasks[t.ID] = existing
		return nil
	}
	return nil
}

func (m *mockRepo) UpdateStatus(ctx context.Context, id, listID uuid.UUID, status string) (*Task, error) {
	t, ok := m.tasks[id.String()]
	if !ok {
		return nil, nil
	}
	if status == "in_progress" && t.StartedAt == nil {
		now := time.Now().UTC()
		t.StartedAt = &now
	}
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	return t, nil
}

func (m *mockRepo) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) (*Task, error) {
	t, ok := m.tasks[id.String()]
	if !ok {
		return nil, nil
	}
	t.Progress = progress
	t.UpdatedAt = time.Now().UTC()
	return t, nil
}

func (m *mockRepo) SoftDelete(ctx context.Context, id, listID uuid.UUID) error {
	if t, ok := m.tasks[id.String()]; ok && t.ListID == listID.String() {
		now := time.Now().UTC()
		t.DeletedAt = &now
		return nil
	}
	return nil
}

func (m *mockRepo) MarkDone(ctx context.Context, id uuid.UUID) (*Task, error) {
	t, ok := m.tasks[id.String()]
	if !ok {
		return nil, nil
	}
	t.Status = "done"
	t.Progress = 100
	now := time.Now().UTC()
	t.CompletedAt = &now
	return t, nil
}

func (m *mockRepo) Reopen(ctx context.Context, id, listID uuid.UUID) (*Task, error) {
	t, ok := m.tasks[id.String()]
	if !ok || t.ListID != listID.String() {
		return nil, nil
	}
	t.Status = "in_progress"
	t.CompletedAt = nil
	t.UpdatedAt = time.Now().UTC()
	return t, nil
}

func (m *mockRepo) List(ctx context.Context, listID, userID uuid.UUID, filter TaskFilter) ([]Task, int, error) {
	var result []Task
	for _, t := range m.tasks {
		if t.ListID == listID.String() && t.DeletedAt == nil {
			if filter.Status != "" && t.Status != filter.Status {
				continue
			}
			if filter.Priority != "" && t.Priority != filter.Priority {
				continue
			}
			result = append(result, *t)
		}
	}
	return result, len(result), nil
}

func (m *mockRepo) GetBoard(ctx context.Context, listID, userID uuid.UUID) (map[string][]Task, error) {
	board := map[string][]Task{
		"backlog":     {},
		"todo":        {},
		"in_progress": {},
		"review":      {},
		"done":        {},
	}
	for _, t := range m.tasks {
		if t.ListID == listID.String() && t.DeletedAt == nil {
			board[t.Status] = append(board[t.Status], *t)
		}
	}
	return board, nil
}

func newTestTaskService() *Service {
	return NewService(newMockRepo(), nil, nil)
}

func TestCreateTask_Defaults(t *testing.T) {
	svc := newTestTaskService()
	userID := uuid.New()
	listID := uuid.New()

	req := CreateTaskRequest{
		Title: "Test Task",
	}

	task, err := svc.Create(context.Background(), listID, userID, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.Status != "backlog" {
		t.Errorf("expected default status 'backlog', got: %s", task.Status)
	}
	if task.Priority != "medium" {
		t.Errorf("expected default priority 'medium', got: %s", task.Priority)
	}
	if task.Progress != 0 {
		t.Errorf("expected default progress 0, got: %d", task.Progress)
	}
}

func TestCreateTask_InvalidStatus(t *testing.T) {
	svc := newTestTaskService()
	req := CreateTaskRequest{
		Title:  "Test",
		Status: "invalid_status",
	}
	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), req)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateTask_InvalidPriority(t *testing.T) {
	svc := newTestTaskService()
	req := CreateTaskRequest{
		Title:    "Test",
		Priority: "invalid_priority",
	}
	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), req)
	if err == nil {
		t.Fatal("expected error for invalid priority")
	}
}

func TestStatusTransition_Valid(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		ok   bool
	}{
		{"backlog to todo", "backlog", "todo", true},
		{"todo to backlog", "todo", "backlog", true},
		{"todo to in_progress", "todo", "in_progress", true},
		{"in_progress to todo", "in_progress", "todo", true},
		{"in_progress to review", "in_progress", "review", true},
		{"review to in_progress", "review", "in_progress", true},
		{"review to done", "review", "done", true},
		{"done to in_progress", "done", "in_progress", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidTransition(tt.from, tt.to); got != tt.ok {
				t.Errorf("isValidTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.ok)
			}
		})
	}
}

func TestStatusTransition_Invalid(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
	}{
		{"todo to done", "todo", "done"},
		{"backlog to review", "backlog", "review"},
		{"backlog to done", "backlog", "done"},
		{"todo to review", "todo", "review"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidTransition(tt.from, tt.to); got != false {
				t.Errorf("isValidTransition(%s, %s) = %v, want false", tt.from, tt.to, got)
			}
		})
	}
}

func TestUpdateStatus_DoneNotAllowedDirectly(t *testing.T) {
	svc := newTestTaskService()
	userID := uuid.New()
	listID := uuid.New()

	// Create task and move to review
	req := CreateTaskRequest{Title: "Test", Status: "in_progress"}
	created, _ := svc.Create(context.Background(), listID, userID, req)
	taskID, _ := uuid.Parse(created.ID)

	// Try to set status to done directly
	_, err := svc.UpdateStatus(context.Background(), taskID, listID, "done", userID)
	if err == nil {
		t.Fatal("expected error: cannot go to done directly")
	}
}

func TestUpdateStatus_ReviewRequires100Percent(t *testing.T) {
	svc := newTestTaskService()
	userID := uuid.New()
	listID := uuid.New()

	req := CreateTaskRequest{Title: "Test", Status: "in_progress"}
	created, _ := svc.Create(context.Background(), listID, userID, req)
	taskID, _ := uuid.Parse(created.ID)

	// Try to move to review with 0% progress
	_, err := svc.UpdateStatus(context.Background(), taskID, listID, "review", userID)
	if err == nil {
		t.Fatal("expected error: progress must be 100%")
	}
}

func TestReopenTask(t *testing.T) {
	svc := newTestTaskService()
	userID := uuid.New()
	listID := uuid.New()

	req := CreateTaskRequest{Title: "Completed Task", Status: "done"}
	created, _ := svc.Create(context.Background(), listID, userID, req)
	taskID, _ := uuid.Parse(created.ID)

	// Manually set to done in mock
	svc.repo.MarkDone(context.Background(), taskID)

	reopened, err := svc.Reopen(context.Background(), taskID, listID, userID)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	if reopened.Status != "in_progress" {
		t.Errorf("expected 'in_progress', got: %s", reopened.Status)
	}
	if reopened.CompletedAt != nil {
		t.Error("expected completed_at to be nil after reopen")
	}
}

func TestReopen_NonDoneTask(t *testing.T) {
	svc := newTestTaskService()
	userID := uuid.New()
	listID := uuid.New()

	req := CreateTaskRequest{Title: "Active Task"}
	created, _ := svc.Create(context.Background(), listID, userID, req)
	taskID, _ := uuid.Parse(created.ID)

	_, err := svc.Reopen(context.Background(), taskID, listID, userID)
	if err == nil {
		t.Fatal("expected error when reopening non-done task")
	}
}

func TestFilterTasks(t *testing.T) {
	svc := newTestTaskService()
	userID := uuid.New()
	listID := uuid.New()

	// Create tasks with different statuses
	svc.Create(context.Background(), listID, userID, CreateTaskRequest{Title: "Task 1", Status: "backlog"})
	svc.Create(context.Background(), listID, userID, CreateTaskRequest{Title: "Task 2", Status: "todo"})
	svc.Create(context.Background(), listID, userID, CreateTaskRequest{Title: "Task 3", Status: "in_progress"})

	// Filter by status
	filter := TaskFilter{Status: "todo", Page: 1, Limit: 10}
	tasks, total, err := svc.List(context.Background(), listID, userID, filter)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 task with status todo, got: %d", total)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got: %d", len(tasks))
	}
}

func TestValidSortFields(t *testing.T) {
	validFields := []string{"created_at", "updated_at", "due_date", "priority", "progress", "title"}
	invalidFields := []string{"id", "status", "random", "DROP TABLE", "1=1"}

	for _, f := range validFields {
		if !ValidSortFields[f] {
			t.Errorf("expected %q to be valid sort field", f)
		}
	}
	for _, f := range invalidFields {
		if ValidSortFields[f] {
			t.Errorf("expected %q to be invalid sort field", f)
		}
	}
}
