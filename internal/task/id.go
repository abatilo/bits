package task

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const (
	minIDLength  = 3
	maxIDLength  = 8
	nonceSize    = 16 // 128 bits of entropy
	hexChunkSize = 4  // Process 4 hex chars (16 bits) at a time for base36 conversion
)

// GenerateID creates a unique task ID using hash-based generation with adaptive length.
// It starts with minIDLength characters and grows up to maxIDLength to avoid collisions.
func GenerateID(title string, createdAt time.Time, existsFn func(string) bool) string {
	// Generate random nonce
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}

	// Create hash from title + timestamp + nonce
	h := sha256.New()
	h.Write([]byte(title))
	h.Write([]byte(createdAt.Format(time.RFC3339Nano)))
	h.Write(nonce)
	hash := h.Sum(nil)

	// Convert to base36
	base36 := hexToBase36(hex.EncodeToString(hash))

	// Try progressively longer prefixes until we find a unique one
	for length := minIDLength; length <= maxIDLength; length++ {
		if length > len(base36) {
			break
		}
		candidate := base36[:length]
		if !existsFn(candidate) {
			return candidate
		}
	}

	// Fallback: use full hash (extremely unlikely to reach here)
	return base36[:maxIDLength]
}

// hexToBase36 converts a hex string to base36.
func hexToBase36(hexStr string) string {
	var result strings.Builder
	// Process hex chars in chunks
	for i := 0; i < len(hexStr); i += hexChunkSize {
		end := min(i+hexChunkSize, len(hexStr))
		chunk := hexStr[i:end]
		val, _ := strconv.ParseUint(chunk, 16, 64)
		result.WriteString(strconv.FormatUint(val, 36))
	}
	return result.String()
}
