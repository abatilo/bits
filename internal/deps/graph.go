package deps

import (
	"slices"
	"sort"

	bitserrors "github.com/abatilo/bits/internal/errors"
	"github.com/abatilo/bits/internal/output"
	"github.com/abatilo/bits/internal/task"
)

// Graph represents the dependency relationships between tasks.
type Graph struct {
	tasks map[string]*task.Task
}

// NewGraph creates a Graph from a list of tasks.
func NewGraph(tasks []*task.Task) *Graph {
	g := &Graph{
		tasks: make(map[string]*task.Task),
	}
	for _, t := range tasks {
		g.tasks[t.ID] = t
	}
	return g
}

// Get returns a task by ID.
func (g *Graph) Get(id string) *task.Task {
	return g.tasks[id]
}

// IsBlocked returns true if the task has any unclosed dependencies.
func (g *Graph) IsBlocked(id string) bool {
	t := g.tasks[id]
	if t == nil {
		return false
	}
	for _, depID := range t.DependsOn {
		dep := g.tasks[depID]
		if dep == nil {
			continue // Missing dependency is not blocking
		}
		if dep.Status != task.StatusClosed {
			return true
		}
	}
	return false
}

// BlockedBy returns the IDs of unclosed tasks that block this task.
func (g *Graph) BlockedBy(id string) []string {
	t := g.tasks[id]
	if t == nil {
		return nil
	}
	var blockers []string
	for _, depID := range t.DependsOn {
		dep := g.tasks[depID]
		if dep == nil {
			continue
		}
		if dep.Status != task.StatusClosed {
			blockers = append(blockers, depID)
		}
	}
	return blockers
}

// WouldCreateCycle checks if adding a dependency from -> to would create a cycle.
// Uses BFS from 'to' to see if we can reach 'from'.
func (g *Graph) WouldCreateCycle(from, to string) bool {
	// If adding from -> to, check if to can reach from through existing edges
	visited := make(map[string]bool)
	queue := []string{to}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == from {
			return true
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		t := g.tasks[current]
		if t == nil {
			continue
		}
		queue = append(queue, t.DependsOn...)
	}
	return false
}

// Ready returns all tasks that are open and have all dependencies closed.
func (g *Graph) Ready() []*task.Task {
	var ready []*task.Task
	for _, t := range g.tasks {
		if t.Status != task.StatusOpen {
			continue
		}
		if !g.IsBlocked(t.ID) {
			ready = append(ready, t)
		}
	}

	// Sort by priority then created_at
	sort.Slice(ready, func(i, j int) bool {
		return taskLess(ready[i], ready[j])
	})

	return ready
}

// Dependents returns IDs of tasks that depend on the given task.
func (g *Graph) Dependents(id string) []string {
	var dependents []string
	for _, t := range g.tasks {
		if slices.Contains(t.DependsOn, id) {
			dependents = append(dependents, t.ID)
		}
	}
	return dependents
}

// ValidateAddDep validates adding a dependency from -> to.
func (g *Graph) ValidateAddDep(from, to string) error {
	if g.tasks[from] == nil {
		return bitserrors.TaskNotFoundError{ID: from}
	}
	if g.tasks[to] == nil {
		return bitserrors.TaskNotFoundError{ID: to}
	}
	if g.WouldCreateCycle(from, to) {
		return bitserrors.CycleError{From: from, To: to}
	}
	return nil
}

// taskLess returns true if task a should be sorted before task b.
// Sorts by priority first (critical < high < medium < low), then by creation time.
func taskLess(a, b *task.Task) bool {
	pa := task.PriorityOrder(a.Priority)
	pb := task.PriorityOrder(b.Priority)
	if pa != pb {
		return pa < pb
	}
	return a.CreatedAt.Before(b.CreatedAt)
}

// BuildTree builds a tree representation of tasks for graph display.
// Returns root nodes (tasks with no dependents).
func (g *Graph) BuildTree() []output.GraphNode {
	// Tasks that are dependencies have children (tasks that depend on them)
	children := make(map[string][]*task.Task)
	for _, t := range g.tasks {
		for _, depID := range t.DependsOn {
			children[depID] = append(children[depID], t)
		}
	}

	// Find roots: tasks that no one depends on
	var roots []*task.Task
	for _, t := range g.tasks {
		if len(g.Dependents(t.ID)) == 0 {
			roots = append(roots, t)
		}
	}

	// Sort roots by priority then created_at
	sort.Slice(roots, func(i, j int) bool {
		return taskLess(roots[i], roots[j])
	})

	// Build tree from roots
	var buildNode func(t *task.Task, visited map[string]bool) output.GraphNode
	buildNode = func(t *task.Task, visited map[string]bool) output.GraphNode {
		node := output.GraphNode{Task: t}
		if visited[t.ID] {
			return node // Prevent infinite recursion
		}
		visited[t.ID] = true

		// Children are tasks that have this task as a dependency
		for _, child := range children[t.ID] {
			node.Children = append(node.Children, buildNode(child, visited))
		}

		// Sort children
		sort.Slice(node.Children, func(i, j int) bool {
			return taskLess(node.Children[i].Task, node.Children[j].Task)
		})

		return node
	}

	var nodes []output.GraphNode
	for _, root := range roots {
		nodes = append(nodes, buildNode(root, make(map[string]bool)))
	}
	return nodes
}
