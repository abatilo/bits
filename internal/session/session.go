package session

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

const sessionFile = "session.json"

// Session represents an active Claude Code session for a project.
type Session struct {
	SessionID      string     `json:"session_id"`
	StartedAt      time.Time  `json:"started_at"`
	Source         string     `json:"source"`
	DrainActive    bool       `json:"drain_active"`
	DrainStartedAt *time.Time `json:"drain_started_at,omitempty"`
}

// StdinInput represents the JSON input from Claude Code hooks.
type StdinInput struct {
	SessionID string `json:"session_id"`
	Source    string `json:"source"`
}

// ReadStdin parses Claude Code hook JSON from stdin.
func ReadStdin() (*StdinInput, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("no input from stdin")
	}

	var input StdinInput
	if unmarshalErr := json.Unmarshal(data, &input); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	if input.SessionID == "" {
		return nil, errors.New("session_id is required")
	}

	return &input, nil
}

// sessionPath returns the full path to session.json for the given base path.
func sessionPath(basePath string) string {
	return filepath.Join(basePath, sessionFile)
}

// Exists checks if a session file exists.
func Exists(basePath string) bool {
	_, err := os.Stat(sessionPath(basePath))
	return err == nil
}

// Load reads the session from disk.
func Load(basePath string) (*Session, error) {
	data, err := os.ReadFile(sessionPath(basePath))
	if err != nil {
		return nil, err
	}

	var s Session
	if unmarshalErr := json.Unmarshal(data, &s); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &s, nil
}

// Save writes the session to disk.
func Save(basePath string, s *Session) error {
	//nolint:gosec // G301: 0755 is appropriate for user-accessible session directory
	if mkdirErr := os.MkdirAll(basePath, 0o755); mkdirErr != nil {
		return mkdirErr
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	//nolint:gosec // G306: 0644 is appropriate for user-readable session files
	return os.WriteFile(sessionPath(basePath), data, 0o644)
}

// Delete removes the session file.
func Delete(basePath string) error {
	err := os.Remove(sessionPath(basePath))
	if os.IsNotExist(err) {
		return nil // Already deleted, not an error
	}
	return err
}

// Claim attempts to claim a session. Returns (claimed, existingOwner, error).
// If another session already owns this project, returns (false, ownerID, nil).
// If we successfully claim, returns (true, "", nil).
func Claim(basePath string, sessionID, source string) (bool, string, error) {
	// Check if session already exists
	existing, loadErr := Load(basePath)
	if loadErr == nil {
		// Session exists, not ours to claim
		return false, existing.SessionID, nil
	}

	if !os.IsNotExist(loadErr) {
		// Some other error reading the file
		return false, "", loadErr
	}

	// No session exists, create one
	now := time.Now().UTC()
	s := &Session{
		SessionID:   sessionID,
		StartedAt:   now,
		Source:      source,
		DrainActive: false,
	}

	if saveErr := Save(basePath, s); saveErr != nil {
		return false, "", saveErr
	}

	return true, "", nil
}

// Release removes the session if the given sessionID is the owner.
// Returns true if released, false if not the owner.
func Release(basePath string, sessionID string) (bool, error) {
	existing, loadErr := Load(basePath)
	if os.IsNotExist(loadErr) {
		// No session to release
		return false, nil
	}
	if loadErr != nil {
		return false, loadErr
	}

	if existing.SessionID != sessionID {
		// Not the owner, can't release
		return false, nil
	}

	if deleteErr := Delete(basePath); deleteErr != nil {
		return false, deleteErr
	}

	return true, nil
}

// SetDrainActive updates the drain_active flag. Only the session owner can do this.
// Returns (success, error).
func SetDrainActive(basePath string, sessionID string, active bool) (bool, error) {
	existing, loadErr := Load(basePath)
	if loadErr != nil {
		return false, loadErr
	}

	if existing.SessionID != sessionID {
		// Not the owner
		return false, nil
	}

	existing.DrainActive = active
	if active {
		now := time.Now().UTC()
		existing.DrainStartedAt = &now
	} else {
		existing.DrainStartedAt = nil
	}

	if saveErr := Save(basePath, existing); saveErr != nil {
		return false, saveErr
	}

	return true, nil
}
