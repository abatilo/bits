package output

import (
	"encoding/json"
	"time"

	"github.com/abatilo/bits/internal/task"
)

// JSONFormatter formats output as JSON.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSONFormatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// taskJSON is the JSON representation of a task.
type taskJSON struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	CreatedAt   string   `json:"created_at"`
	ClosedAt    *string  `json:"closed_at,omitempty"`
	CloseReason *string  `json:"close_reason,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty"`
	Description string   `json:"description,omitempty"`
}

func toTaskJSON(t *task.Task) taskJSON {
	tj := taskJSON{
		ID:          t.ID,
		Title:       t.Title,
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		CloseReason: t.CloseReason,
		DependsOn:   t.DependsOn,
		Description: t.Description,
	}
	if t.ClosedAt != nil {
		s := t.ClosedAt.Format(time.RFC3339)
		tj.ClosedAt = &s
	}
	return tj
}

// FormatTask formats a single task as JSON.
func (f *JSONFormatter) FormatTask(t *task.Task) string {
	data, _ := json.MarshalIndent(toTaskJSON(t), "", "  ")
	return string(data) + "\n"
}

// FormatTaskList formats a list of tasks as JSON.
func (f *JSONFormatter) FormatTaskList(tasks []*task.Task) string {
	jsonTasks := make([]taskJSON, len(tasks))
	for i, t := range tasks {
		jsonTasks[i] = toTaskJSON(t)
	}
	data, _ := json.MarshalIndent(jsonTasks, "", "  ")
	return string(data) + "\n"
}

// errorJSON is the JSON representation of an error.
type errorJSON struct {
	Error string `json:"error"`
}

// FormatError formats an error as JSON.
func (f *JSONFormatter) FormatError(err error) string {
	data, _ := json.MarshalIndent(errorJSON{Error: err.Error()}, "", "  ")
	return string(data) + "\n"
}

// messageJSON is the JSON representation of a message.
type messageJSON struct {
	Message string `json:"message"`
}

// FormatMessage formats a simple message as JSON.
func (f *JSONFormatter) FormatMessage(msg string) string {
	data, _ := json.MarshalIndent(messageJSON{Message: msg}, "", "  ")
	return string(data) + "\n"
}

// graphNodeJSON is the JSON representation of a graph node.
type graphNodeJSON struct {
	ID       string          `json:"id"`
	Title    string          `json:"title"`
	Status   string          `json:"status"`
	Priority string          `json:"priority"`
	Children []graphNodeJSON `json:"children,omitempty"`
}

func toGraphNodeJSON(node GraphNode) graphNodeJSON {
	children := make([]graphNodeJSON, len(node.Children))
	for i, c := range node.Children {
		children[i] = toGraphNodeJSON(c)
	}
	return graphNodeJSON{
		ID:       node.Task.ID,
		Title:    node.Task.Title,
		Status:   string(node.Task.Status),
		Priority: string(node.Task.Priority),
		Children: children,
	}
}

// FormatGraph formats a dependency graph as JSON.
func (f *JSONFormatter) FormatGraph(nodes []GraphNode) string {
	jsonNodes := make([]graphNodeJSON, len(nodes))
	for i, n := range nodes {
		jsonNodes[i] = toGraphNodeJSON(n)
	}
	data, _ := json.MarshalIndent(jsonNodes, "", "  ")
	return string(data) + "\n"
}
