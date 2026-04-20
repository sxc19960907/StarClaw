package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactSecrets_AWSKey(t *testing.T) {
	input := "Access key: AKIAIOSFODNN7EXAMPLE"
	expected := "Access key: [REDACTED]"
	assert.Equal(t, expected, RedactSecrets(input))
}

func TestRedactSecrets_JWT(t *testing.T) {
	input := "Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_BearerToken(t *testing.T) {
	input := "Authorization: Bearer abc123def456"
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_EnvVar(t *testing.T) {
	input := "export API_KEY=secret123"
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_GitHubToken(t *testing.T) {
	input := "Token: ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")

	input = "Token: gho_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_SKKey(t *testing.T) {
	input := "sk-abc123def456ghi789jkl012mno345pqr678stu9"
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_KeyStyle(t *testing.T) {
	input := "key-abc123def456ghi789jkl012mno345pqr678stu9"
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_PEM(t *testing.T) {
	input := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA..."
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_GenericAPIKey(t *testing.T) {
	input := `api_key: "abcdef1234567890abcdef1234567890"`
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")

	input = `api-key = "abcdef1234567890abcdef1234567890"`
	assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_NoSecrets(t *testing.T) {
	input := "This is normal text without secrets"
	assert.Equal(t, input, RedactSecrets(input))
}

func TestRedactSecrets_MultipleSecrets(t *testing.T) {
	input := "AWS: AKIAIOSFODNN7EXAMPLE and token: Bearer abc123def456"
	result := RedactSecrets(input)
	assert.Contains(t, result, "[REDACTED]")
	assert.NotContains(t, result, "AKIAIOSFODNN7EXAMPLE")
	assert.NotContains(t, result, "Bearer abc123def456")
}

func TestTruncate(t *testing.T) {
	// Short text - unchanged
	assert.Equal(t, "hello", truncate("hello", 10))

	// Exactly at limit - unchanged
	assert.Equal(t, "helloworld", truncate("helloworld", 10))

	// Over limit - truncated with ...
	assert.Equal(t, "hello w...", truncate("hello world", 10))

	// Long text
	longText := strings.Repeat("a", 1000)
	truncated := truncate(longText, 500)
	assert.Equal(t, 500, len(truncated))
	assert.True(t, strings.HasSuffix(truncated, "..."))
}

func TestNewAuditLogger(t *testing.T) {
	tmpDir := t.TempDir()

	logger, err := NewAuditLogger(tmpDir)
	require.NoError(t, err)
	defer logger.Close()

	// Check log file exists
	logPath := filepath.Join(tmpDir, "audit.log")
	_, err = os.Stat(logPath)
	require.NoError(t, err, "Log file should exist")

	// Check file permissions (should be 0600)
	info, err := os.Stat(logPath)
	require.NoError(t, err)
	// Note: Windows doesn't support Unix permissions the same way
	// so we skip this check on Windows
	if os.PathSeparator == '/' {
		mode := info.Mode().Perm()
		assert.Equal(t, os.FileMode(0600), mode, "Log file should have 0600 permissions")
	}
}

func TestNewAuditLogger_EmptyDir(t *testing.T) {
	_, err := NewAuditLogger("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "logDir must not be empty")
}

func TestNewAuditLogger_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "nested", "logs")

	logger, err := NewAuditLogger(logDir)
	require.NoError(t, err)
	defer logger.Close()

	// Check directory exists
	_, err = os.Stat(logDir)
	require.NoError(t, err, "Log directory should be created")
}

func TestAuditLogger_Log(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewAuditLogger(tmpDir)
	require.NoError(t, err)
	defer logger.Close()

	entry := AuditEntry{
		Timestamp:     time.Now(),
		SessionID:     "test-session",
		ToolName:      "file_read",
		InputSummary:  `{"file_path": "/tmp/test"}`,
		OutputSummary: "file contents",
		Decision:      "approved",
		Approved:      true,
		DurationMs:    100,
	}

	logger.Log(entry)

	// Read and verify log file
	content, err := os.ReadFile(filepath.Join(tmpDir, "audit.log"))
	require.NoError(t, err)

	// Should be valid JSON
	var loggedEntry AuditEntry
	err = json.Unmarshal(content, &loggedEntry)
	require.NoError(t, err, "Log entry should be valid JSON")

	assert.Equal(t, "file_read", loggedEntry.ToolName)
	assert.Equal(t, "test-session", loggedEntry.SessionID)
	assert.Equal(t, "approved", loggedEntry.Decision)
	assert.True(t, loggedEntry.Approved)
	assert.Equal(t, int64(100), loggedEntry.DurationMs)
}

func TestAuditLogger_Log_Truncation(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewAuditLogger(tmpDir)
	require.NoError(t, err)
	defer logger.Close()

	// Create a long input that should be truncated
	longInput := strings.Repeat("a", 1000)

	entry := AuditEntry{
		Timestamp:     time.Now(),
		SessionID:     "test",
		ToolName:      "test_tool",
		InputSummary:  longInput,
		OutputSummary: longInput,
		Approved:      true,
		DurationMs:    50,
	}

	logger.Log(entry)

	// Read and verify log file
	content, err := os.ReadFile(filepath.Join(tmpDir, "audit.log"))
	require.NoError(t, err)

	var loggedEntry AuditEntry
	err = json.Unmarshal(content, &loggedEntry)
	require.NoError(t, err)

	// Should be truncated to maxSummaryLen (500)
	assert.LessOrEqual(t, len(loggedEntry.InputSummary), 500)
	assert.LessOrEqual(t, len(loggedEntry.OutputSummary), 500)
}

func TestAuditLogger_Log_Redaction(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewAuditLogger(tmpDir)
	require.NoError(t, err)
	defer logger.Close()

	entry := AuditEntry{
		Timestamp:     time.Now(),
		SessionID:     "test",
		ToolName:      "test_tool",
		InputSummary:  "key=secret123",
		OutputSummary: "normal output",
		Approved:      true,
		DurationMs:    50,
	}

	logger.Log(entry)

	// Read and verify log file
	content, err := os.ReadFile(filepath.Join(tmpDir, "audit.log"))
	require.NoError(t, err)

	// Secret should be redacted
	assert.Contains(t, string(content), "[REDACTED]")
	assert.NotContains(t, string(content), "secret123")
}

func TestAuditLogger_ThreadSafe(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewAuditLogger(tmpDir)
	require.NoError(t, err)
	defer logger.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			logger.Log(AuditEntry{
				Timestamp: time.Now(),
				SessionID: fmt.Sprintf("session-%d", n),
				ToolName:  "test",
			})
		}(i)
	}
	wg.Wait()

	// Verify all entries were written
	content, err := os.ReadFile(filepath.Join(tmpDir, "audit.log"))
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 100)

	// Verify each line is valid JSON
	for _, line := range lines {
		var entry AuditEntry
		err := json.Unmarshal([]byte(line), &entry)
		require.NoError(t, err, "Each line should be valid JSON: %s", line)
	}
}

func TestAuditLogger_Close(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewAuditLogger(tmpDir)
	require.NoError(t, err)

	// Close should work
	err = logger.Close()
	require.NoError(t, err)

	// Close again should not error
	err = logger.Close()
	require.NoError(t, err)

	// Log after close should be no-op (not panic)
	logger.Log(AuditEntry{
		ToolName: "test",
	})
}

func TestAuditLogger_Close_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	logger, err := NewAuditLogger(tmpDir)
	require.NoError(t, err)

	// Multiple closes should be safe
	for i := 0; i < 5; i++ {
		err = logger.Close()
		require.NoError(t, err)
	}
}
