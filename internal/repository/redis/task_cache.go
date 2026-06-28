package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/boskuv/task-manager/internal/repository"
)

const taskListCacheTTL = 5 * time.Minute

const (
	defaultTaskPageSize = 20
	maxTaskPageSize     = 100
)

var _ repository.TaskListCache = (*TaskCache)(nil)

// TaskCache implements repository.TaskListCache with Redis.
type TaskCache struct {
	client *goredis.Client
}

// NewTaskCache returns a Redis-backed task list cache.
func NewTaskCache(client *goredis.Client) *TaskCache {
	return &TaskCache{client: client}
}

// Get returns a cached task list for the given filter.
// The second return value is true when the key exists in Redis.
func (c *TaskCache) Get(ctx context.Context, filter repository.TaskFilter) (repository.TaskListResult, bool, error) {
	data, err := c.client.Get(ctx, taskListCacheKey(filter)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return repository.TaskListResult{}, false, nil
		}
		return repository.TaskListResult{}, false, fmt.Errorf("get task list cache: %w", err)
	}

	var result repository.TaskListResult
	if err := json.Unmarshal(data, &result); err != nil {
		return repository.TaskListResult{}, false, fmt.Errorf("decode task list cache: %w", err)
	}

	return result, true, nil
}

// Set stores a task list result under the filter key with a 5-minute TTL.
func (c *TaskCache) Set(ctx context.Context, filter repository.TaskFilter, result repository.TaskListResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("encode task list cache: %w", err)
	}

	if err := c.client.Set(ctx, taskListCacheKey(filter), data, taskListCacheTTL).Err(); err != nil {
		return fmt.Errorf("set task list cache: %w", err)
	}

	return nil
}

func taskListCacheKey(filter repository.TaskFilter) string {
	return fmt.Sprintf("tasks:team:%d:filter:%s", filter.TeamID, hashTaskFilter(filter))
}

func hashTaskFilter(filter repository.TaskFilter) string {
	page, pageSize := normalizeTaskPagination(filter.Page, filter.PageSize)

	var b strings.Builder
	b.WriteString("status=")
	if filter.Status != nil {
		b.WriteString(string(*filter.Status))
	} else {
		b.WriteByte('*')
	}
	b.WriteString("|assignee=")
	if filter.AssigneeID != nil {
		b.WriteString(strconv.FormatInt(*filter.AssigneeID, 10))
	} else {
		b.WriteByte('*')
	}
	b.WriteString("|page=")
	b.WriteString(strconv.Itoa(page))
	b.WriteString("|page_size=")
	b.WriteString(strconv.Itoa(pageSize))

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func normalizeTaskPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultTaskPageSize
	}
	if pageSize > maxTaskPageSize {
		pageSize = maxTaskPageSize
	}
	return page, pageSize
}
