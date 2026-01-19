//nolint:testpackage // Tests require internal access for thorough testing
package task

import (
	"testing"
	"time"
)

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status Status
		valid  bool
	}{
		{StatusOpen, true},
		{StatusActive, true},
		{StatusClosed, true},
		{Status("invalid"), false},
		{Status(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := IsValidStatus(tt.status); got != tt.valid {
				t.Errorf("IsValidStatus(%q) = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority Priority
		valid    bool
	}{
		{PriorityCritical, true},
		{PriorityHigh, true},
		{PriorityMedium, true},
		{PriorityLow, true},
		{Priority("invalid"), false},
		{Priority(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			if got := IsValidPriority(tt.priority); got != tt.valid {
				t.Errorf("IsValidPriority(%q) = %v, want %v", tt.priority, got, tt.valid)
			}
		})
	}
}

func TestPriorityOrder(t *testing.T) {
	// Critical should be lower (higher priority) than high
	if PriorityOrder(PriorityCritical) >= PriorityOrder(PriorityHigh) {
		t.Error("Critical should have lower order than High")
	}
	if PriorityOrder(PriorityHigh) >= PriorityOrder(PriorityMedium) {
		t.Error("High should have lower order than Medium")
	}
	if PriorityOrder(PriorityMedium) >= PriorityOrder(PriorityLow) {
		t.Error("Medium should have lower order than Low")
	}
}

func TestFindActive(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []*Task
		wantID  string
		wantNil bool
	}{
		{
			name:    "returns nil for nil slice",
			tasks:   nil,
			wantNil: true,
		},
		{
			name:    "returns nil for empty slice",
			tasks:   []*Task{},
			wantNil: true,
		},
		{
			name: "returns nil when no active tasks",
			tasks: []*Task{
				{ID: "t1", Status: StatusOpen},
				{ID: "t2", Status: StatusClosed},
			},
			wantNil: true,
		},
		{
			name: "returns active task when one exists",
			tasks: []*Task{
				{ID: "t1", Status: StatusOpen},
				{ID: "t2", Status: StatusActive},
				{ID: "t3", Status: StatusClosed},
			},
			wantID: "t2",
		},
		{
			name: "returns first active task when multiple exist",
			tasks: []*Task{
				{ID: "t1", Status: StatusActive},
				{ID: "t2", Status: StatusActive},
			},
			wantID: "t1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindActive(tt.tasks)
			if tt.wantNil {
				if got != nil {
					t.Errorf("FindActive() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("FindActive() = nil, want non-nil")
			}
			if got.ID != tt.wantID {
				t.Errorf("FindActive().ID = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	now := time.Now()

	// Should generate a unique ID
	id := GenerateID("Test task", now, func(_ string) bool { return false })
	if len(id) < 3 {
		t.Errorf("ID too short: %s", id)
	}
	if len(id) > 8 {
		t.Errorf("ID too long: %s", id)
	}

	// Should grow if collisions exist
	existingIDs := map[string]bool{}
	existsFn := func(id string) bool {
		return existingIDs[id]
	}

	id1 := GenerateID("Test", now, existsFn)
	existingIDs[id1] = true

	// Different title should generate different ID
	id2 := GenerateID("Different", now, existsFn)
	if id1 == id2 {
		t.Error("Expected different IDs for different titles")
	}
}
