//nolint:testpackage // Tests require internal access for thorough testing
package deps

import (
	"testing"
	"time"

	"github.com/abatilo/bits/internal/task"
)

func makeTask(id string, status task.Status, deps ...string) *task.Task {
	return &task.Task{
		ID:        id,
		Title:     "Task " + id,
		Status:    status,
		Priority:  task.PriorityMedium,
		CreatedAt: time.Now(),
		DependsOn: deps,
	}
}

func TestIsBlocked(t *testing.T) {
	tasks := []*task.Task{
		makeTask("a", task.StatusOpen),
		makeTask("b", task.StatusOpen, "a"), // b depends on a
		makeTask("c", task.StatusClosed),
		makeTask("d", task.StatusOpen, "c"), // d depends on c (closed)
	}

	g := NewGraph(tasks)

	tests := []struct {
		id      string
		blocked bool
	}{
		{"a", false}, // No dependencies
		{"b", true},  // Depends on open task
		{"c", false}, // No dependencies
		{"d", false}, // Depends on closed task
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := g.IsBlocked(tt.id); got != tt.blocked {
				t.Errorf("IsBlocked(%q) = %v, want %v", tt.id, got, tt.blocked)
			}
		})
	}
}

func TestBlockedBy(t *testing.T) {
	tasks := []*task.Task{
		makeTask("a", task.StatusOpen),
		makeTask("b", task.StatusActive),
		makeTask("c", task.StatusOpen, "a", "b"), // c depends on a and b
	}

	g := NewGraph(tasks)

	blockers := g.BlockedBy("c")
	if len(blockers) != 2 {
		t.Fatalf("BlockedBy length = %d, want 2", len(blockers))
	}
}

func TestWouldCreateCycle(t *testing.T) {
	// a -> b -> c (a depends on b, b depends on c)
	tasks := []*task.Task{
		makeTask("a", task.StatusOpen, "b"),
		makeTask("b", task.StatusOpen, "c"),
		makeTask("c", task.StatusOpen),
	}

	g := NewGraph(tasks)

	tests := []struct {
		from, to string
		cycle    bool
	}{
		{"c", "a", true},  // c -> a would create cycle (a -> b -> c -> a)
		{"c", "b", true},  // c -> b would create cycle (b -> c -> b)
		{"a", "c", false}, // a -> c is fine (already a -> b -> c)
		{"c", "d", false}, // d doesn't exist, no cycle
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			if got := g.WouldCreateCycle(tt.from, tt.to); got != tt.cycle {
				t.Errorf("WouldCreateCycle(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.cycle)
			}
		})
	}
}

func TestReady(t *testing.T) {
	tasks := []*task.Task{
		makeTask("a", task.StatusOpen),      // Ready
		makeTask("b", task.StatusOpen, "a"), // Blocked by a
		makeTask("c", task.StatusClosed),    // Closed
		makeTask("d", task.StatusOpen, "c"), // Ready (c is closed)
		makeTask("e", task.StatusActive),    // Active, not ready
	}

	g := NewGraph(tasks)
	ready := g.Ready()

	if len(ready) != 2 {
		t.Fatalf("Ready length = %d, want 2", len(ready))
	}

	ids := map[string]bool{}
	for _, r := range ready {
		ids[r.ID] = true
	}

	if !ids["a"] || !ids["d"] {
		t.Errorf("Ready should contain a and d, got %v", ids)
	}
}

func TestDependents(t *testing.T) {
	tasks := []*task.Task{
		makeTask("a", task.StatusOpen),
		makeTask("b", task.StatusOpen, "a"),
		makeTask("c", task.StatusOpen, "a"),
		makeTask("d", task.StatusOpen, "b"),
	}

	g := NewGraph(tasks)

	deps := g.Dependents("a")
	if len(deps) != 2 {
		t.Fatalf("Dependents(a) length = %d, want 2", len(deps))
	}

	deps = g.Dependents("b")
	if len(deps) != 1 {
		t.Fatalf("Dependents(b) length = %d, want 1", len(deps))
	}

	deps = g.Dependents("d")
	if len(deps) != 0 {
		t.Fatalf("Dependents(d) length = %d, want 0", len(deps))
	}
}

func TestValidateAddDep(t *testing.T) {
	tasks := []*task.Task{
		makeTask("a", task.StatusOpen, "b"),
		makeTask("b", task.StatusOpen),
	}

	g := NewGraph(tasks)

	// Valid: b -> a
	if err := g.ValidateAddDep("b", "a"); err == nil {
		t.Error("Expected cycle error for b -> a")
	}

	// Valid: add new dependency
	if err := g.ValidateAddDep("a", "b"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invalid: task doesn't exist
	if err := g.ValidateAddDep("x", "a"); err == nil {
		t.Error("Expected error for non-existent task")
	}
}
