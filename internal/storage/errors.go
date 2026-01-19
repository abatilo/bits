package storage

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

// NotInRepoError indicates the command was run outside a git repository.
type NotInRepoError struct{}

func (e NotInRepoError) Error() string {
	return "not in a git repository (bits requires a project root)"
}
