package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandHome expands ~ to home directory
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return path
}

// IsSafePath checks if a path is safe to access
// It prevents path traversal attacks and access to sensitive directories
func IsSafePath(path string) error {
	// Expand home directory
	path = ExpandHome(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Clean the path
	cleanPath := filepath.Clean(absPath)

	// Check for path traversal after cleaning
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains traversal: %s", path)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	// Check if path is under CWD or home directory
	home, _ := os.UserHomeDir()

	// Allow paths under CWD
	if strings.HasPrefix(cleanPath, cwd) {
		return nil
	}

	// Allow paths under home directory
	if home != "" && strings.HasPrefix(cleanPath, home) {
		return nil
	}

	// Block access to system directories
	sensitivePaths := []string{
		"/etc",
		"/usr",
		"/bin",
		"/sbin",
		"/lib",
		"/lib64",
		"/opt",
		"/sys",
		"/proc",
		"/dev",
		"/boot",
	}

	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(cleanPath, sensitive) {
			return fmt.Errorf("access to system directory denied: %s", cleanPath)
		}
	}

	return nil
}

// IsPathUnderCWD checks if a path is under current working directory
func IsPathUnderCWD(path string) bool {
	path = ExpandHome(path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	return strings.HasPrefix(absPath, cwd)
}

// NormalizePath converts path to absolute and clean
func NormalizePath(path string) (string, error) {
	path = ExpandHome(path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absPath), nil
}
