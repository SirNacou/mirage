package domain

import (
	"errors"
	"path"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type SafeRelPath string

func ParseSafeRelPath(raw string) (SafeRelPath, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/")
	normalized = norm.NFC.String(normalized)

	if normalized == "" || normalized == "." || normalized == "/" {
		return "", errors.New("relative path cannot be empty")
	}
	if strings.HasPrefix(normalized, "/") {
		return "", errors.New("relative path must not start with '/'")
	}

	cleaned := path.Clean(normalized)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", errors.New("path traversal is not allowed")
	}

	return SafeRelPath(cleaned), nil
}

func (p SafeRelPath) String() string {
	return string(p)
}
