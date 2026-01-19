//nolint:testpackage // Tests require internal access for thorough testing
package errors

import (
	"testing"
)

func TestActiveTaskExistsError(t *testing.T) {
	tests := []struct {
		name string
		err  ActiveTaskExistsError
		want string
	}{
		{
			name: "formats error with id and title",
			err:  ActiveTaskExistsError{ID: "abc123", Title: "My Task"},
			want: "task abc123 (My Task) is already active; release or close it first",
		},
		{
			name: "handles empty title",
			err:  ActiveTaskExistsError{ID: "def456", Title: ""},
			want: "task def456 () is already active; release or close it first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ActiveTaskExistsError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTaskNotFoundError(t *testing.T) {
	err := TaskNotFoundError{ID: "xyz789"}
	want := "task not found: xyz789"
	if got := err.Error(); got != want {
		t.Errorf("TaskNotFoundError.Error() = %q, want %q", got, want)
	}
}

func TestBlockedError(t *testing.T) {
	err := BlockedError{ID: "task1", BlockedBy: []string{"task2", "task3"}}
	want := "task task1 is blocked by: [task2 task3]"
	if got := err.Error(); got != want {
		t.Errorf("BlockedError.Error() = %q, want %q", got, want)
	}
}

func TestInvalidStatusError(t *testing.T) {
	err := InvalidStatusError{ID: "abc", Current: "closed", Expected: "open"}
	want := "task abc has status 'closed', expected 'open'"
	if got := err.Error(); got != want {
		t.Errorf("InvalidStatusError.Error() = %q, want %q", got, want)
	}
}
