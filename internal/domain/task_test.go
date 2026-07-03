package domain

import "testing"

func TestTaskStatusIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status TaskStatus
		valid  bool
	}{
		{TaskStatusTodo, true},
		{TaskStatusInProgress, true},
		{TaskStatusDone, true},
		{TaskStatus("blocked"), false},
		{TaskStatus(""), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.valid {
			t.Errorf("TaskStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}
