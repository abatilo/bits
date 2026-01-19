//nolint:testpackage // Tests require internal access for thorough testing
package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	bitserrors "github.com/abatilo/bits/internal/errors"
	"github.com/abatilo/bits/internal/task"
)

func TestParseMarkdown(t *testing.T) {
	content := []byte(`---
id: abc123
title: Test task
status: open
priority: high
created_at: 2024-01-15T10:30:00Z
---

This is the description.
`)

	task, err := ParseMarkdown(content)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}

	if task.ID != "abc123" {
		t.Errorf("ID = %q, want %q", task.ID, "abc123")
	}
	if task.Title != "Test task" {
		t.Errorf("Title = %q, want %q", task.Title, "Test task")
	}
	if task.Status != "open" {
		t.Errorf("Status = %q, want %q", task.Status, "open")
	}
	if task.Priority != "high" {
		t.Errorf("Priority = %q, want %q", task.Priority, "high")
	}
	if task.Description != "This is the description." {
		t.Errorf("Description = %q, want %q", task.Description, "This is the description.")
	}
}

func TestParseMarkdownWithDependencies(t *testing.T) {
	content := []byte(`---
id: abc123
title: Test task
status: open
priority: medium
created_at: 2024-01-15T10:30:00Z
depends_on:
  - def456
  - ghi789
---
`)

	task, err := ParseMarkdown(content)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}

	if len(task.DependsOn) != 2 {
		t.Fatalf("DependsOn length = %d, want 2", len(task.DependsOn))
	}
	if task.DependsOn[0] != "def456" || task.DependsOn[1] != "ghi789" {
		t.Errorf("DependsOn = %v, want [def456, ghi789]", task.DependsOn)
	}
}

func TestSerializeMarkdown(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	task := &task.Task{
		ID:          "abc123",
		Title:       "Test task",
		Status:      "open",
		Priority:    "high",
		CreatedAt:   now,
		Description: "Description here",
	}

	data, err := SerializeMarkdown(task)
	if err != nil {
		t.Fatalf("SerializeMarkdown failed: %v", err)
	}

	// Parse it back
	parsed, err := ParseMarkdown(data)
	if err != nil {
		t.Fatalf("ParseMarkdown failed: %v", err)
	}

	if parsed.ID != task.ID {
		t.Errorf("Round-trip ID = %q, want %q", parsed.ID, task.ID)
	}
	if parsed.Title != task.Title {
		t.Errorf("Round-trip Title = %q, want %q", parsed.Title, task.Title)
	}
	if parsed.Description != task.Description {
		t.Errorf("Round-trip Description = %q, want %q", parsed.Description, task.Description)
	}
}

func TestStoreOperations(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	basePath := filepath.Join(tmpDir, ".bits")
	store := NewStoreWithPath(basePath)

	// Test init
	if store.IsInitialized() {
		t.Error("Store should not be initialized yet")
	}

	if err := store.Init(false); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if !store.IsInitialized() {
		t.Error("Store should be initialized")
	}

	// Test create task
	tk, err := store.CreateTask("Test task", "Description", task.PriorityMedium)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if tk.ID == "" {
		t.Error("Task ID should not be empty")
	}

	// Test load
	loaded, err := store.Load(tk.ID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Title != tk.Title {
		t.Errorf("Loaded title = %q, want %q", loaded.Title, tk.Title)
	}

	// Test list
	tasks, err := store.List(StatusFilter{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("List length = %d, want 1", len(tasks))
	}

	// Test delete
	if err = store.Delete(tk.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	tasks, _ = store.List(StatusFilter{})
	if len(tasks) != 0 {
		t.Errorf("After delete, list length = %d, want 0", len(tasks))
	}
}

func TestStatusFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter StatusFilter
		status task.Status
		want   bool
	}{
		{"empty filter matches open", StatusFilter{}, task.StatusOpen, true},
		{"empty filter matches active", StatusFilter{}, task.StatusActive, true},
		{"empty filter matches closed", StatusFilter{}, task.StatusClosed, true},
		{"open filter matches open", StatusFilter{Open: true}, task.StatusOpen, true},
		{"open filter rejects active", StatusFilter{Open: true}, task.StatusActive, false},
		{"active filter matches active", StatusFilter{Active: true}, task.StatusActive, true},
		{"closed filter matches closed", StatusFilter{Closed: true}, task.StatusClosed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.Matches(tt.status); got != tt.want {
				t.Errorf("filter.Matches(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple path", "/Users/abatilo/myproject", "Users-abatilo-myproject"},
		{"path with spaces", "/Users/john doe/my project", "Users-john-doe-my-project"},
		{"path with special chars", "/home/user/my.project-v2", "home-user-my-project-v2"},
		{"root path", "/", ""},
		{"nested path", "/a/b/c/d/e", "a-b-c-d-e"},
		{"trailing slash", "/Users/abatilo/project/", "Users-abatilo-project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizePath(tt.input); got != tt.want {
				t.Errorf("SanitizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

//nolint:gocognit // Test setup/teardown requires multiple nested subtests
func TestFindProjectRoot(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Resolve symlinks in temp dir (macOS /var -> /private/var)
	var err error
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks: %v", err)
	}

	// Create nested directories: tmpDir/parent/child/grandchild
	grandchild := filepath.Join(tmpDir, "parent", "child", "grandchild")
	if err = os.MkdirAll(grandchild, 0o755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	t.Run("finds .git in current directory", func(t *testing.T) {
		// Create .git in grandchild
		gitDir := filepath.Join(grandchild, ".git")
		if err := os.Mkdir(gitDir, 0o755); err != nil { //nolint:govet // Intentional shadow in subtest
			t.Fatalf("Failed to create .git: %v", err)
		}
		defer os.RemoveAll(gitDir)

		// Change to grandchild and find root
		t.Chdir(grandchild)

		root, err := FindProjectRoot() //nolint:govet // Intentional shadow in subtest
		if err != nil {
			t.Fatalf("FindProjectRoot() error = %v", err)
		}
		if root != grandchild {
			t.Errorf("FindProjectRoot() = %q, want %q", root, grandchild)
		}
	})

	t.Run("finds .git in parent directory", func(t *testing.T) {
		// Create .git in parent
		parent := filepath.Join(tmpDir, "parent")
		gitDir := filepath.Join(parent, ".git")
		if err := os.Mkdir(gitDir, 0o755); err != nil { //nolint:govet // Intentional shadow in subtest
			t.Fatalf("Failed to create .git: %v", err)
		}
		defer os.RemoveAll(gitDir)

		// Change to grandchild and find root
		t.Chdir(grandchild)

		root, err := FindProjectRoot() //nolint:govet // Intentional shadow in subtest
		if err != nil {
			t.Fatalf("FindProjectRoot() error = %v", err)
		}
		if root != parent {
			t.Errorf("FindProjectRoot() = %q, want %q", root, parent)
		}
	})

	t.Run("returns error when no .git found", func(t *testing.T) {
		// Create a directory with no .git anywhere up the tree
		noGitDir := filepath.Join(tmpDir, "no-git-here")
		if err := os.Mkdir(noGitDir, 0o755); err != nil { //nolint:govet // Intentional shadow in subtest
			t.Fatalf("Failed to create directory: %v", err)
		}

		t.Chdir(noGitDir)

		_, err := FindProjectRoot() //nolint:govet // Intentional shadow in subtest
		if err == nil {
			t.Error("FindProjectRoot() should return error when no .git found")
		}
		if !errors.Is(err, bitserrors.NotInRepoError{}) {
			var notInRepo bitserrors.NotInRepoError
			if !errors.As(err, &notInRepo) {
				t.Errorf("FindProjectRoot() error = %v, want NotInRepoError", err)
			}
		}
	})
}
