package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/abatilo/bits/internal/session"
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
