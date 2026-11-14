package domain

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

type BlockHash string

func ParseBlockHash(raw string) (BlockHash, error) {
	clean := strings.ToLower(strings.TrimSpace(raw))
	if len(clean) != 64 {
		return "", fmt.Errorf("invalid blake3 hash length: got %d, want 64", len(clean))
	}
	if _, err := hex.DecodeString(clean); err != nil {
		return "", fmt.Errorf("invalid hex in blake3 hash: %w", err)
	}
	return BlockHash(clean), nil
}

// DiskPath returns the 2-character prefix sharded path (e.g., "<baseDir>/57/57a49b...").
func (h BlockHash) DiskPath(baseDir string) string {
	s := string(h)
	return filepath.Join(baseDir, s[:2], s)
}

func (h BlockHash) String() string {
	return string(h)
}
