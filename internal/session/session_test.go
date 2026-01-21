//nolint:testpackage // Tests require internal access for thorough testing
package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a session
	original := &Session{
		SessionID:   "test-session-123",
		Source:      "claude-code",
		DrainActive: false,
	}

	// Save it
	if err := Save(tmpDir, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load it back
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.SessionID != original.SessionID {
		t.Errorf("SessionID = %q, want %q", loaded.SessionID, original.SessionID)
	}
	if loaded.Source != original.Source {
		t.Errorf("Source = %q, want %q", loaded.Source, original.Source)
	}
	if loaded.DrainActive != original.DrainActive {
		t.Errorf("DrainActive = %v, want %v", loaded.DrainActive, original.DrainActive)
	}
}

func TestSessionClaimFirstWriterWins(t *testing.T) {
	tmpDir := t.TempDir()

	// First claim should succeed
	claimed1, owner1, err := Claim(tmpDir, "session-1", "claude-code")
	if err != nil {
		t.Fatalf("First Claim failed: %v", err)
	}
	if !claimed1 {
		t.Error("First Claim should return claimed=true")
	}
	if owner1 != "" {
		t.Errorf("First Claim should return empty owner, got %q", owner1)
	}

	// Second claim should fail
	claimed2, owner2, err := Claim(tmpDir, "session-2", "claude-code")
	if err != nil {
		t.Fatalf("Second Claim failed: %v", err)
	}
	if claimed2 {
		t.Error("Second Claim should return claimed=false")
	}
	if owner2 != "session-1" {
		t.Errorf("Second Claim should return owner=%q, got %q", "session-1", owner2)
	}
}

func TestSessionRelease(t *testing.T) {
	tmpDir := t.TempDir()

	// Claim first
	_, _, err := Claim(tmpDir, "session-1", "claude-code")
	if err != nil {
		t.Fatalf("Claim failed: %v", err)
	}

	// Non-owner can't release
	released, err := Release(tmpDir, "session-2")
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}
	if released {
		t.Error("Non-owner Release should return false")
	}

	// Owner can release
	released, err = Release(tmpDir, "session-1")
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}
	if !released {
		t.Error("Owner Release should return true")
	}

	// Session should be deleted
	if Exists(tmpDir) {
		t.Error("Session file should be deleted after release")
	}
}

func TestSessionDelete(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a session
	s := &Session{SessionID: "test"}
	if err := Save(tmpDir, s); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !Exists(tmpDir) {
		t.Error("Session should exist after Save")
	}

	// Delete it
	if err := Delete(tmpDir); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if Exists(tmpDir) {
		t.Error("Session should not exist after Delete")
	}

	// Deleting non-existent session should not error
	if err := Delete(tmpDir); err != nil {
		t.Errorf("Delete of non-existent session should not error: %v", err)
	}
}

func TestReadStdin(t *testing.T) {
	// Save original stdin
	oldStdin := os.Stdin

	// Create a pipe with valid JSON
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Write test data
	testJSON := `{"session_id": "abc123", "source": "claude-code"}`
	go func() {
		_, _ = w.WriteString(testJSON)
		w.Close()
	}()

	//nolint:reassign // Testing requires reassigning os.Stdin
	os.Stdin = r
	defer func() {
		//nolint:reassign // Testing requires reassigning os.Stdin
		os.Stdin = oldStdin
	}()

	input, err := ReadStdin()
	if err != nil {
		t.Fatalf("ReadStdin failed: %v", err)
	}

	if input.SessionID != "abc123" {
		t.Errorf("SessionID = %q, want %q", input.SessionID, "abc123")
	}
	if input.Source != "claude-code" {
		t.Errorf("Source = %q, want %q", input.Source, "claude-code")
	}
}

func TestReadStdinInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty input", ""},
		{"invalid JSON", "not json"},
		{"missing session_id", `{"source": "test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdin
			oldStdin := os.Stdin
			defer func() {
				//nolint:reassign // Testing requires reassigning os.Stdin
				os.Stdin = oldStdin
			}()

			// Create a pipe with test input
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}

			go func() {
				_, _ = w.WriteString(tt.input)
				w.Close()
			}()

			//nolint:reassign // Testing requires reassigning os.Stdin
			os.Stdin = r

			_, err = ReadStdin()
			if err == nil {
				t.Error("ReadStdin should return error for invalid input")
			}
		})
	}
}

func TestSetDrainActive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a session
	_, _, err := Claim(tmpDir, "session-1", "claude-code")
	if err != nil {
		t.Fatalf("Claim failed: %v", err)
	}

	// Non-owner can't set drain
	success, err := SetDrainActive(tmpDir, "session-2", true)
	if err != nil {
		t.Fatalf("SetDrainActive failed: %v", err)
	}
	if success {
		t.Error("Non-owner SetDrainActive should return false")
	}

	// Owner can set drain active
	success, err = SetDrainActive(tmpDir, "session-1", true)
	if err != nil {
		t.Fatalf("SetDrainActive failed: %v", err)
	}
	if !success {
		t.Error("Owner SetDrainActive should return true")
	}

	// Verify drain is active
	s, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !s.DrainActive {
		t.Error("DrainActive should be true")
	}
	if s.DrainStartedAt == nil {
		t.Error("DrainStartedAt should be set")
	}

	// Owner can deactivate drain
	success, err = SetDrainActive(tmpDir, "session-1", false)
	if err != nil {
		t.Fatalf("SetDrainActive failed: %v", err)
	}
	if !success {
		t.Error("Owner SetDrainActive(false) should return true")
	}

	// Verify drain is inactive
	s, err = Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if s.DrainActive {
		t.Error("DrainActive should be false")
	}
	if s.DrainStartedAt != nil {
		t.Error("DrainStartedAt should be nil")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Should not exist initially
	if Exists(tmpDir) {
		t.Error("Session should not exist initially")
	}

	// Create session file
	sessionPath := filepath.Join(tmpDir, "session.json")
	if err := os.WriteFile(sessionPath, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Should exist now
	if !Exists(tmpDir) {
		t.Error("Session should exist after creation")
	}
}

func TestLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Load(tmpDir)
	if err == nil {
		t.Error("Load should return error for non-existent session")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Load should return os.IsNotExist error, got: %v", err)
	}
}
