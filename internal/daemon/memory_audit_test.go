package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMemoryAuditEmptyDir verifies that an empty directory produces an empty report.
func TestMemoryAuditEmptyDir(t *testing.T) {
	dir := t.TempDir()
	ma := NewMemoryAudit()

	report := ma.Audit(dir)
	if report.TotalEntries != 0 {
		t.Errorf("TotalEntries = %d, want 0", report.TotalEntries)
	}
	if report.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0", report.TotalSize)
	}
	if len(report.AutoFiles) != 0 {
		t.Errorf("AutoFiles = %d, want 0", len(report.AutoFiles))
	}
	if report.NeedsConsolidation {
		t.Errorf("NeedsConsolidation should be false for empty dir")
	}
}

// TestMemoryAuditRegularFiles verifies that regular .md files are counted.
func TestMemoryAuditRegularFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "memory.md"), "hello world")
	writeFile(t, filepath.Join(dir, "notes.md"), "some notes")

	ma := NewMemoryAudit()
	report := ma.Audit(dir)

	if report.TotalEntries != 2 {
		t.Errorf("TotalEntries = %d, want 2", report.TotalEntries)
	}
	if report.TotalSize <= 0 {
		t.Errorf("TotalSize should be > 0, got %d", report.TotalSize)
	}
	if len(report.AutoFiles) != 0 {
		t.Errorf("AutoFiles = %d, want 0", len(report.AutoFiles))
	}
	if report.NeedsConsolidation {
		t.Errorf("NeedsConsolidation should be false for regular files only")
	}
}

// TestMemoryAuditAutoFiles triggers consolidation when exceeding threshold.
func TestMemoryAuditAutoFiles(t *testing.T) {
	dir := t.TempDir()
	// Write enough auto files to exceed threshold (5)
	for i := 0; i < 7; i++ {
		writeFile(t, filepath.Join(dir, "auto-"+itos(i)+".md"), "data")
	}

	ma := NewMemoryAudit()
	report := ma.Audit(dir)

	if report.TotalEntries != 7 {
		t.Errorf("TotalEntries = %d, want 7", report.TotalEntries)
	}
	if len(report.AutoFiles) != 7 {
		t.Errorf("AutoFiles = %d, want 7", len(report.AutoFiles))
	}
	if !report.NeedsConsolidation {
		t.Errorf("NeedsConsolidation should be true for 7 auto files")
	}
}

// TestMemoryAuditMixedFiles reports only auto-* files in AutoFiles.
func TestMemoryAuditMixedFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "core.md"), "core memory")
	writeFile(t, filepath.Join(dir, "auto-session-1.md"), "auto data")
	writeFile(t, filepath.Join(dir, "auto-session-2.md"), "more auto data")
	writeFile(t, filepath.Join(dir, "user-notes.md"), "user notes")

	ma := NewMemoryAudit()
	report := ma.Audit(dir)

	if report.TotalEntries != 4 {
		t.Errorf("TotalEntries = %d, want 4", report.TotalEntries)
	}
	if len(report.AutoFiles) != 2 {
		t.Errorf("AutoFiles = %d, want 2", len(report.AutoFiles))
	}
	if report.NeedsConsolidation {
		t.Errorf("NeedsConsolidation should be false for only 2 auto files")
	}
}

// TestMemoryAuditNonExistentDir does not panic.
func TestMemoryAuditNonExistentDir(t *testing.T) {
	ma := NewMemoryAudit()
	report := ma.Audit("/tmp/nonexistent-dir-12345")
	// Should return empty report, not panic
	if report.TotalEntries != 0 {
		t.Errorf("TotalEntries = %d, want 0", report.TotalEntries)
	}
}

// TestMemoryAuditFileSizes verifies that TotalSize is calculated correctly.
func TestMemoryAuditFileSizes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.md"), "12345")       // 5 bytes
	writeFile(t, filepath.Join(dir, "b.md"), "1234567890") // 10 bytes

	ma := NewMemoryAudit()
	report := ma.Audit(dir)

	if report.TotalEntries != 2 {
		t.Errorf("TotalEntries = %d, want 2", report.TotalEntries)
	}
	if report.TotalSize != 15 {
		t.Errorf("TotalSize = %d, want 15", report.TotalSize)
	}
}

// TestCheckFallback returns true when auto files exceed threshold.
func TestCheckFallback(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 7; i++ {
		writeFile(t, filepath.Join(dir, "auto-"+itos(i)+".md"), "data")
	}

	ma := NewMemoryAudit()
	if !ma.CheckFallback(dir) {
		t.Errorf("CheckFallback should return true for 7 auto files")
	}
}

// TestCheckFallbackBelowThreshold returns false.
func TestCheckFallbackBelowThreshold(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "auto-1.md"), "data")
	writeFile(t, filepath.Join(dir, "auto-2.md"), "data")

	ma := NewMemoryAudit()
	if ma.CheckFallback(dir) {
		t.Errorf("CheckFallback should return false for only 2 auto files")
	}
}

// TestCheckFallbackEmptyDir returns false.
func TestCheckFallbackEmptyDir(t *testing.T) {
	dir := t.TempDir()
	ma := NewMemoryAudit()
	if ma.CheckFallback(dir) {
		t.Errorf("CheckFallback should return false for empty dir")
	}
}

// TestCheckFallbackEmptyString returns false.
func TestCheckFallbackEmptyString(t *testing.T) {
	ma := NewMemoryAudit()
	if ma.CheckFallback("") {
		t.Errorf("CheckFallback should return false for empty string")
	}
}

// writeFile creates a file with the given content. Fails the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

// itos converts an integer to a string for test helper.
func itos(i int) string {
	return string(rune('0' + i))
}
