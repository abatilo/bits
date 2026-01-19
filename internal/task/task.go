package task

import "time"

// Status represents the current state of a task.
type Status string

const (
	StatusOpen   Status = "open"
	StatusActive Status = "active"
	StatusClosed Status = "closed"
)

// Priority represents the importance level of a task.
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

// PriorityOrder returns the sort order for a priority (lower = higher priority).
func PriorityOrder(p Priority) int {
	switch p {
	case PriorityCritical:
		return 0
	case PriorityHigh:
		return 1
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 3
	default:
		return 4
	}
}

// Task represents a tracked work item.
type Task struct {
	ID          string     `yaml:"id"`
	Title       string     `yaml:"title"`
	Status      Status     `yaml:"status"`
	Priority    Priority   `yaml:"priority"`
	CreatedAt   time.Time  `yaml:"created_at"`
	ClosedAt    *time.Time `yaml:"closed_at,omitempty"`
	CloseReason *string    `yaml:"close_reason,omitempty"`
	DependsOn   []string   `yaml:"depends_on,omitempty"`
	Description string     `yaml:"-"` // Stored as markdown body, not frontmatter
}

// IsValidStatus checks if a status string is valid.
func IsValidStatus(s Status) bool {
	switch s {
	case StatusOpen, StatusActive, StatusClosed:
		return true
	default:
		return false
	}
}

// IsValidPriority checks if a priority string is valid.
func IsValidPriority(p Priority) bool {
	switch p {
	case PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow:
		return true
	default:
		return false
	}
}
