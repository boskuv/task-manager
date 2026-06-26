package task

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/repository"
)

func TestCreateSuccess(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleOwner,
	}
	teams.members[memberKey{teamID: 1, userID: 2}] = domain.TeamMember{
		TeamID: 1,
		UserID: 2,
		Role:   domain.TeamRoleMember,
	}

	assigneeID := int64(2)
	svc := NewService(newMockTaskRepo(), teams)

	task, err := svc.Create(context.Background(), 1, CreateInput{
		TeamID:      1,
		Title:       "  Fix bug  ",
		Description: " details ",
		AssigneeID:  &assigneeID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task.Title != "Fix bug" {
		t.Errorf("task.Title = %q, want Fix bug", task.Title)
	}
	if task.Description != "details" {
		t.Errorf("task.Description = %q, want details", task.Description)
	}
	if task.Status != domain.TaskStatusTodo {
		t.Errorf("task.Status = %q, want todo", task.Status)
	}
	if task.AssigneeID == nil || *task.AssigneeID != 2 {
		t.Errorf("task.AssigneeID = %v, want 2", task.AssigneeID)
	}
	if task.CreatedBy != 1 {
		t.Errorf("task.CreatedBy = %d, want 1", task.CreatedBy)
	}
}

func TestCreateWithoutAssignee(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}
	svc := NewService(newMockTaskRepo(), teams)

	task, err := svc.Create(context.Background(), 1, CreateInput{
		TeamID: 1,
		Title:  "Task",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task.AssigneeID != nil {
		t.Fatalf("task.AssigneeID = %v, want nil", task.AssigneeID)
	}
}

func TestCreateForbiddenForNonMember(t *testing.T) {
	t.Parallel()

	svc := NewService(newMockTaskRepo(), newMockTeamRepo())

	_, err := svc.Create(context.Background(), 99, CreateInput{
		TeamID: 1,
		Title:  "Task",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestCreateAssigneeNotInTeam(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}
	assigneeID := int64(2)
	svc := NewService(newMockTaskRepo(), teams)

	_, err := svc.Create(context.Background(), 1, CreateInput{
		TeamID:     1,
		Title:      "Task",
		AssigneeID: &assigneeID,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestCreateInvalidInput(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}
	svc := NewService(newMockTaskRepo(), teams)

	tests := []struct {
		name  string
		user  int64
		input CreateInput
	}{
		{name: "invalid user", user: 0, input: CreateInput{TeamID: 1, Title: "Task"}},
		{name: "invalid team", user: 1, input: CreateInput{TeamID: 0, Title: "Task"}},
		{name: "empty title", user: 1, input: CreateInput{TeamID: 1, Title: "  "}},
		{name: "title too long", user: 1, input: CreateInput{TeamID: 1, Title: strings.Repeat("a", maxTaskTitleLength+1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := svc.Create(context.Background(), tt.user, tt.input)
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestUpdateSuccess(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}
	teams.members[memberKey{teamID: 1, userID: 2}] = domain.TeamMember{
		TeamID: 1,
		UserID: 2,
		Role:   domain.TeamRoleMember,
	}

	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{
		ID:          1,
		TeamID:      1,
		Title:       "Old",
		Description: "desc",
		Status:      domain.TaskStatusTodo,
		CreatedBy:   1,
	}

	title := "New"
	status := "in_progress"
	assigneeID := int64(2)
	svc := NewService(tasks, teams)

	task, err := svc.Update(context.Background(), 1, 1, UpdateInput{
		Title:      &title,
		Status:     &status,
		AssigneeID: &assigneeID,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if task.Title != "New" {
		t.Errorf("task.Title = %q, want New", task.Title)
	}
	if task.Status != domain.TaskStatusInProgress {
		t.Errorf("task.Status = %q, want in_progress", task.Status)
	}
	if task.AssigneeID == nil || *task.AssigneeID != 2 {
		t.Errorf("task.AssigneeID = %v, want 2", task.AssigneeID)
	}
}

func TestUpdateForbiddenForNonMember(t *testing.T) {
	t.Parallel()

	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{ID: 1, TeamID: 1, Title: "Task", Status: domain.TaskStatusTodo, CreatedBy: 1}

	title := "New"
	svc := NewService(tasks, newMockTeamRepo())

	_, err := svc.Update(context.Background(), 99, 1, UpdateInput{Title: &title})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestUpdateAssigneeNotInTeam(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}
	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{ID: 1, TeamID: 1, Title: "Task", Status: domain.TaskStatusTodo, CreatedBy: 1}

	assigneeID := int64(2)
	svc := NewService(tasks, teams)

	_, err := svc.Update(context.Background(), 1, 1, UpdateInput{AssigneeID: &assigneeID})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestListSuccess(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}
	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{ID: 1, TeamID: 1, Title: "A", Status: domain.TaskStatusTodo, CreatedBy: 1}
	tasks.tasks[2] = domain.Task{ID: 2, TeamID: 1, Title: "B", Status: domain.TaskStatusDone, CreatedBy: 1}
	svc := NewService(tasks, teams)

	result, err := svc.List(context.Background(), 1, ListInput{TeamID: 1, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("result.Total = %d, want 2", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("items count = %d, want 2", len(result.Items))
	}
}

func TestListForbiddenForNonMember(t *testing.T) {
	t.Parallel()

	svc := NewService(newMockTaskRepo(), newMockTeamRepo())

	_, err := svc.List(context.Background(), 99, ListInput{TeamID: 1})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestListInvalidStatus(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}
	svc := NewService(newMockTaskRepo(), teams)

	_, err := svc.List(context.Background(), 1, ListInput{TeamID: 1, Status: "invalid"})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

type memberKey struct {
	teamID int64
	userID int64
}

type mockTeamRepo struct {
	members map[memberKey]domain.TeamMember
}

func newMockTeamRepo() *mockTeamRepo {
	return &mockTeamRepo{members: make(map[memberKey]domain.TeamMember)}
}

func (m *mockTeamRepo) Create(context.Context, domain.Team) (domain.Team, error) {
	panic("not implemented")
}

func (m *mockTeamRepo) ListByUserID(context.Context, int64) ([]domain.Team, error) {
	panic("not implemented")
}

func (m *mockTeamRepo) GetByID(context.Context, int64) (domain.Team, error) {
	panic("not implemented")
}

func (m *mockTeamRepo) AddMember(context.Context, domain.TeamMember) error {
	panic("not implemented")
}

func (m *mockTeamRepo) GetMemberRole(_ context.Context, teamID, userID int64) (domain.TeamRole, error) {
	member, ok := m.members[memberKey{teamID: teamID, userID: userID}]
	if !ok {
		return "", domain.ErrNotFound
	}
	return member.Role, nil
}

type mockTaskRepo struct {
	tasks  map[int64]domain.Task
	nextID int64
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{
		tasks:  make(map[int64]domain.Task),
		nextID: 1,
	}
}

func (m *mockTaskRepo) Create(_ context.Context, task domain.Task) (domain.Task, error) {
	task.ID = m.nextID
	m.nextID++
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskRepo) Update(_ context.Context, task domain.Task) (domain.Task, error) {
	if _, ok := m.tasks[task.ID]; !ok {
		return domain.Task{}, domain.ErrNotFound
	}
	task.UpdatedAt = time.Now().UTC()
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskRepo) GetByID(_ context.Context, id int64) (domain.Task, error) {
	task, ok := m.tasks[id]
	if !ok {
		return domain.Task{}, domain.ErrNotFound
	}
	return task, nil
}

func (m *mockTaskRepo) List(_ context.Context, filter repository.TaskFilter) (repository.TaskListResult, error) {
	items := make([]domain.Task, 0)
	for _, task := range m.tasks {
		if task.TeamID != filter.TeamID {
			continue
		}
		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}
		if filter.AssigneeID != nil {
			if task.AssigneeID == nil || *task.AssigneeID != *filter.AssigneeID {
				continue
			}
		}
		items = append(items, task)
	}
	return repository.TaskListResult{Items: items, Total: len(items)}, nil
}
