package deps

import (
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
		pi := task.PriorityOrder(ready[i].Priority)
		pj := task.PriorityOrder(ready[j].Priority)
		if pi != pj {
			return pi < pj
		}
		return ready[i].CreatedAt.Before(ready[j].CreatedAt)
	})

	return ready
}

// Dependents returns IDs of tasks that depend on the given task.
func (g *Graph) Dependents(id string) []string {
	var dependents []string
	for _, t := range g.tasks {
		for _, depID := range t.DependsOn {
			if depID == id {
				dependents = append(dependents, t.ID)
				break
			}
		}
	}
	return dependents
}

// ValidateAddDep validates adding a dependency from -> to.
func (g *Graph) ValidateAddDep(from, to string) error {
	if g.tasks[from] == nil {
		return bitserrors.ErrTaskNotFound{ID: from}
	}
	if g.tasks[to] == nil {
		return bitserrors.ErrTaskNotFound{ID: to}
	}
	if g.WouldCreateCycle(from, to) {
		return bitserrors.ErrCycle{From: from, To: to}
	}
	return nil
}

// BuildTree builds a tree representation of tasks for graph display.
// Returns root nodes (tasks with no dependents).
func (g *Graph) BuildTree() []output.GraphNode {
	// Find root tasks (tasks that nothing depends on)
	hasParent := make(map[string]bool)
	for _, t := range g.tasks {
		for _, depID := range t.DependsOn {
			hasParent[depID] = false // Mark as a dependency
		}
	}

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
		isRoot := true
		for _, other := range g.tasks {
			for _, depID := range other.DependsOn {
				if depID == t.ID {
					isRoot = false
					break
				}
			}
			if !isRoot {
				break
			}
		}
		if isRoot {
			roots = append(roots, t)
		}
	}

	// Sort roots by priority then created_at
	sort.Slice(roots, func(i, j int) bool {
		pi := task.PriorityOrder(roots[i].Priority)
		pj := task.PriorityOrder(roots[j].Priority)
		if pi != pj {
			return pi < pj
		}
		return roots[i].CreatedAt.Before(roots[j].CreatedAt)
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
			pi := task.PriorityOrder(node.Children[i].Task.Priority)
			pj := task.PriorityOrder(node.Children[j].Task.Priority)
			if pi != pj {
				return pi < pj
			}
			return node.Children[i].Task.CreatedAt.Before(node.Children[j].Task.CreatedAt)
		})

		return node
	}

	var nodes []output.GraphNode
	for _, root := range roots {
		nodes = append(nodes, buildNode(root, make(map[string]bool)))
	}
	return nodes
}
