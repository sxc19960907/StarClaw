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

// IsSafePath checks if a path is safe to access.
// It prevents path traversal attacks and access to sensitive directories.
func IsSafePath(path string) error {
	// Expand home directory
	path = ExpandHome(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Clean the path (before symlink resolution)
	cleanPath := filepath.Clean(absPath)

	// Check for path traversal after cleaning
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains traversal: %s", path)
	}

	// Block access to system directories — check before symlink resolution
	// because on macOS /etc is a symlink to /private/etc.
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
		if strings.HasPrefix(cleanPath, sensitive+string(filepath.Separator)) || cleanPath == sensitive {
			return fmt.Errorf("access to system directory denied: %s", cleanPath)
		}
	}

	// Resolve symlinks to get the real path for CWD/home checks.
	// This prevents symlink attacks where a file inside the project
	// points to a sensitive location.
	resolvedPath := cleanPath
	if rp, err := filepath.EvalSymlinks(cleanPath); err == nil {
		resolvedPath = filepath.Clean(rp)
	}

	// Check sensitive paths again on the resolved path
	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(resolvedPath, sensitive+string(filepath.Separator)) || resolvedPath == sensitive {
			return fmt.Errorf("access to system directory denied: %s", cleanPath)
		}
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	// Allow paths under CWD (check with separator to avoid prefix collisions)
	if strings.HasPrefix(resolvedPath, cwd+string(filepath.Separator)) || resolvedPath == cwd {
		return nil
	}

	// Allow paths under home directory
	home, _ := os.UserHomeDir()
	if home != "" && (strings.HasPrefix(resolvedPath, home+string(filepath.Separator)) || resolvedPath == home) {
		return nil
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

	return strings.HasPrefix(absPath, cwd+string(filepath.Separator)) || absPath == cwd
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
