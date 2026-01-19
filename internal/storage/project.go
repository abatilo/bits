package storage

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	bitserrors "github.com/abatilo/bits/internal/errors"
)

var nonAlphanumericRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// FindProjectRoot walks up from cwd looking for .git directory.
// Returns the directory containing .git, or error if not found.
func FindProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		var info os.FileInfo
		info, err = os.Stat(gitPath)
		if err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding .git
			return "", bitserrors.NotInRepoError{}
		}
		dir = parent
	}
}

// SanitizePath converts an absolute path to a safe directory name.
// "/Users/abatilo/myproject" -> "Users-abatilo-myproject".
func SanitizePath(path string) string {
	// Remove leading slash
	result := strings.TrimPrefix(path, "/")

	// Replace non-alphanumeric chars with dash
	result = nonAlphanumericRe.ReplaceAllString(result, "-")

	// Trim leading/trailing dashes
	result = strings.Trim(result, "-")

	return result
}
