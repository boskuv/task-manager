package task

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
	svc := newTestService(newMockTaskRepo(), teams)

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
	svc := newTestService(newMockTaskRepo(), teams)

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

	svc := newTestService(newMockTaskRepo(), newMockTeamRepo())

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
	svc := newTestService(newMockTaskRepo(), teams)

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
	svc := newTestService(newMockTaskRepo(), teams)

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
	svc := newTestService(tasks, teams)

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
	svc := newTestService(tasks, newMockTeamRepo())

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
	svc := newTestService(tasks, teams)

	_, err := svc.Update(context.Background(), 1, 1, UpdateInput{AssigneeID: &assigneeID})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestUpdateOrphanAssigneeFixed(t *testing.T) {
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

	assigneeID := int64(99)
	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{
		ID:         1,
		TeamID:     1,
		Title:      "Task",
		Status:     domain.TaskStatusTodo,
		AssigneeID: &assigneeID,
		CreatedBy:  1,
	}
	tasks.orphans[1] = true

	newAssigneeID := int64(2)
	svc := newTestService(tasks, teams)

	task, err := svc.Update(context.Background(), 1, 1, UpdateInput{AssigneeID: &newAssigneeID})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if task.AssigneeID == nil || *task.AssigneeID != 2 {
		t.Fatalf("task.AssigneeID = %v, want 2", task.AssigneeID)
	}
}

func TestUpdateBlockedOnOrphanAssignee(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}

	assigneeID := int64(99)
	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{
		ID:         1,
		TeamID:     1,
		Title:      "Task",
		Status:     domain.TaskStatusTodo,
		AssigneeID: &assigneeID,
		CreatedBy:  1,
	}
	tasks.orphans[1] = true

	title := "Updated"
	svc := newTestService(tasks, teams)

	_, err := svc.Update(context.Background(), 1, 1, UpdateInput{Title: &title})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestUpdateRecordsHistory(t *testing.T) {
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
		Description: "old desc",
		Status:      domain.TaskStatusTodo,
		CreatedBy:   1,
	}

	history := newMockTaskHistoryRepo()
	title := "New"
	description := "new desc"
	status := "in_progress"
	assigneeID := int64(2)
	svc := NewService(tasks, teams, history, nil)

	_, err := svc.Update(context.Background(), 1, 1, UpdateInput{
		Title:       &title,
		Description: &description,
		Status:      &status,
		AssigneeID:  &assigneeID,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if len(history.entries) != 4 {
		t.Fatalf("history entries = %d, want 4", len(history.entries))
	}

	want := []domain.TaskHistory{
		{TaskID: 1, ChangedBy: 1, Field: domain.HistoryFieldTitle, OldValue: "Old", NewValue: "New"},
		{TaskID: 1, ChangedBy: 1, Field: domain.HistoryFieldDescription, OldValue: "old desc", NewValue: "new desc"},
		{TaskID: 1, ChangedBy: 1, Field: domain.HistoryFieldStatus, OldValue: "todo", NewValue: "in_progress"},
		{TaskID: 1, ChangedBy: 1, Field: domain.HistoryFieldAssignee, OldValue: "", NewValue: "2"},
	}
	for i, entry := range history.entries {
		if entry.TaskID != want[i].TaskID ||
			entry.ChangedBy != want[i].ChangedBy ||
			entry.Field != want[i].Field ||
			entry.OldValue != want[i].OldValue ||
			entry.NewValue != want[i].NewValue {
			t.Errorf("entry[%d] = %+v, want %+v", i, entry, want[i])
		}
	}
}

func TestUpdateSkipsHistoryWhenNothingChanged(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}

	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{
		ID:          1,
		TeamID:      1,
		Title:       "Same",
		Description: "desc",
		Status:      domain.TaskStatusTodo,
		CreatedBy:   1,
	}

	history := newMockTaskHistoryRepo()
	title := "Same"
	description := "desc"
	status := "todo"
	svc := NewService(tasks, teams, history, nil)

	_, err := svc.Update(context.Background(), 1, 1, UpdateInput{
		Title:       &title,
		Description: &description,
		Status:      &status,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if len(history.entries) != 0 {
		t.Fatalf("history entries = %d, want 0", len(history.entries))
	}
}

func TestListHistorySuccess(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}

	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{
		ID:        1,
		TeamID:    1,
		Title:     "Task",
		Status:    domain.TaskStatusTodo,
		CreatedBy: 1,
	}

	history := newMockTaskHistoryRepo()
	history.entries = []domain.TaskHistory{
		{ID: 1, TaskID: 1, ChangedBy: 1, Field: domain.HistoryFieldStatus, OldValue: "todo", NewValue: "done"},
	}

	svc := NewService(tasks, teams, history, nil)

	entries, err := svc.ListHistory(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("ListHistory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries count = %d, want 1", len(entries))
	}
	if entries[0].Field != domain.HistoryFieldStatus {
		t.Errorf("entries[0].Field = %q, want status", entries[0].Field)
	}
}

func TestListHistoryForbiddenForNonMember(t *testing.T) {
	t.Parallel()

	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{ID: 1, TeamID: 1, Title: "Task", Status: domain.TaskStatusTodo, CreatedBy: 1}
	svc := newTestService(tasks, newMockTeamRepo())

	_, err := svc.ListHistory(context.Background(), 99, 1)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestListHistoryNotFound(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}
	svc := newTestService(newMockTaskRepo(), teams)

	_, err := svc.ListHistory(context.Background(), 1, 99)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
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
	svc := newTestService(tasks, teams)

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

	svc := newTestService(newMockTaskRepo(), newMockTeamRepo())

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
	svc := newTestService(newMockTaskRepo(), teams)

	_, err := svc.List(context.Background(), 1, ListInput{TeamID: 1, Status: "invalid"})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func newTestService(tasks repository.TaskRepository, teams repository.TeamRepository) *Service {
	return NewService(tasks, teams, newMockTaskHistoryRepo(), nil)
}

func TestListUsesCache(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}

	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{ID: 1, TeamID: 1, Title: "Cached", Status: domain.TaskStatusTodo, CreatedBy: 1}

	cache := newMockTaskListCache()
	filter := repository.TaskFilter{TeamID: 1, Page: 1, PageSize: 10}
	cache.store[cacheKey(filter)] = repository.TaskListResult{
		Items: []domain.Task{{ID: 99, TeamID: 1, Title: "From cache", Status: domain.TaskStatusTodo, CreatedBy: 1}},
		Total: 1,
	}

	svc := NewService(tasks, teams, newMockTaskHistoryRepo(), cache)

	result, err := svc.List(context.Background(), 1, ListInput{TeamID: 1, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Title != "From cache" {
		t.Fatalf("result = %+v, want cached item", result)
	}
	if tasks.listCalls != 0 {
		t.Fatalf("tasks.listCalls = %d, want 0", tasks.listCalls)
	}
}

func TestListPopulatesCacheOnMiss(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}

	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{ID: 1, TeamID: 1, Title: "From DB", Status: domain.TaskStatusTodo, CreatedBy: 1}

	cache := newMockTaskListCache()
	svc := NewService(tasks, teams, newMockTaskHistoryRepo(), cache)

	filter := repository.TaskFilter{TeamID: 1, Page: 1, PageSize: 10}
	_, err := svc.List(context.Background(), 1, ListInput{TeamID: 1, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if tasks.listCalls != 1 {
		t.Fatalf("tasks.listCalls = %d, want 1", tasks.listCalls)
	}

	cached, ok := cache.store[cacheKey(filter)]
	if !ok {
		t.Fatal("expected cache entry after miss")
	}
	if len(cached.Items) != 1 || cached.Items[0].Title != "From DB" {
		t.Fatalf("cached = %+v, want DB item", cached)
	}
}

func TestCreateInvalidatesTeamCache(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}

	cache := newMockTaskListCache()
	filter := repository.TaskFilter{TeamID: 1, Page: 1, PageSize: 10}
	cache.store[cacheKey(filter)] = repository.TaskListResult{Total: 1}
	cache.store[cacheKey(repository.TaskFilter{TeamID: 2, Page: 1, PageSize: 10})] = repository.TaskListResult{Total: 2}

	svc := NewService(newMockTaskRepo(), teams, newMockTaskHistoryRepo(), cache)

	_, err := svc.Create(context.Background(), 1, CreateInput{TeamID: 1, Title: "New task"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, ok := cache.store[cacheKey(filter)]; ok {
		t.Fatal("expected team 1 cache to be invalidated")
	}
	if _, ok := cache.store[cacheKey(repository.TaskFilter{TeamID: 2, Page: 1, PageSize: 10})]; !ok {
		t.Fatal("expected team 2 cache to remain")
	}
}

func TestUpdateInvalidatesTeamCache(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}

	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{ID: 1, TeamID: 1, Title: "Task", Status: domain.TaskStatusTodo, CreatedBy: 1}

	cache := newMockTaskListCache()
	filter := repository.TaskFilter{TeamID: 1, Page: 1, PageSize: 10}
	cache.store[cacheKey(filter)] = repository.TaskListResult{Total: 1}

	svc := NewService(tasks, teams, newMockTaskHistoryRepo(), cache)

	title := "Updated"
	_, err := svc.Update(context.Background(), 1, 1, UpdateInput{Title: &title})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, ok := cache.store[cacheKey(filter)]; ok {
		t.Fatal("expected team cache to be invalidated after update")
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
	tasks     map[int64]domain.Task
	orphans   map[int64]bool
	nextID    int64
	listCalls int
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{
		tasks:   make(map[int64]domain.Task),
		orphans: make(map[int64]bool),
		nextID:  1,
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
	m.listCalls++
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

func (m *mockTaskRepo) HasOrphanAssignee(_ context.Context, taskID int64) (bool, error) {
	return m.orphans[taskID], nil
}

func (m *mockTaskRepo) ListOrphanAssignees(_ context.Context) ([]domain.Task, error) {
	items := make([]domain.Task, 0)
	for id, orphan := range m.orphans {
		if !orphan {
			continue
		}
		task, ok := m.tasks[id]
		if !ok {
			continue
		}
		items = append(items, task)
	}
	return items, nil
}

type mockTaskHistoryRepo struct {
	entries []domain.TaskHistory
	nextID  int64
}

func newMockTaskHistoryRepo() *mockTaskHistoryRepo {
	return &mockTaskHistoryRepo{nextID: 1}
}

func (m *mockTaskHistoryRepo) Insert(_ context.Context, entry domain.TaskHistory) (domain.TaskHistory, error) {
	entry.ID = m.nextID
	m.nextID++
	entry.CreatedAt = time.Now().UTC()
	m.entries = append(m.entries, entry)
	return entry, nil
}

func (m *mockTaskHistoryRepo) ListByTaskID(_ context.Context, taskID int64) ([]domain.TaskHistory, error) {
	items := make([]domain.TaskHistory, 0)
	for _, entry := range m.entries {
		if entry.TaskID == taskID {
			items = append(items, entry)
		}
	}
	return items, nil
}

type mockTaskListCache struct {
	store map[string]repository.TaskListResult
}

func newMockTaskListCache() *mockTaskListCache {
	return &mockTaskListCache{store: make(map[string]repository.TaskListResult)}
}

func (m *mockTaskListCache) Get(_ context.Context, filter repository.TaskFilter) (repository.TaskListResult, bool, error) {
	result, ok := m.store[cacheKey(filter)]
	return result, ok, nil
}

func (m *mockTaskListCache) Set(_ context.Context, filter repository.TaskFilter, result repository.TaskListResult) error {
	m.store[cacheKey(filter)] = result
	return nil
}

func (m *mockTaskListCache) InvalidateTeam(_ context.Context, teamID int64) error {
	prefix := fmt.Sprintf("%d|", teamID)
	for key := range m.store {
		if strings.HasPrefix(key, prefix) {
			delete(m.store, key)
		}
	}
	return nil
}

func cacheKey(filter repository.TaskFilter) string {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	status := "*"
	if filter.Status != nil {
		status = string(*filter.Status)
	}

	assignee := "*"
	if filter.AssigneeID != nil {
		assignee = strconv.FormatInt(*filter.AssigneeID, 10)
	}

	return fmt.Sprintf("%d|%s|%s|%d|%d", filter.TeamID, status, assignee, page, pageSize)
}
