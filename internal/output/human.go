package output

import (
	"fmt"
	"strings"

	"github.com/abatilo/bits/internal/task"
)

// HumanFormatter formats output for human-readable terminal display.
type HumanFormatter struct{}

// NewHumanFormatter creates a new HumanFormatter.
func NewHumanFormatter() *HumanFormatter {
	return &HumanFormatter{}
}

// FormatTask formats a single task for display.
func (f *HumanFormatter) FormatTask(t *task.Task) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[%s] %s\n", t.ID, t.Title))
	sb.WriteString(fmt.Sprintf("  Status:   %s\n", t.Status))
	sb.WriteString(fmt.Sprintf("  Priority: %s\n", t.Priority))
	sb.WriteString(fmt.Sprintf("  Created:  %s\n", t.CreatedAt.Format("2006-01-02 15:04")))

	if t.ClosedAt != nil {
		sb.WriteString(fmt.Sprintf("  Closed:   %s\n", t.ClosedAt.Format("2006-01-02 15:04")))
	}
	if t.CloseReason != nil && *t.CloseReason != "" {
		sb.WriteString(fmt.Sprintf("  Reason:   %s\n", *t.CloseReason))
	}
	if len(t.DependsOn) > 0 {
		sb.WriteString(fmt.Sprintf("  Depends:  %s\n", strings.Join(t.DependsOn, ", ")))
	}
	if t.Description != "" {
		sb.WriteString("\n")
		sb.WriteString(t.Description)
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatTaskList formats a list of tasks for display.
func (f *HumanFormatter) FormatTaskList(tasks []*task.Task) string {
	if len(tasks) == 0 {
		return "No tasks found.\n"
	}

	var sb strings.Builder
	for _, t := range tasks {
		sb.WriteString(f.formatTaskLine(t))
	}
	return sb.String()
}

// formatTaskLine formats a single task as a compact one-liner.
func (f *HumanFormatter) formatTaskLine(t *task.Task) string {
	statusIcon := f.statusIcon(t.Status)
	priorityMark := f.priorityMark(t.Priority)
	deps := ""
	if len(t.DependsOn) > 0 {
		deps = fmt.Sprintf(" [blocked by: %s]", strings.Join(t.DependsOn, ", "))
	}
	return fmt.Sprintf("%s %s [%s] %s%s\n", statusIcon, priorityMark, t.ID, t.Title, deps)
}

func (f *HumanFormatter) statusIcon(s task.Status) string {
	switch s {
	case task.StatusOpen:
		return "[ ]"
	case task.StatusActive:
		return "[*]"
	case task.StatusClosed:
		return "[X]"
	default:
		return "[?]"
	}
}

func (f *HumanFormatter) priorityMark(p task.Priority) string {
	switch p {
	case task.PriorityCritical:
		return "P0"
	case task.PriorityHigh:
		return "P1"
	case task.PriorityMedium:
		return "P2"
	case task.PriorityLow:
		return "P3"
	default:
		return "P?"
	}
}

// FormatError formats an error for display.
func (f *HumanFormatter) FormatError(err error) string {
	return fmt.Sprintf("Error: %s\n", err.Error())
}

// FormatMessage formats a simple message.
func (f *HumanFormatter) FormatMessage(msg string) string {
	return msg + "\n"
}

// FormatGraph formats a dependency graph as ASCII art.
func (f *HumanFormatter) FormatGraph(nodes []GraphNode) string {
	if len(nodes) == 0 {
		return "No tasks found.\n"
	}

	var sb strings.Builder
	for _, node := range nodes {
		f.formatGraphNode(&sb, node, "", true)
	}
	return sb.String()
}

func (f *HumanFormatter) formatGraphNode(sb *strings.Builder, node GraphNode, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if prefix == "" {
		connector = ""
	}

	statusIcon := f.statusIcon(node.Task.Status)
	fmt.Fprintf(sb, "%s%s%s [%s] %s\n", prefix, connector, statusIcon, node.Task.ID, node.Task.Title)

	childPrefix := prefix
	if prefix != "" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range node.Children {
		f.formatGraphNode(sb, child, childPrefix, i == len(node.Children)-1)
	}
}
