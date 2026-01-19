//nolint:revive // Package name intentionally matches stdlib for domain clarity
package errors

import "fmt"

// NotInitializedError indicates ~/.bits doesn't exist.
type NotInitializedError struct{}

func (e NotInitializedError) Error() string {
	return "bits not initialized: run 'bits init' first"
}

// AlreadyInitializedError indicates ~/.bits already exists.
type AlreadyInitializedError struct{}

func (e AlreadyInitializedError) Error() string {
	return "bits already initialized"
}

// TaskNotFoundError indicates the task ID doesn't match any file.
type TaskNotFoundError struct {
	ID string
}

func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("task not found: %s", e.ID)
}

// AlreadyExistsError indicates an ID collision.
type AlreadyExistsError struct {
	ID string
}

func (e AlreadyExistsError) Error() string {
	return fmt.Sprintf("task already exists: %s", e.ID)
}

// BlockedError indicates a task has unclosed dependencies.
type BlockedError struct {
	ID        string
	BlockedBy []string
}

func (e BlockedError) Error() string {
	return fmt.Sprintf("task %s is blocked by: %v", e.ID, e.BlockedBy)
}

// CycleError indicates adding a dependency would create a cycle.
type CycleError struct {
	From string
	To   string
}

func (e CycleError) Error() string {
	return fmt.Sprintf("adding dependency %s -> %s would create a cycle", e.From, e.To)
}

// InvalidStatusError indicates the task has the wrong status for the operation.
type InvalidStatusError struct {
	ID       string
	Current  string
	Expected string
}

func (e InvalidStatusError) Error() string {
	return fmt.Sprintf("task %s has status '%s', expected '%s'", e.ID, e.Current, e.Expected)
}

// MissingReasonError indicates close was called without a reason.
type MissingReasonError struct{}

func (e MissingReasonError) Error() string {
	return "close reason is required"
}

// InvalidPriorityError indicates an invalid priority value.
type InvalidPriorityError struct {
	Value string
}

func (e InvalidPriorityError) Error() string {
	return fmt.Sprintf("invalid priority: %s (valid: critical, high, medium, low)", e.Value)
}

// NotInRepoError indicates the command was run outside a git repository.
type NotInRepoError struct{}

func (e NotInRepoError) Error() string {
	return "not in a git repository (bits requires a project root)"
}

// ActiveTaskExistsError indicates a task is already active.
type ActiveTaskExistsError struct {
	ID    string
	Title string
}

func (e ActiveTaskExistsError) Error() string {
	return fmt.Sprintf("task %s (%s) is already active; release or close it first", e.ID, e.Title)
}
