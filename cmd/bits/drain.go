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
		Short: "Activate drain mode (reads session_id from stdin)",
		Run: func(_ *cobra.Command, _ []string) {
			input, err := session.ReadStdin()
			if err != nil {
				resp := drainResponse{
					Success: false,
					Message: "No session input provided",
				}
				data, _ := json.Marshal(resp)
				printOutput(string(data) + "\n")
				os.Exit(1)
			}

			store, err := getStore()
			if err != nil {
				printError(err)
			}

			// Check if session exists
			if !session.Exists(store.BasePath()) {
				resp := drainResponse{
					Success: false,
					Message: "No session file exists. Run 'bits session claim' first.",
				}
				data, _ := json.Marshal(resp)
				printOutput(string(data) + "\n")
				os.Exit(1)
			}

			success, err := session.SetDrainActive(store.BasePath(), input.SessionID, true)
			if err != nil {
				printError(err)
			}

			if !success {
				resp := drainResponse{
					Success:     false,
					DrainActive: false,
					Message:     "Not session owner - cannot activate drain mode",
				}
				data, _ := json.Marshal(resp)
				printOutput(string(data) + "\n")
				os.Exit(1)
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
		Short: "Deactivate drain mode (reads session_id from stdin)",
		Run: func(_ *cobra.Command, _ []string) {
			input, err := session.ReadStdin()
			if err != nil {
				resp := drainResponse{
					Success: false,
					Message: "No session input provided",
				}
				data, _ := json.Marshal(resp)
				printOutput(string(data) + "\n")
				os.Exit(1)
			}

			store, err := getStore()
			if err != nil {
				printError(err)
			}

			// Check if session exists
			if !session.Exists(store.BasePath()) {
				resp := drainResponse{
					Success:     false,
					DrainActive: false,
					Message:     "No session file exists",
				}
				data, _ := json.Marshal(resp)
				printOutput(string(data) + "\n")
				return
			}

			success, err := session.SetDrainActive(store.BasePath(), input.SessionID, false)
			if err != nil {
				printError(err)
			}

			if !success {
				resp := drainResponse{
					Success:     false,
					DrainActive: false,
					Message:     "Not session owner - cannot deactivate drain mode",
				}
				data, _ := json.Marshal(resp)
				printOutput(string(data) + "\n")
				os.Exit(1)
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
