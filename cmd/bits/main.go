package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/spf13/cobra"

	"github.com/abatilo/bits/internal/deps"
	"github.com/abatilo/bits/internal/output"
	"github.com/abatilo/bits/internal/storage"
	"github.com/abatilo/bits/internal/task"
)

//nolint:gochecknoglobals // CLI flags and formatter are package-level by design
var (
	jsonOutput bool
	formatter  output.Formatter
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "bits",
		Short: "A minimal, file-based task tracker",
		Long:  "bits - A minimal, file-based task tracker optimized for AI agents.",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if jsonOutput {
				formatter = output.NewJSONFormatter()
			} else {
				formatter = output.NewHumanFormatter()
			}
		},
	}

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	rootCmd.AddCommand(
		initCmd(),
		addCmd(),
		listCmd(),
		showCmd(),
		readyCmd(),
		claimCmd(),
		releaseCmd(),
		closeCmd(),
		depCmd(),
		undepCmd(),
		pruneCmd(),
		rmCmd(),
		sessionCmd(),
		drainCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getStore() (*storage.Store, error) {
	return storage.NewStore()
}

func printOutput(s string) {
	os.Stdout.WriteString(s) //nolint:gosec // stdout write errors are unrecoverable
}

func printError(err error) {
	os.Stdout.WriteString(formatter.FormatError(err)) //nolint:gosec // stdout write errors are unrecoverable
	os.Exit(1)
}

// initCmd implements 'bits init'.
func initCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize bits task directory",
		Run: func(_ *cobra.Command, _ []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}
			if force {
				if err = store.Init(true); err != nil {
					printError(err)
				}
				printOutput(formatter.FormatMessage(fmt.Sprintf("Reinitialized bits at %s", store.BasePath())))
			} else {
				// Ensure initialized (implicit init)
				if err = store.EnsureInitialized(); err != nil {
					printError(err)
				}
				printOutput(formatter.FormatMessage(fmt.Sprintf("bits storage: %s", store.BasePath())))
			}
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Wipe and reinitialize")
	return cmd
}

// addCmd implements 'bits add'.
func addCmd() *cobra.Command {
	var description string
	var priority string
	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			p := task.Priority(priority)
			if !task.IsValidPriority(p) {
				printError(InvalidPriorityError{Value: priority})
			}

			t, err := store.CreateTask(args[0], description, p)
			if err != nil {
				printError(err)
			}
			printOutput(formatter.FormatTask(t))
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "Task description")
	cmd.Flags().StringVarP(&priority, "priority", "p", "medium", "Priority (critical, high, medium, low)")
	return cmd
}

// listCmd implements 'bits list'.
func listCmd() *cobra.Command {
	var showOpen, showActive, showClosed bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List tasks",
		Run: func(_ *cobra.Command, _ []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			filter := storage.StatusFilter{
				Open:   showOpen,
				Active: showActive,
				Closed: showClosed,
			}

			tasks, err := store.List(filter)
			if err != nil {
				printError(err)
			}
			printOutput(formatter.FormatTaskList(tasks))
		},
	}
	cmd.Flags().BoolVar(&showOpen, "open", false, "Show only open tasks")
	cmd.Flags().BoolVar(&showActive, "active", false, "Show only active tasks")
	cmd.Flags().BoolVar(&showClosed, "closed", false, "Show only closed tasks")
	return cmd
}

// showCmd implements 'bits show'.
func showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			t, err := store.Load(args[0])
			if err != nil {
				printError(err)
			}
			printOutput(formatter.FormatTask(t))
		},
	}
}

// readyCmd implements 'bits ready'.
func readyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ready",
		Short: "List tasks ready to be worked on",
		Run: func(_ *cobra.Command, _ []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			tasks, err := store.List(storage.StatusFilter{})
			if err != nil {
				printError(err)
			}

			graph := deps.NewGraph(tasks)
			ready := graph.Ready()
			printOutput(formatter.FormatTaskList(ready))
		},
	}
}

// claimCmd implements 'bits claim'.
func claimCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claim <id>",
		Short: "Claim a task (mark as active)",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			t, err := store.Load(args[0])
			if err != nil {
				printError(err)
			}

			if t.Status != task.StatusOpen {
				printError(InvalidStatusError{
					ID:       t.ID,
					Current:  string(t.Status),
					Expected: string(task.StatusOpen),
				})
			}

			// Check dependencies and active tasks
			tasks, err := store.List(storage.StatusFilter{})
			if err != nil {
				printError(err)
			}

			// Check if another task is already active
			if active := task.FindActive(tasks); active != nil {
				printError(ActiveTaskExistsError{ID: active.ID, Title: active.Title})
			}

			graph := deps.NewGraph(tasks)
			blockers := graph.BlockedBy(t.ID)
			if len(blockers) > 0 {
				printError(deps.BlockedError{ID: t.ID, BlockedBy: blockers})
			}

			t.Status = task.StatusActive
			if err = store.Save(t); err != nil {
				printError(err)
			}
			printOutput(formatter.FormatTask(t))
		},
	}
}

// releaseCmd implements 'bits release'.
func releaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "release <id>",
		Short: "Release a task (mark as open)",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			t, err := store.Load(args[0])
			if err != nil {
				printError(err)
			}

			if t.Status != task.StatusActive {
				printError(InvalidStatusError{
					ID:       t.ID,
					Current:  string(t.Status),
					Expected: string(task.StatusActive),
				})
			}

			t.Status = task.StatusOpen
			if err = store.Save(t); err != nil {
				printError(err)
			}
			printOutput(formatter.FormatTask(t))
		},
	}
}

// closeCmd implements 'bits close'.
func closeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "close <id> <reason>",
		Short: "Close a task",
		Args:  cobra.ExactArgs(2), //nolint:mnd // CLI takes 2 positional args
		Run: func(_ *cobra.Command, args []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			t, err := store.Load(args[0])
			if err != nil {
				printError(err)
			}

			if t.Status != task.StatusActive {
				printError(InvalidStatusError{
					ID:       t.ID,
					Current:  string(t.Status),
					Expected: string(task.StatusActive),
				})
			}

			reason := args[1]
			if reason == "" {
				printError(MissingReasonError{})
			}

			now := time.Now().UTC()
			t.Status = task.StatusClosed
			t.ClosedAt = &now
			t.CloseReason = &reason

			if err = store.Save(t); err != nil {
				printError(err)
			}
			printOutput(formatter.FormatTask(t))
		},
	}
}

// depCmd implements 'bits dep'.
func depCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dep <id> <depends-on-id>",
		Short: "Add a dependency",
		Args:  cobra.ExactArgs(2), //nolint:mnd // CLI takes 2 positional args
		Run: func(_ *cobra.Command, args []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			taskID := args[0]
			depID := args[1]

			// Load all tasks for cycle detection
			tasks, err := store.List(storage.StatusFilter{})
			if err != nil {
				printError(err)
			}

			graph := deps.NewGraph(tasks)
			if err = graph.ValidateAddDep(taskID, depID); err != nil {
				printError(err)
			}

			t, err := store.Load(taskID)
			if err != nil {
				printError(err)
			}

			// Check if already depends on
			if slices.Contains(t.DependsOn, depID) {
				printOutput(formatter.FormatMessage("Dependency already exists"))
				return
			}

			t.DependsOn = append(t.DependsOn, depID)
			if err = store.Save(t); err != nil {
				printError(err)
			}
			printOutput(formatter.FormatTask(t))
		},
	}
}

// undepCmd implements 'bits undep'.
func undepCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "undep <id> <depends-on-id>",
		Short: "Remove a dependency",
		Args:  cobra.ExactArgs(2), //nolint:mnd // CLI takes 2 positional args
		Run: func(_ *cobra.Command, args []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			t, err := store.Load(args[0])
			if err != nil {
				printError(err)
			}

			depID := args[1]
			originalLen := len(t.DependsOn)
			t.DependsOn = slices.DeleteFunc(t.DependsOn, func(d string) bool {
				return d == depID
			})
			if len(t.DependsOn) == originalLen {
				printOutput(formatter.FormatMessage("Dependency not found"))
				return
			}
			if err = store.Save(t); err != nil {
				printError(err)
			}
			printOutput(formatter.FormatTask(t))
		},
	}
}

// pruneCmd implements 'bits prune'.
func pruneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Remove all closed tasks",
		Run: func(_ *cobra.Command, _ []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			tasks, err := store.List(storage.StatusFilter{Closed: true})
			if err != nil {
				printError(err)
			}

			if len(tasks) == 0 {
				printOutput(formatter.FormatMessage("No closed tasks to prune"))
				return
			}

			for _, t := range tasks {
				if err = store.Delete(t.ID); err != nil {
					printError(err)
				}
			}
			printOutput(formatter.FormatMessage(fmt.Sprintf("Pruned %d closed task(s)", len(tasks))))
		},
	}
}

// rmCmd implements 'bits rm'.
func rmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <id>",
		Short: "Remove a task",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			taskID := args[0]

			// First check task exists
			_, err = store.Load(taskID)
			if err != nil {
				printError(err)
			}

			// Remove from other tasks' dependencies
			if err = store.RemoveDependency(taskID); err != nil {
				printError(err)
			}

			// Delete the task
			if err = store.Delete(taskID); err != nil {
				printError(err)
			}
			printOutput(formatter.FormatMessage(fmt.Sprintf("Removed task %s", taskID)))
		},
	}
}

type hookResponse struct {
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	SystemMessage string `json:"systemMessage"`
}

func formatActiveBlock(t *task.Task) string {
	resp := hookResponse{
		Decision: "block",
		Reason: fmt.Sprintf(
			"Continue working on task %s. Run 'bits show %s' for details. When complete: bits close %s \"reason\".",
			t.ID,
			t.ID,
			t.ID,
		),
		SystemMessage: fmt.Sprintf("Task %s: Still active", t.ID),
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

func formatOpenBlock(count int) string {
	resp := hookResponse{
		Decision: "block",
		Reason: fmt.Sprintf(
			"There are %d open tasks remaining. Use 'bits ready' to see available work.",
			count,
		),
		SystemMessage: fmt.Sprintf("%d open tasks remaining", count),
	}
	b, _ := json.Marshal(resp)
	return string(b)
}
