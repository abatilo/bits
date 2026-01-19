package main

import "fmt"

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

// ActiveTaskExistsError indicates a task is already active.
type ActiveTaskExistsError struct {
	ID    string
	Title string
}

func (e ActiveTaskExistsError) Error() string {
	return fmt.Sprintf("task %s (%s) is already active; release or close it first", e.ID, e.Title)
}
