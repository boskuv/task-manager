package model

import (
	"testing"
	"time"

	"github.com/boskuv/task-manager/internal/domain"
)

func TestUserMappingRoundTrip(t *testing.T) {
	t.Parallel()

	original := domain.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}

	got := UserToModel(original).ToDomain()
	if got != original {
		t.Fatalf("round trip mismatch:\n got  %+v\n want %+v", got, original)
	}
}

func TestTeamMappingRoundTrip(t *testing.T) {
	t.Parallel()

	original := domain.Team{
		ID:        10,
		Name:      "backend",
		CreatedBy: 1,
		CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	got := TeamToModel(original).ToDomain()
	if got != original {
		t.Fatalf("round trip mismatch:\n got  %+v\n want %+v", got, original)
	}
}

func TestTeamMemberMappingRoundTrip(t *testing.T) {
	t.Parallel()

	original := domain.TeamMember{
		TeamID:   10,
		UserID:   2,
		Role:     domain.TeamRoleAdmin,
		JoinedAt: time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
	}

	got := TeamMemberToModel(original).ToDomain()
	if got != original {
		t.Fatalf("round trip mismatch:\n got  %+v\n want %+v", got, original)
	}
}

func TestTaskMappingRoundTrip(t *testing.T) {
	t.Parallel()

	assigneeID := int64(5)
	original := domain.Task{
		ID:          100,
		TeamID:      10,
		Title:       "ship feature",
		Description: "details",
		Status:      domain.TaskStatusInProgress,
		AssigneeID:  &assigneeID,
		CreatedBy:   1,
		CreatedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
	}

	got := TaskToModel(original).ToDomain()
	if got != original {
		t.Fatalf("round trip mismatch:\n got  %+v\n want %+v", got, original)
	}
}

func TestTaskMappingRoundTripWithoutAssignee(t *testing.T) {
	t.Parallel()

	original := domain.Task{
		ID:     101,
		TeamID: 10,
		Title:  "unassigned",
		Status: domain.TaskStatusTodo,
	}

	got := TaskToModel(original).ToDomain()
	if got != original {
		t.Fatalf("round trip mismatch:\n got  %+v\n want %+v", got, original)
	}
}

func TestTaskHistoryMappingRoundTrip(t *testing.T) {
	t.Parallel()

	original := domain.TaskHistory{
		ID:        1,
		TaskID:    100,
		ChangedBy: 2,
		Field:     domain.HistoryFieldStatus,
		OldValue:  "todo",
		NewValue:  "in_progress",
		CreatedAt: time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
	}

	got := TaskHistoryToModel(original).ToDomain()
	if got != original {
		t.Fatalf("round trip mismatch:\n got  %+v\n want %+v", got, original)
	}
}

func TestCommentMappingRoundTrip(t *testing.T) {
	t.Parallel()

	original := domain.Comment{
		ID:        1,
		TaskID:    100,
		UserID:    2,
		Body:      "looks good",
		CreatedAt: time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC),
	}

	got := CommentToModel(original).ToDomain()
	if got != original {
		t.Fatalf("round trip mismatch:\n got  %+v\n want %+v", got, original)
	}
}
