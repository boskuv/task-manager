package task

import (
	"context"
	"strconv"
	"strings"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/repository"
)

const maxTaskTitleLength = 255

// CreateInput holds fields for creating a task.
type CreateInput struct {
	TeamID      int64
	Title       string
	Description string
	AssigneeID  *int64
}

// UpdateInput holds optional fields for updating a task.
type UpdateInput struct {
	Title       *string
	Description *string
	Status      *string
	AssigneeID  *int64
}

// ListInput holds filters and pagination for listing tasks.
type ListInput struct {
	TeamID     int64
	Status     string
	AssigneeID *int64
	Page       int
	PageSize   int
}

// Service handles task creation, updates, and listing.
type Service struct {
	tasks   repository.TaskRepository
	teams   repository.TeamRepository
	history repository.TaskHistoryRepository
}

// NewService creates a task use case service.
func NewService(tasks repository.TaskRepository, teams repository.TeamRepository, history repository.TaskHistoryRepository) *Service {
	return &Service{
		tasks:   tasks,
		teams:   teams,
		history: history,
	}
}

// Create creates a task when the actor is a team member.
func (s *Service) Create(ctx context.Context, actorUserID int64, input CreateInput) (domain.Task, error) {
	if actorUserID <= 0 || input.TeamID <= 0 {
		return domain.Task{}, domain.ErrInvalidInput
	}

	title, description, err := validateTaskContent(input.Title, input.Description)
	if err != nil {
		return domain.Task{}, err
	}

	if err := s.ensureTeamMember(ctx, input.TeamID, actorUserID); err != nil {
		return domain.Task{}, err
	}
	if err := s.ensureAssigneeInTeam(ctx, input.TeamID, input.AssigneeID); err != nil {
		return domain.Task{}, err
	}

	return s.tasks.Create(ctx, domain.Task{
		TeamID:      input.TeamID,
		Title:       title,
		Description: description,
		Status:      domain.TaskStatusTodo,
		AssigneeID:  input.AssigneeID,
		CreatedBy:   actorUserID,
	})
}

// Update updates a task when the actor is a member of the task's team.
func (s *Service) Update(ctx context.Context, actorUserID, taskID int64, input UpdateInput) (domain.Task, error) {
	if actorUserID <= 0 || taskID <= 0 {
		return domain.Task{}, domain.ErrInvalidInput
	}

	task, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	before := task

	if err := s.ensureTeamMember(ctx, task.TeamID, actorUserID); err != nil {
		return domain.Task{}, err
	}

	orphan, err := s.tasks.HasOrphanAssignee(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	if orphan && input.AssigneeID == nil {
		return domain.Task{}, domain.ErrInvalidInput
	}

	if input.Title != nil {
		title, _, err := validateTaskContent(*input.Title, task.Description)
		if err != nil {
			return domain.Task{}, err
		}
		task.Title = title
	}
	if input.Description != nil {
		task.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		status := domain.TaskStatus(strings.TrimSpace(strings.ToLower(*input.Status)))
		if !status.IsValid() {
			return domain.Task{}, domain.ErrInvalidInput
		}
		task.Status = status
	}
	if input.AssigneeID != nil {
		if err := s.ensureAssigneeInTeam(ctx, task.TeamID, input.AssigneeID); err != nil {
			return domain.Task{}, err
		}
		task.AssigneeID = input.AssigneeID
	}

	updated, err := s.tasks.Update(ctx, task)
	if err != nil {
		return domain.Task{}, err
	}

	if err := s.recordTaskHistory(ctx, actorUserID, before, updated); err != nil {
		return domain.Task{}, err
	}

	return updated, nil
}

// List returns paginated tasks for a team when the actor is a member.
func (s *Service) List(ctx context.Context, actorUserID int64, input ListInput) (repository.TaskListResult, error) {
	if actorUserID <= 0 || input.TeamID <= 0 {
		return repository.TaskListResult{}, domain.ErrInvalidInput
	}

	if err := s.ensureTeamMember(ctx, input.TeamID, actorUserID); err != nil {
		return repository.TaskListResult{}, err
	}

	filter := repository.TaskFilter{
		TeamID:     input.TeamID,
		AssigneeID: input.AssigneeID,
		Page:       input.Page,
		PageSize:   input.PageSize,
	}

	status := strings.TrimSpace(strings.ToLower(input.Status))
	if status != "" {
		taskStatus := domain.TaskStatus(status)
		if !taskStatus.IsValid() {
			return repository.TaskListResult{}, domain.ErrInvalidInput
		}
		filter.Status = &taskStatus
	}

	return s.tasks.List(ctx, filter)
}

func (s *Service) ensureTeamMember(ctx context.Context, teamID, userID int64) error {
	_, err := s.teams.GetMemberRole(ctx, teamID, userID)
	if err != nil {
		if err == domain.ErrNotFound {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}

func (s *Service) ensureAssigneeInTeam(ctx context.Context, teamID int64, assigneeID *int64) error {
	if assigneeID == nil {
		return nil
	}
	if *assigneeID <= 0 {
		return domain.ErrInvalidInput
	}

	_, err := s.teams.GetMemberRole(ctx, teamID, *assigneeID)
	if err != nil {
		if err == domain.ErrNotFound {
			return domain.ErrInvalidInput
		}
		return err
	}
	return nil
}

func (s *Service) recordTaskHistory(ctx context.Context, changedBy int64, before, after domain.Task) error {
	for _, entry := range buildTaskHistoryEntries(changedBy, before, after) {
		if _, err := s.history.Insert(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

func buildTaskHistoryEntries(changedBy int64, before, after domain.Task) []domain.TaskHistory {
	entries := make([]domain.TaskHistory, 0, 4)

	if before.Title != after.Title {
		entries = append(entries, domain.TaskHistory{
			TaskID:    after.ID,
			ChangedBy: changedBy,
			Field:     domain.HistoryFieldTitle,
			OldValue:  before.Title,
			NewValue:  after.Title,
		})
	}
	if before.Description != after.Description {
		entries = append(entries, domain.TaskHistory{
			TaskID:    after.ID,
			ChangedBy: changedBy,
			Field:     domain.HistoryFieldDescription,
			OldValue:  before.Description,
			NewValue:  after.Description,
		})
	}
	if before.Status != after.Status {
		entries = append(entries, domain.TaskHistory{
			TaskID:    after.ID,
			ChangedBy: changedBy,
			Field:     domain.HistoryFieldStatus,
			OldValue:  string(before.Status),
			NewValue:  string(after.Status),
		})
	}
	if !assigneeEqual(before.AssigneeID, after.AssigneeID) {
		entries = append(entries, domain.TaskHistory{
			TaskID:    after.ID,
			ChangedBy: changedBy,
			Field:     domain.HistoryFieldAssignee,
			OldValue:  formatAssigneeValue(before.AssigneeID),
			NewValue:  formatAssigneeValue(after.AssigneeID),
		})
	}

	return entries
}

func assigneeEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func formatAssigneeValue(assigneeID *int64) string {
	if assigneeID == nil {
		return ""
	}
	return strconv.FormatInt(*assigneeID, 10)
}

func validateTaskContent(title, description string) (string, string, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return "", "", domain.ErrInvalidInput
	}
	if len(title) > maxTaskTitleLength {
		return "", "", domain.ErrInvalidInput
	}
	return title, description, nil
}
