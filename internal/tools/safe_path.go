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

	resolvedPath, err := resolvePath(cleanPath)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// Check sensitive paths again on the resolved path
	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(resolvedPath, sensitive+string(filepath.Separator)) || resolvedPath == sensitive {
			return fmt.Errorf("access to system directory denied: %s", cleanPath)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}
	resolvedCWD, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		resolvedCWD = filepath.Clean(cwd)
	}

	if isSubpath(resolvedCWD, resolvedPath) {
		return nil
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		resolvedHome, err := filepath.EvalSymlinks(home)
		if err != nil {
			resolvedHome = filepath.Clean(home)
		}
		if isSubpath(resolvedHome, resolvedPath) {
			return nil
		}
	}

	return fmt.Errorf("path outside allowed directories: %s", cleanPath)
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

	resolvedPath, err := resolvePath(filepath.Clean(absPath))
	if err != nil {
		return false
	}
	resolvedCWD, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		resolvedCWD = filepath.Clean(cwd)
	}

	return isSubpath(resolvedCWD, resolvedPath)
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

func resolvePath(path string) (string, error) {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return filepath.Clean(resolved), nil
	}

	dir := path
	var missing []string
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no existing ancestor for %s", path)
		}
		missing = append(missing, filepath.Base(dir))
		dir = parent
		resolvedParent, err := filepath.EvalSymlinks(dir)
		if err == nil {
			resolved := filepath.Clean(resolvedParent)
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return filepath.Clean(resolved), nil
		}
	}
}

func isSubpath(base, target string) bool {
	base = filepath.Clean(base)
	target = filepath.Clean(target)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
