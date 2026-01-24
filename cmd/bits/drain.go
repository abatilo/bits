package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/abatilo/bits/internal/session"
	"github.com/abatilo/bits/internal/storage"
)

// drainCmd implements 'bits drain' command group.
func drainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drain",
		Short: "Drain mode management",
	}

	cmd.AddCommand(
		drainClaimCmd(),
		drainReleaseCmd(),
	)

	return cmd
}

type drainResponse struct {
	Success     bool   `json:"success"`
	DrainActive bool   `json:"drain_active"`
	Message     string `json:"message,omitempty"`
}

// drainClaimCmd implements 'bits drain claim'.
func drainClaimCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claim",
		Short: "Activate drain mode",
		Run: func(_ *cobra.Command, _ []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			// Load existing session to get session_id
			sess, err := session.Load(store.BasePath())
			if err != nil {
				resp := drainResponse{
					Success: false,
					Message: "No session file exists. Run 'bits session claim' first.",
				}
				data, _ := json.Marshal(resp)
				printOutput(string(data) + "\n")
				os.Exit(1)
			}

			// Use session's own ID - no external verification needed
			_, err = session.SetDrainActive(store.BasePath(), sess.SessionID, true)
			if err != nil {
				printError(err)
			}

			resp := drainResponse{
				Success:     true,
				DrainActive: true,
				Message:     "Drain mode activated",
			}
			data, _ := json.Marshal(resp)
			printOutput(string(data) + "\n")
		},
	}
}

// drainReleaseCmd implements 'bits drain release'.
func drainReleaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "release",
		Short: "Deactivate drain mode",
		Run: func(_ *cobra.Command, _ []string) {
			store, err := getStore()
			if err != nil {
				printError(err)
			}

			// Load existing session to get session_id
			sess, err := session.Load(store.BasePath())
			if err != nil {
				resp := drainResponse{
					Success:     false,
					DrainActive: false,
					Message:     "No session file exists",
				}
				data, _ := json.Marshal(resp)
				printOutput(string(data) + "\n")
				return
			}

			// Check for remaining tasks before allowing release
			activeTasks, err := store.List(storage.StatusFilter{Active: true})
			if err != nil {
				printError(err)
			}
			openTasks, err := store.List(storage.StatusFilter{Open: true})
			if err != nil {
				printError(err)
			}

			if len(activeTasks) > 0 || len(openTasks) > 0 {
				msg := fmt.Sprintf(
					`Claude, you attempted to release drain mode but there are still %d active and %d open tasks remaining.

It looks like you may have forgotten about pending work or misunderstood when drain mode should end.

Please:
1. Re-read your instructions for the current workflow
2. Run 'bits ready' to see what tasks are available
3. Continue working until all tasks are complete

Drain mode should only be released when ALL tasks are finished, not when you want to pause or ask the user a question.`,
					len(activeTasks),
					len(openTasks),
				)

				resp := drainResponse{
					Success:     false,
					DrainActive: true,
					Message:     msg,
				}
				data, _ := json.Marshal(resp)
				printOutput(string(data) + "\n")
				os.Exit(1)
			}

			// Use session's own ID - no external verification needed
			_, err = session.SetDrainActive(store.BasePath(), sess.SessionID, false)
			if err != nil {
				printError(err)
			}

			resp := drainResponse{
				Success:     true,
				DrainActive: false,
				Message:     "Drain mode deactivated",
			}
			data, _ := json.Marshal(resp)
			printOutput(string(data) + "\n")
		},
	}
}
