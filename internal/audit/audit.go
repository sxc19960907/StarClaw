// Package audit provides audit logging for tool calls
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEntry represents a single audited tool call event
type AuditEntry struct {
	Timestamp     time.Time `json:"timestamp"`
	SessionID     string    `json:"session_id"`
	ToolName      string    `json:"tool_name"`
	InputSummary  string    `json:"input_summary"`
	OutputSummary string    `json:"output_summary"`
	Decision      string    `json:"decision"`
	Approved      bool      `json:"approved"`
	DurationMs    int64     `json:"duration_ms"`
}

// AuditLogger writes audit entries as JSON lines
type AuditLogger struct {
	mu     sync.Mutex
	file   *os.File
	closed bool
}

const maxSummaryLen = 500

// NewAuditLogger creates a logger that writes to logDir/audit.log
// Creates the directory if it does not exist
func NewAuditLogger(logDir string) (*AuditLogger, error) {
	if logDir == "" {
		return nil, fmt.Errorf("logDir must not be empty")
	}

	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	logPath := filepath.Join(logDir, "audit.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log %s: %w", logPath, err)
	}

	return &AuditLogger{file: f}, nil
}

// Log records a tool call event with automatic redaction
func (a *AuditLogger) Log(entry AuditEntry) {
	// Truncate and redact
	entry.InputSummary = RedactSecrets(truncate(entry.InputSummary, maxSummaryLen))
	entry.OutputSummary = RedactSecrets(truncate(entry.OutputSummary, maxSummaryLen))

	data, err := json.Marshal(entry)
	if err != nil {
		// Silently drop on marshal error - don't break the flow
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return
	}

	a.file.Write(data)
	a.file.Write([]byte("\n"))
}

// Close closes the underlying log file
func (a *AuditLogger) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return nil
	}
	a.closed = true
	return a.file.Close()
}

// truncate shortens text to maxLen, appending "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
