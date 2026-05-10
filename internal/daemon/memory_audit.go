package daemon

import (
	"os"
	"path/filepath"
	"strings"
)

const consolidationThreshold = 5

// AuditReport contains the results of a memory directory audit.
type AuditReport struct {
	// TotalEntries is the total number of memory files found.
	TotalEntries int
	// TotalSize is the total size in bytes of all memory files.
	TotalSize int64
	// AutoFiles lists the paths of auto-generated memory files (auto-*.md).
	AutoFiles []string
	// NeedsConsolidation is true when auto-generated files exceed the threshold.
	NeedsConsolidation bool
}

// MemoryAudit audits agent memory usage and provides fallback checks.
type MemoryAudit struct{}

// NewMemoryAudit creates a new MemoryAudit.
func NewMemoryAudit() *MemoryAudit {
	return &MemoryAudit{}
}

// Audit scans memory files in the given directory and returns an AuditReport.
// Returns an empty report if the directory does not exist or is inaccessible.
func (ma *MemoryAudit) Audit(memoryDir string) AuditReport {
	var report AuditReport

	entries, err := filepath.Glob(filepath.Join(memoryDir, "*.md"))
	if err != nil {
		return report
	}

	for _, entry := range entries {
		info, err := os.Stat(entry)
		if err != nil {
			continue
		}
		report.TotalEntries++
		report.TotalSize += info.Size()
		if strings.HasPrefix(filepath.Base(entry), "auto-") {
			report.AutoFiles = append(report.AutoFiles, entry)
		}
	}

	if len(report.AutoFiles) > consolidationThreshold {
		report.NeedsConsolidation = true
	}

	return report
}

// CheckFallback checks if fallback memory (auto-*.md files) exists and needs
// consolidation. Returns true when there are more auto-generated files than
// the consolidation threshold.
func (ma *MemoryAudit) CheckFallback(memoryDir string) bool {
	if memoryDir == "" {
		return false
	}
	entries, err := filepath.Glob(filepath.Join(memoryDir, "auto-*.md"))
	if err != nil {
		return false
	}
	return len(entries) > consolidationThreshold
}
