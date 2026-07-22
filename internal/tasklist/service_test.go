package tasklist

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// mockRepo implements Repository for testing.
type mockRepo struct {
	lists map[string]*TaskList
}

func newMockRepo() *mockRepo {
	return &mockRepo{lists: make(map[string]*TaskList)}
}

func (m *mockRepo) Create(ctx context.Context, tl *TaskList) error {
	tl.ID = uuid.New().String()
	m.lists[tl.ID] = tl
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*TaskList, error) {
	tl, ok := m.lists[id.String()]
	if !ok || tl.UserID != userID.String() {
		return nil, nil
	}
	return tl, nil
}

func (m *mockRepo) List(ctx context.Context, userID uuid.UUID) ([]TaskList, error) {
	var result []TaskList
	for _, tl := range m.lists {
		if tl.UserID == userID.String() {
			result = append(result, *tl)
		}
	}
	return result, nil
}

func (m *mockRepo) Update(ctx context.Context, tl *TaskList) error {
	if existing, ok := m.lists[tl.ID]; ok && existing.UserID == tl.UserID {
		m.lists[tl.ID] = tl
		return nil
	}
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if tl, ok := m.lists[id.String()]; ok && tl.UserID == userID.String() {
		delete(m.lists, id.String())
		return nil
	}
	return nil
}

func (m *mockRepo) Archive(ctx context.Context, id, userID uuid.UUID) (*TaskList, error) {
	tl, ok := m.lists[id.String()]
	if !ok || tl.UserID != userID.String() {
		return nil, nil
	}
	tl.IsArchived = true
	return tl, nil
}

func (m *mockRepo) Restore(ctx context.Context, id, userID uuid.UUID) (*TaskList, error) {
	tl, ok := m.lists[id.String()]
	if !ok || tl.UserID != userID.String() {
		return nil, nil
	}
	tl.IsArchived = false
	return tl, nil
}

func TestCreateTaskList_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	userID := uuid.New()

	req := CreateTaskListRequest{
		Name:        "My Tasks",
		Description: strPtr("Personal tasks"),
	}

	tl, err := svc.Create(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tl.Name != "My Tasks" {
		t.Errorf("expected 'My Tasks', got: %s", tl.Name)
	}
	if tl.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetByID(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if err.Error() != "task list tidak ditemukan" {
		t.Errorf("got: %v", err)
	}
}

func TestUserCannotSeeOtherList(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	userA := uuid.New()
	userB := uuid.New()

	req := CreateTaskListRequest{Name: "A's List"}
	created, _ := svc.Create(context.Background(), userA, req)

	listID, _ := uuid.Parse(created.ID)

	// User B tries to access
	_, err := svc.GetByID(context.Background(), listID, userB)
	if err == nil {
		t.Fatal("user B should not see user A's list")
	}
}

func TestArchiveAndRestore(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	userID := uuid.New()

	req := CreateTaskListRequest{Name: "Test List"}
	created, _ := svc.Create(context.Background(), userID, req)
	listID, _ := uuid.Parse(created.ID)

	// Archive
	archived, err := svc.Archive(context.Background(), listID, userID)
	if err != nil {
		t.Fatalf("archive failed: %v", err)
	}
	if !archived.IsArchived {
		t.Error("expected IsArchived to be true")
	}

	// Restore
	restored, err := svc.Restore(context.Background(), listID, userID)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if restored.IsArchived {
		t.Error("expected IsArchived to be false")
	}
}

func TestDeleteList(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	userID := uuid.New()

	req := CreateTaskListRequest{Name: "Delete Me"}
	created, _ := svc.Create(context.Background(), userID, req)
	listID, _ := uuid.Parse(created.ID)

	err := svc.Delete(context.Background(), listID, userID)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Should not be found after delete
	_, err = svc.GetByID(context.Background(), listID, userID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func strPtr(s string) *string {
	return &s
}
