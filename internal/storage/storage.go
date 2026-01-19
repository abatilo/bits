package storage

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/abatilo/bits/internal/task"
)

const (
	bitsDir = ".bits"
	fileExt = ".md"
)

// Store handles task file operations.
type Store struct {
	basePath string
}

// NewStore creates a Store with a project-scoped path (~/.bits/<sanitized-project-root>/).
func NewStore() (*Store, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sanitized := SanitizePath(projectRoot)
	basePath := filepath.Join(home, bitsDir, sanitized)
	return &Store{basePath: basePath}, nil
}

// NewStoreWithPath creates a Store with a custom base path.
func NewStoreWithPath(path string) *Store {
	return &Store{basePath: path}
}

// BasePath returns the base path of the store.
func (s *Store) BasePath() string {
	return s.basePath
}

// IsInitialized checks if the bits directory exists.
func (s *Store) IsInitialized() bool {
	info, err := os.Stat(s.basePath)
	return err == nil && info.IsDir()
}

// EnsureInitialized creates the bits directory if it doesn't exist.
func (s *Store) EnsureInitialized() error {
	if s.IsInitialized() {
		return nil
	}
	//nolint:gosec // G301: 0755 is appropriate for user-accessible task directory
	return os.MkdirAll(s.basePath, 0o755)
}

// Init initializes the bits directory. With force=true, it wipes and recreates.
func (s *Store) Init(force bool) error {
	if force {
		if err := os.RemoveAll(s.basePath); err != nil {
			return err
		}
	}
	//nolint:gosec // G301: 0755 is appropriate for user-accessible task directory
	return os.MkdirAll(s.basePath, 0o755)
}

// taskPath returns the full path for a task file.
func (s *Store) taskPath(id string) string {
	return filepath.Join(s.basePath, id+fileExt)
}

// Exists checks if a task with the given ID exists.
func (s *Store) Exists(id string) bool {
	_, err := os.Stat(s.taskPath(id))
	return err == nil
}

// Save writes a task to disk.
func (s *Store) Save(t *task.Task) error {
	if err := s.EnsureInitialized(); err != nil {
		return err
	}
	content, err := SerializeMarkdown(t)
	if err != nil {
		return err
	}
	//nolint:gosec // G306: 0644 is appropriate for user-readable task files
	return os.WriteFile(s.taskPath(t.ID), content, 0o644)
}

// Load reads a task from disk.
func (s *Store) Load(id string) (*task.Task, error) {
	if err := s.EnsureInitialized(); err != nil {
		return nil, err
	}
	content, err := os.ReadFile(s.taskPath(id))
	if os.IsNotExist(err) {
		return nil, TaskNotFoundError{ID: id}
	}
	if err != nil {
		return nil, err
	}
	return ParseMarkdown(content)
}

// Delete removes a task file.
func (s *Store) Delete(id string) error {
	if err := s.EnsureInitialized(); err != nil {
		return err
	}
	err := os.Remove(s.taskPath(id))
	if os.IsNotExist(err) {
		return TaskNotFoundError{ID: id}
	}
	return err
}

// List returns all tasks, optionally filtered and sorted.
func (s *Store) List(filter StatusFilter) ([]*task.Task, error) {
	if err := s.EnsureInitialized(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, err
	}

	var tasks []*task.Task
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), fileExt) {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), fileExt)
		var t *task.Task
		t, err = s.Load(id)
		if err != nil {
			continue // Skip malformed files
		}
		if filter.Matches(t.Status) {
			tasks = append(tasks, t)
		}
	}

	// Sort by priority (highest first), then by created_at (oldest first)
	sort.Slice(tasks, func(i, j int) bool {
		pi := task.PriorityOrder(tasks[i].Priority)
		pj := task.PriorityOrder(tasks[j].Priority)
		if pi != pj {
			return pi < pj
		}
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})

	return tasks, nil
}

// AllIDs returns all task IDs (for ID generation collision checking).
func (s *Store) AllIDs() (map[string]bool, error) {
	if err := s.EnsureInitialized(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, err
	}

	ids := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), fileExt) {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), fileExt)
		ids[id] = true
	}
	return ids, nil
}

// RemoveDependency removes a dependency from all tasks that reference it.
func (s *Store) RemoveDependency(depID string) error {
	tasks, err := s.List(StatusFilter{})
	if err != nil {
		return err
	}

	for _, t := range tasks {
		originalLen := len(t.DependsOn)
		t.DependsOn = slices.DeleteFunc(t.DependsOn, func(d string) bool {
			return d == depID
		})
		if len(t.DependsOn) != originalLen {
			if err = s.Save(t); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateTask creates a new task with generated ID.
func (s *Store) CreateTask(title, description string, priority task.Priority) (*task.Task, error) {
	if err := s.EnsureInitialized(); err != nil {
		return nil, err
	}

	createdAt := time.Now().UTC()

	// Generate unique ID
	existingIDs, err := s.AllIDs()
	if err != nil {
		return nil, err
	}
	existsFn := func(id string) bool {
		return existingIDs[id]
	}
	id := task.GenerateID(title, createdAt, existsFn)

	t := &task.Task{
		ID:          id,
		Title:       title,
		Status:      task.StatusOpen,
		Priority:    priority,
		CreatedAt:   createdAt,
		Description: description,
	}

	if err = s.Save(t); err != nil {
		return nil, err
	}
	return t, nil
}

// StatusFilter controls which statuses to include in list results.
type StatusFilter struct {
	Open   bool
	Active bool
	Closed bool
}

// Matches returns true if the status should be included.
func (f StatusFilter) Matches(status task.Status) bool {
	// If no filter is set, include all
	if !f.Open && !f.Active && !f.Closed {
		return true
	}
	switch status {
	case task.StatusOpen:
		return f.Open
	case task.StatusActive:
		return f.Active
	case task.StatusClosed:
		return f.Closed
	default:
		return false
	}
}
