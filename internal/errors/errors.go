package errors

import "fmt"

// ErrNotInitialized indicates ~/.bits doesn't exist.
type ErrNotInitialized struct{}

func (e ErrNotInitialized) Error() string {
	return "bits not initialized: run 'bits init' first"
}

// ErrAlreadyInitialized indicates ~/.bits already exists.
type ErrAlreadyInitialized struct{}

func (e ErrAlreadyInitialized) Error() string {
	return "bits already initialized"
}

// ErrTaskNotFound indicates the task ID doesn't match any file.
type ErrTaskNotFound struct {
	ID string
}

func (e ErrTaskNotFound) Error() string {
	return fmt.Sprintf("task not found: %s", e.ID)
}

// ErrAlreadyExists indicates an ID collision.
type ErrAlreadyExists struct {
	ID string
}

func (e ErrAlreadyExists) Error() string {
	return fmt.Sprintf("task already exists: %s", e.ID)
}

// ErrBlocked indicates a task has unclosed dependencies.
type ErrBlocked struct {
	ID           string
	BlockedBy    []string
}

func (e ErrBlocked) Error() string {
	return fmt.Sprintf("task %s is blocked by: %v", e.ID, e.BlockedBy)
}

// ErrCycle indicates adding a dependency would create a cycle.
type ErrCycle struct {
	From string
	To   string
}

func (e ErrCycle) Error() string {
	return fmt.Sprintf("adding dependency %s -> %s would create a cycle", e.From, e.To)
}

// ErrInvalidStatus indicates the task has the wrong status for the operation.
type ErrInvalidStatus struct {
	ID       string
	Current  string
	Expected string
}

func (e ErrInvalidStatus) Error() string {
	return fmt.Sprintf("task %s has status '%s', expected '%s'", e.ID, e.Current, e.Expected)
}

// ErrMissingReason indicates close was called without a reason.
type ErrMissingReason struct{}

func (e ErrMissingReason) Error() string {
	return "close reason is required"
}

// ErrInvalidPriority indicates an invalid priority value.
type ErrInvalidPriority struct {
	Value string
}

func (e ErrInvalidPriority) Error() string {
	return fmt.Sprintf("invalid priority: %s (valid: critical, high, medium, low)", e.Value)
}

// ErrNotInRepo indicates the command was run outside a git repository.
type ErrNotInRepo struct{}

func (e ErrNotInRepo) Error() string {
	return "not in a git repository (bits requires a project root)"
}
