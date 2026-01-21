package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/abatilo/bits/internal/session"
	"github.com/abatilo/bits/internal/storage"
)

// sessionCmd implements 'bits session' command group.
func sessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Session management for Claude Code integration",
	}

	cmd.AddCommand(
		sessionClaimCmd(),
		sessionReleaseCmd(),
		sessionPruneCmd(),
		sessionHookCmd(),
	)

	return cmd
}

type claimResponse struct {
	Claimed bool   `json:"claimed"`
	Owner   string `json:"owner,omitempty"`
}

// sessionClaimCmd implements 'bits session claim'.
func sessionClaimCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claim",
		Short: "Claim primary session (reads session_id from stdin)",
		Run: func(_ *cobra.Command, _ []string) {
			input, err := session.ReadStdin()
			if err != nil {
				return // No stdin - just exit 0
			}

			store, err := getStore()
			if err != nil {
				printError(err)
			}

			claimed, owner, err := session.Claim(store.BasePath(), input.SessionID, input.Source)
			if err != nil {
				printError(err)
			}

			resp := claimResponse{
				Claimed: claimed,
				Owner:   owner,
			}
			data, _ := json.Marshal(resp)
			printOutput(string(data) + "\n")
		},
	}
}

type releaseResponse struct {
	Released bool `json:"released"`
}

// sessionReleaseCmd implements 'bits session release'.
func sessionReleaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "release",
		Short: "Release session (reads session_id from stdin)",
		Run: func(_ *cobra.Command, _ []string) {
			input, err := session.ReadStdin()
			if err != nil {
				return // No stdin - just exit 0
			}

			store, err := getStore()
			if err != nil {
				os.Exit(0) // Allow graceful exit
			}

			released, err := session.Release(store.BasePath(), input.SessionID)
			if err != nil {
				os.Exit(0) // Allow graceful exit
			}

			resp := releaseResponse{Released: released}
			data, _ := json.Marshal(resp)
			printOutput(string(data) + "\n")
		},
	}
}

// sessionPruneCmd implements 'bits session prune'.
func sessionPruneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Remove stale session file (manual cleanup)",
		Run: func(_ *cobra.Command, _ []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			if !session.Exists(store.BasePath()) {
				printOutput(formatter.FormatMessage("No session file to prune"))
				return
			}

			if deleteErr := session.Delete(store.BasePath()); deleteErr != nil {
				printError(deleteErr)
			}

			printOutput(formatter.FormatMessage("Session file pruned"))
		},
	}
}

// sessionHookCmd implements 'bits session hook' for stop hook integration.
func sessionHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Stop hook with session ownership check",
		Run: func(_ *cobra.Command, _ []string) {
			input, err := session.ReadStdin()
			if err != nil {
				// No session input - allow stop
				os.Exit(0)
			}

			store, err := getStore()
			if err != nil {
				os.Exit(0) // Allow stop on error
			}

			// Check if session file exists
			if !session.Exists(store.BasePath()) {
				os.Exit(0) // No session - allow stop
			}

			// Load session
			sess, err := session.Load(store.BasePath())
			if err != nil {
				os.Exit(0) // Allow stop on error
			}

			// Check if this is the primary session
			if sess.SessionID != input.SessionID {
				os.Exit(0) // Not primary - allow stop
			}

			// Check if drain mode is active
			if !sess.DrainActive {
				os.Exit(0) // Not draining - allow stop
			}

			// Drain mode is active for primary session - check for remaining tasks
			// Check for active tasks first
			activeTasks, err := store.List(storage.StatusFilter{Active: true})
			if err == nil && len(activeTasks) > 0 {
				_, _ = os.Stdout.WriteString(formatActiveBlock(activeTasks[0]) + "\n")
				return
			}

			// Check for open tasks
			openTasks, err := store.List(storage.StatusFilter{Open: true})
			if err == nil && len(openTasks) > 0 {
				_, _ = os.Stdout.WriteString(formatOpenBlock(len(openTasks)) + "\n")
				return
			}

			// All tasks complete - allow stop (drain complete)
			os.Exit(0)
		},
	}
}
