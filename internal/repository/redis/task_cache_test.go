package redis

import (
	"testing"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/repository"
)

func TestTaskListCacheKeyFormat(t *testing.T) {
	t.Parallel()

	filter := repository.TaskFilter{
		TeamID:   42,
		Page:     1,
		PageSize: 20,
	}

	key := taskListCacheKey(filter)
	if !stringsHasPrefix(key, "tasks:team:42:filter:") {
		t.Fatalf("key = %q, want prefix tasks:team:42:filter:", key)
	}

	hashPart := key[len("tasks:team:42:filter:"):]
	if len(hashPart) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hashPart))
	}
}

func TestHashTaskFilterStableForEquivalentPagination(t *testing.T) {
	t.Parallel()

	status := domain.TaskStatusTodo
	assigneeID := int64(5)

	base := repository.TaskFilter{
		TeamID:     1,
		Status:     &status,
		AssigneeID: &assigneeID,
		Page:       1,
		PageSize:   20,
	}
	withDefaults := repository.TaskFilter{
		TeamID:     1,
		Status:     &status,
		AssigneeID: &assigneeID,
		Page:       0,
		PageSize:   0,
	}

	if hashTaskFilter(base) != hashTaskFilter(withDefaults) {
		t.Fatal("expected equivalent pagination to produce the same filter hash")
	}
}

func TestHashTaskFilterDiffersByFilter(t *testing.T) {
	t.Parallel()

	statusTodo := domain.TaskStatusTodo
	statusDone := domain.TaskStatusDone

	filters := []repository.TaskFilter{
		{TeamID: 1, Page: 1, PageSize: 20},
		{TeamID: 1, Status: &statusTodo, Page: 1, PageSize: 20},
		{TeamID: 1, Status: &statusDone, Page: 1, PageSize: 20},
		{TeamID: 1, Page: 2, PageSize: 20},
		{TeamID: 1, Page: 1, PageSize: 50},
	}

	seen := make(map[string]struct{}, len(filters))
	for _, filter := range filters {
		hash := hashTaskFilter(filter)
		if _, ok := seen[hash]; ok {
			t.Fatalf("duplicate hash %q for filter %+v", hash, filter)
		}
		seen[hash] = struct{}{}
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
