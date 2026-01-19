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

// Priority sort order constants (lower = higher priority).
const (
	priorityOrderCritical = 0
	priorityOrderHigh     = 1
	priorityOrderMedium   = 2
	priorityOrderLow      = 3
	priorityOrderUnknown  = 4
)

// PriorityOrder returns the sort order for a priority (lower = higher priority).
func PriorityOrder(p Priority) int {
	switch p {
	case PriorityCritical:
		return priorityOrderCritical
	case PriorityHigh:
		return priorityOrderHigh
	case PriorityMedium:
		return priorityOrderMedium
	case PriorityLow:
		return priorityOrderLow
	default:
		return priorityOrderUnknown
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
