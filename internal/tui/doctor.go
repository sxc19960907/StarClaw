package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/starclaw/starclaw/internal/config"
)

// CheckStatus represents the status of a health check.
type CheckStatus int

const (
	CheckPass CheckStatus = iota
	CheckFail
	CheckWarn
)

// CheckResult holds the outcome of a single health check.
type CheckResult struct {
	Name    string
	Status  CheckStatus
	Message string
}

// Doctor runs health check diagnostics.
type Doctor struct{}

// NewDoctor creates a new Doctor.
func NewDoctor() *Doctor {
	return &Doctor{}
}

// RunChecks runs all health checks and returns the results.
func (d *Doctor) RunChecks() []CheckResult {
	return []CheckResult{
		d.checkGoVersion(),
		d.checkConfig(),
		d.checkStarclawDir(),
		d.checkToolBinaries(),
	}
}

func (d *Doctor) checkGoVersion() CheckResult {
	output, err := exec.Command("go", "version").Output()
	if err != nil {
		return CheckResult{
			Name:    "go_version",
			Status:  CheckFail,
			Message: "Go is not installed or not in PATH",
		}
	}
	version := strings.TrimSpace(string(output))
	// Extract the semver from "go version go1.xx.x darwin/amd64"
	parts := strings.Fields(version)
	if len(parts) >= 3 {
		ver := strings.TrimPrefix(parts[2], "go")
		if compareGoVersion(ver, "1.18") < 0 {
			return CheckResult{
				Name:    "go_version",
				Status:  CheckWarn,
				Message: version + " (minimum recommended: 1.18)",
			}
		}
	}
	return CheckResult{
		Name:    "go_version",
		Status:  CheckPass,
		Message: version,
	}
}

func (d *Doctor) checkConfig() CheckResult {
	cfg, err := config.Load()
	if err != nil {
		return CheckResult{
			Name:    "config",
			Status:  CheckFail,
			Message: "Failed to load config: " + err.Error(),
		}
	}
	if config.NeedsSetup(cfg) {
		return CheckResult{
			Name:    "config",
			Status:  CheckWarn,
			Message: "Configuration is incomplete (API key missing)",
		}
	}
	return CheckResult{
		Name:    "config",
		Status:  CheckPass,
		Message: "Configuration is valid",
	}
}

func (d *Doctor) checkStarclawDir() CheckResult {
	dir := config.StarclawDir()
	if dir == "" {
		return CheckResult{
			Name:    "starclaw_dir",
			Status:  CheckFail,
			Message: "Home directory not resolvable",
		}
	}
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return CheckResult{
				Name:    "starclaw_dir",
				Status:  CheckWarn,
				Message: dir + " does not exist (will be created on first run)",
			}
		}
		return CheckResult{
			Name:    "starclaw_dir",
			Status:  CheckFail,
			Message: "Cannot stat: " + err.Error(),
		}
	}
	if !info.IsDir() {
		return CheckResult{
			Name:    "starclaw_dir",
			Status:  CheckFail,
			Message: dir + " exists but is not a directory",
		}
	}
	return CheckResult{
		Name:    "starclaw_dir",
		Status:  CheckPass,
		Message: dir + " exists",
	}
}

func (d *Doctor) checkToolBinaries() CheckResult {
	tools := []string{"git", "go"}
	if runtime.GOOS == "darwin" {
		tools = append(tools, "sw_vers")
	}
	var missing []string
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}
	if len(missing) > 0 {
		return CheckResult{
			Name:    "tool_binaries",
			Status:  CheckWarn,
			Message: "Missing: " + strings.Join(missing, ", "),
		}
	}
	return CheckResult{
		Name:    "tool_binaries",
		Status:  CheckPass,
		Message: "All required tools found",
	}
}

// compareGoVersion compares two Go semver strings a and b.
// Trailing segments are treated as zero, so "1.18" equals "1.18.0".
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareGoVersion(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}
	for i := 0; i < maxLen; i++ {
		var ai, bi int
		if i < len(aParts) {
			fmt.Sscanf(aParts[i], "%d", &ai)
		}
		if i < len(bParts) {
			fmt.Sscanf(bParts[i], "%d", &bi)
		}
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
	}
	return 0
}
