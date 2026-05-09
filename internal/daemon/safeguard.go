package daemon

import (
	"os"
	"path/filepath"
	"strings"
)

// Safeguard provides safety checks for dangerous file system operations
// and shell commands.  It is configured with a list of allowed or denied
// path prefixes and dangerous command patterns.
type Safeguard struct {
	starclawDir string
}

// NewSafeguard creates a new Safeguard scoped to starclawDir.
func NewSafeguard(starclawDir string) *Safeguard {
	return &Safeguard{starclawDir: starclawDir}
}

// commandCheck describes a dangerous command pattern and how to detect it.
type commandCheck struct {
	reason  string
	matcher func(cmd string) bool
}

// blockedCommands lists shell command checks that are considered dangerous.
var blockedCommands = []commandCheck{
	{
		reason: "recursive root deletion",
		matcher: func(cmd string) bool {
			// Match rm -rf where one of the path arguments is exactly "/".
			parts := strings.Fields(cmd)
			for i := 0; i < len(parts); i++ {
				if parts[i] == "rm" && i+1 < len(parts) && parts[i+1] == "-rf" {
					for j := i + 2; j < len(parts); j++ {
						arg := parts[j]
						// Skip flags (e.g. --no-preserve-root)
						if strings.HasPrefix(arg, "--") {
							continue
						}
						// Check if the path resolves to root.
						abs, err := filepath.Abs(arg)
						if err == nil && abs == "/" {
							return true
						}
					}
				}
			}
			return false
		},
	},
	{
		reason: "filesystem creation",
		matcher: func(cmd string) bool {
			parts := strings.Fields(cmd)
			for _, p := range parts {
				if strings.HasPrefix(p, "mkfs") {
					return true
				}
			}
			return false
		},
	},
	{
		reason: "raw disk write",
		matcher: func(cmd string) bool {
			lower := strings.ToLower(cmd)
			return strings.HasPrefix(lower, "dd if=")
		},
	},
	{
		reason: "fork bomb",
		matcher: func(cmd string) bool {
			normalised := strings.ReplaceAll(cmd, " ", "")
			if normalised == ":(){:|:&};:" {
				return true
			}
			// Also catch variants.
			if strings.Contains(normalised, ":()") && strings.Contains(normalised, "{|:") {
				return true
			}
			return false
		},
	},
	{
		reason: "permission removal",
		matcher: func(cmd string) bool {
			// Match chmod -R 000 (with any path).
			parts := strings.Fields(cmd)
			for i := 0; i < len(parts); i++ {
				if parts[i] == "chmod" && i+1 < len(parts) && parts[i+1] == "-r" && i+2 < len(parts) && parts[i+2] == "000" {
					return true
				}
				if parts[i] == "chmod" && i+1 < len(parts) && strings.EqualFold(parts[i+1], "-R") && i+2 < len(parts) && parts[i+2] == "000" {
					return true
				}
			}
			return false
		},
	},
	{
		reason: "ownership change",
		matcher: func(cmd string) bool {
			parts := strings.Fields(cmd)
			for i := 0; i < len(parts); i++ {
				if parts[i] == "chown" && i+1 < len(parts) {
					arg := parts[i+1]
					// -R flag (possibly combined with other flags).
					if arg == "-r" || strings.EqualFold(arg, "-R") || strings.HasPrefix(strings.ToLower(arg), "-r") {
						return true
					}
				}
			}
			return false
		},
	},
}

// CheckCommand inspects a shell command and returns whether it should
// be blocked along with a human-readable reason.  The check is
// case-insensitive and matches against known dangerous patterns.
func (s *Safeguard) CheckCommand(cmd string) (blocked bool, reason string) {
	normalised := strings.TrimSpace(cmd)
	normalised = strings.ToLower(normalised)

	for _, bc := range blockedCommands {
		if bc.matcher(normalised) {
			return true, bc.reason
		}
	}
	return false, ""
}

// CheckPath verifies that the given path is safe to read or write by
// ensuring it is within the starclaw directory tree.  Returns true if
// the path is safe.
func (s *Safeguard) CheckPath(path string) bool {
	if s.starclawDir == "" {
		return false
	}
	if path == "" {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absStarclaw, err := filepath.Abs(s.starclawDir)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(absStarclaw, absPath)
	if err != nil {
		return false
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}

	// Non-existent paths are considered safe as long as they are
	// logically within starclawDir.
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return true
	}

	return true
}
