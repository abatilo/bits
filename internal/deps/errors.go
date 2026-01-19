package deps

import "fmt"

// CycleError indicates adding a dependency would create a cycle.
type CycleError struct {
	From string
	To   string
}

func (e CycleError) Error() string {
	return fmt.Sprintf("adding dependency %s -> %s would create a cycle", e.From, e.To)
}

// BlockedError indicates a task has unclosed dependencies.
type BlockedError struct {
	ID        string
	BlockedBy []string
}

func (e BlockedError) Error() string {
	return fmt.Sprintf("task %s is blocked by: %v", e.ID, e.BlockedBy)
}
