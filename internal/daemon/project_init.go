package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectInit initializes a project's .starclaw directory with default
// configuration and instruction files for the agent.
type ProjectInit struct{}

// InitProject creates the .starclaw directory in the given project dir
// with default config.local.yaml and instructions.md files.
// Existing files are not overwritten.
func (pi *ProjectInit) InitProject(dir string) error {
	starclawDir := filepath.Join(dir, ".starclaw")

	if err := os.MkdirAll(starclawDir, 0700); err != nil {
		return fmt.Errorf("create .starclaw directory: %w", err)
	}

	// config.local.yaml — project-level overrides on top of global config.
	configPath := filepath.Join(starclawDir, "config.local.yaml")
	if err := writeIfNotExists(configPath, defaultLocalConfig); err != nil {
		return fmt.Errorf("write config.local.yaml: %w", err)
	}

	// instructions.md — project-specific instructions for the agent.
	instructionsPath := filepath.Join(starclawDir, "instructions.md")
	if err := writeIfNotExists(instructionsPath, defaultInstructions); err != nil {
		return fmt.Errorf("write instructions.md: %w", err)
	}

	return nil
}

const defaultLocalConfig = `# Project-level StarClaw configuration overrides.
# These values are merged on top of ~/.starclaw/config.yaml.

# Example overrides (uncomment to use):
# agent:
#   max_iterations: 50
#   temperature: 0.2
# tools:
#   bash_timeout: 300
#   allowed:
#     - "file_read"
#     - "bash"
#     - "glob"
`

const defaultInstructions = `# Project Instructions for the Agent

Add project-specific instructions, guidelines, and notes here.
`

// writeIfNotExists writes content to path only if the file does not already
// exist. Returns nil if the file already exists.
func writeIfNotExists(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists, do not overwrite
	}
	return os.WriteFile(path, []byte(content), 0600)
}
