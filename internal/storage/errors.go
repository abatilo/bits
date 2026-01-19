package storage

import "fmt"

// TaskNotFoundError indicates the task ID doesn't match any file.
type TaskNotFoundError struct {
	ID string
}

func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("task not found: %s", e.ID)
}

// NotInRepoError indicates the command was run outside a git repository.
type NotInRepoError struct{}

func (e NotInRepoError) Error() string {
	return "not in a git repository (bits requires a project root)"
}
