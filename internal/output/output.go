package output

import "github.com/abatilo/bits/internal/task"

// Formatter defines the interface for output formatting.
type Formatter interface {
	FormatTask(t *task.Task) string
	FormatTaskList(tasks []*task.Task) string
	FormatError(err error) string
	FormatMessage(msg string) string
}
