package storage

import (
	"bytes"
	"strings"
	"time"

	"github.com/abatilo/bits/internal/task"
	"gopkg.in/yaml.v3"
)

const frontmatterDelimiter = "---"

// taskFrontmatter is the YAML-serializable portion of a task.
type taskFrontmatter struct {
	ID          string        `yaml:"id"`
	Title       string        `yaml:"title"`
	Status      task.Status   `yaml:"status"`
	Priority    task.Priority `yaml:"priority"`
	CreatedAt   string        `yaml:"created_at"`
	ClosedAt    *string       `yaml:"closed_at,omitempty"`
	CloseReason *string       `yaml:"close_reason,omitempty"`
	DependsOn   []string      `yaml:"depends_on,omitempty"`
}

// ParseMarkdown parses a markdown file with YAML frontmatter into a Task.
func ParseMarkdown(content []byte) (*task.Task, error) {
	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != frontmatterDelimiter {
		return nil, &parseError{"missing YAML frontmatter"}
	}

	// Find closing delimiter
	var frontmatterEnd int
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == frontmatterDelimiter {
			frontmatterEnd = i
			break
		}
	}
	if frontmatterEnd == 0 {
		return nil, &parseError{"unclosed YAML frontmatter"}
	}

	// Parse YAML
	yamlContent := strings.Join(lines[1:frontmatterEnd], "\n")
	var fm taskFrontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, &parseError{"invalid YAML: " + err.Error()}
	}

	// Parse timestamps
	createdAt, err := parseTime(fm.CreatedAt)
	if err != nil {
		return nil, &parseError{"invalid created_at: " + err.Error()}
	}

	var closedAt *time.Time
	if fm.ClosedAt != nil {
		t, err := parseTime(*fm.ClosedAt)
		if err != nil {
			return nil, &parseError{"invalid closed_at: " + err.Error()}
		}
		closedAt = &t
	}

	// Extract description (everything after frontmatter)
	var description string
	if frontmatterEnd+1 < len(lines) {
		description = strings.TrimSpace(strings.Join(lines[frontmatterEnd+1:], "\n"))
	}

	return &task.Task{
		ID:          fm.ID,
		Title:       fm.Title,
		Status:      fm.Status,
		Priority:    fm.Priority,
		CreatedAt:   createdAt,
		ClosedAt:    closedAt,
		CloseReason: fm.CloseReason,
		DependsOn:   fm.DependsOn,
		Description: description,
	}, nil
}

// SerializeMarkdown converts a Task to markdown with YAML frontmatter.
func SerializeMarkdown(t *task.Task) ([]byte, error) {
	fm := taskFrontmatter{
		ID:          t.ID,
		Title:       t.Title,
		Status:      t.Status,
		Priority:    t.Priority,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		CloseReason: t.CloseReason,
		DependsOn:   t.DependsOn,
	}
	if t.ClosedAt != nil {
		s := t.ClosedAt.Format(time.RFC3339)
		fm.ClosedAt = &s
	}

	var buf bytes.Buffer
	buf.WriteString(frontmatterDelimiter + "\n")

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(fm); err != nil {
		return nil, err
	}
	enc.Close()

	buf.WriteString(frontmatterDelimiter + "\n")

	if t.Description != "" {
		buf.WriteString("\n")
		buf.WriteString(t.Description)
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// parseError represents a parsing error.
type parseError struct {
	msg string
}

func (e *parseError) Error() string {
	return e.msg
}

// parseTime tries to parse a time string in common formats.
func parseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &parseError{"unrecognized time format"}
}
