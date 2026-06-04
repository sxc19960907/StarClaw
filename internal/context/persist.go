package context

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/filelock"
)

const persistPrompt = `You are extracting durable knowledge from a conversation before context is compacted.

Review the conversation and identify facts worth remembering in FUTURE conversations. Focus on:
- Decisions made (technical, design, business, or personal preferences)
- User corrections or preferences about how they want to work
- Important facts about projects, people, systems, or environments
- Patterns, gotchas, or insights discovered
- Configuration, setup, or process details that were hard to find
- Contacts, resources, or reference information mentioned

Do NOT include:
- Current task progress or status (captured separately)
- Verbatim code, file contents, or command output
- Ephemeral information only relevant to this conversation
- Things already present in the existing memory shown below

Format rules:
- Return a markdown bulleted list, one fact per bullet
- Each bullet should be a SHORT one-line summary (max ~100 chars)
- If nothing new is worth persisting, return exactly "NONE"`

const (
	// MaxMemoryLines is the maximum lines in MEMORY.md before overflow to detail file.
	MaxMemoryLines = 150

	consolidateThreshold = 12
	consolidateCooldown  = 7 * 24 * time.Hour

	consolidatePrompt = `You are consolidating an AI agent's auto-persisted memory entries.

These entries were extracted automatically from past conversations. Many are duplicated or superseded by newer entries.

Merge them into a single clean list:
- Deduplicate facts that say the same thing
- When entries conflict, keep the most recent version
- Drop ephemeral or stale information no longer useful
- Keep: decisions, preferences, patterns, gotchas, references, contacts, people

Output a markdown bulleted list, one fact per bullet (max ~100 chars each).
Target: under 100 lines total.
If nothing worth keeping survives, return exactly "NONE".`
)

// PersistLearnings extracts durable knowledge from a conversation and appends
// it to MEMORY.md before context compaction discards the messages.
func PersistLearnings(ctx context.Context, c Completer, messages []client.Message, memoryDir string) error {
	if memoryDir == "" || c == nil {
		return nil
	}

	memoryPath := filepath.Join(memoryDir, "MEMORY.md")
	existingMemory, _ := os.ReadFile(memoryPath)

	var transcript strings.Builder
	for _, m := range messages {
		if m.Role == "system" {
			continue
		}
		text := m.Content
		if text == "" {
			continue
		}
		if r := []rune(text); len(r) > 500 {
			text = string(r[:500]) + "..."
		}
		fmt.Fprintf(&transcript, "[%s]: %s\n\n", m.Role, text)
	}

	if transcript.Len() == 0 {
		return nil
	}

	var userMsg strings.Builder
	if len(existingMemory) > 0 {
		fmt.Fprintf(&userMsg, "## Existing Memory (do not duplicate)\n\n%s\n\n---\n\n", string(existingMemory))
	}
	fmt.Fprintf(&userMsg, "## Conversation to Extract From\n\n%s", transcript.String())

	req := client.CompletionRequest{
		Messages: []client.Message{
			{Role: "system", Content: persistPrompt},
			{Role: "user", Content: userMsg.String()},
		},
		ModelTier:   "small",
		Temperature: 0.2,
		MaxTokens:   1000,
	}

	resp, err := c.Complete(ctx, req)
	if err != nil {
		return fmt.Errorf("persist learnings failed: %w", err)
	}

	result := strings.TrimSpace(resp.OutputText)
	if result == "" || strings.EqualFold(result, "NONE") {
		return nil
	}

	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02 15:04")
	entry := fmt.Sprintf("\n## Auto-persisted (%s)\n\n%s\n", timestamp, result)

	f, err := os.OpenFile(memoryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(entry); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// ──────────────────────────────────────────────
// BoundedAppend — flock-protected memory append
// ──────────────────────────────────────────────

// memoryDirCtxKey is used to pass the memory directory through context.
type memoryDirCtxKey struct{}

// WithMemoryDir returns a context with the memory directory set.
func WithMemoryDir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, memoryDirCtxKey{}, dir)
}

// MemoryDirFromContext returns the memory directory from context, or "".
func MemoryDirFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(memoryDirCtxKey{}).(string); ok {
		return v
	}
	return ""
}

// BoundedAppend appends content to MEMORY.md with flock protection.
// If the result would exceed MaxMemoryLines, content is written to a
// timestamped detail file and a one-line pointer is added to MEMORY.md.
func BoundedAppend(memoryDir, content string) error {
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}

	memoryPath := filepath.Join(memoryDir, "MEMORY.md")
	lockPath := memoryPath + ".lock"

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("open lock: %w", err)
	}
	defer func() {
		_ = lockFile.Close()
	}()

	if err := filelock.Exclusive(lockFile); err != nil {
		return err
	}
	defer func() {
		_ = filelock.Unlock(lockFile)
	}()

	existing, _ := os.ReadFile(memoryPath)
	writeContent := content
	if len(existing) > 0 && !strings.HasPrefix(content, "\n") {
		writeContent = "\n" + writeContent
	}

	projectedLines := countLines(append(append([]byte{}, existing...), []byte(writeContent)...))

	if projectedLines > MaxMemoryLines {
		detailFile, err := writeDetailFile(memoryDir, content)
		if err != nil {
			return fmt.Errorf("write detail file: %w", err)
		}
		timestamp := time.Now().Format("2006-01-02")
		writeContent = fmt.Sprintf("\n- [%s] See [%s](%s) for details\n", timestamp, detailFile, detailFile)
	}

	f, err := os.OpenFile(memoryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(writeContent)
	return err
}

func writeDetailFile(memoryDir, content string) (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random suffix: %w", err)
	}
	suffix := hex.EncodeToString(b)

	timestamp := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("auto-%s-%s.md", timestamp, suffix)
	path := filepath.Join(memoryDir, filename)

	body := fmt.Sprintf("# Auto-persisted Learnings (%s)\n\n%s\n", timestamp, content)
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		return "", err
	}
	return filename, nil
}

func countLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	n := bytes.Count(content, []byte{'\n'})
	if content[len(content)-1] != '\n' {
		n++
	}
	return n
}

// ──────────────────────────────────────────────
// ConsolidateMemory — LLM-based fragment merging
// ──────────────────────────────────────────────

// ConsolidateMemory merges auto-persisted detail files and auto sections in
// MEMORY.md into a single deduplicated block via an LLM call. Runs only when
// ≥12 auto-*.md files exist and last GC was ≥7 days ago.
// Safe for concurrent use (flock on MEMORY.md.lock).
func ConsolidateMemory(ctx context.Context, c Completer, memoryDir string) error {
	if memoryDir == "" || c == nil {
		return nil
	}

	autoFiles, err := filepath.Glob(filepath.Join(memoryDir, "auto-*.md"))
	if err != nil || len(autoFiles) < consolidateThreshold {
		return nil
	}

	markerPath := filepath.Join(memoryDir, ".memory_gc")
	if info, err := os.Stat(markerPath); err == nil {
		if time.Since(info.ModTime()) < consolidateCooldown {
			return nil
		}
	}

	memoryPath := filepath.Join(memoryDir, "MEMORY.md")
	lockPath := memoryPath + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("consolidate: open lock: %w", err)
	}
	defer func() {
		_ = lockFile.Close()
	}()
	if err := filelock.Exclusive(lockFile); err != nil {
		return fmt.Errorf("consolidate: %w", err)
	}
	defer func() {
		_ = filelock.Unlock(lockFile)
	}()

	existing, _ := os.ReadFile(memoryPath)
	userContent, autoFromMemory := splitMemory(string(existing))

	var autoContent strings.Builder
	var consumedFiles []string
	if autoFromMemory != "" {
		autoContent.WriteString(autoFromMemory)
		autoContent.WriteString("\n\n")
	}
	for _, f := range autoFiles {
		data, readErr := os.ReadFile(f)
		if readErr != nil {
			consumedFiles = append(consumedFiles, f)
			continue
		}
		consumedFiles = append(consumedFiles, f)
		autoContent.WriteString(string(data))
		autoContent.WriteString("\n")
	}

	if autoContent.Len() == 0 {
		return nil
	}

	req := client.CompletionRequest{
		Messages: []client.Message{
			{Role: "system", Content: consolidatePrompt},
			{Role: "user", Content: autoContent.String()},
		},
		ModelTier:   "small",
		Temperature: 0.2,
		MaxTokens:   2000,
	}

	resp, err := c.Complete(ctx, req)
	if err != nil {
		return fmt.Errorf("consolidate: LLM call failed: %w", err)
	}

	consolidated := strings.TrimSpace(resp.OutputText)

	var result strings.Builder
	if userContent != "" {
		result.WriteString(userContent)
	}
	if consolidated != "" && !strings.EqualFold(consolidated, "NONE") {
		if result.Len() > 0 {
			result.WriteString("\n\n")
		}
		timestamp := time.Now().Format("2006-01-02 15:04")
		fmt.Fprintf(&result, "## Auto-consolidated (%s)\n\n%s\n", timestamp, consolidated)
	}

	tmpPath := memoryPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(result.String()), 0644); err != nil {
		return fmt.Errorf("consolidate: write temp: %w", err)
	}
	if err := os.Rename(tmpPath, memoryPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("consolidate: rename: %w", err)
	}

	for _, f := range consumedFiles {
		os.Remove(f)
	}

	if err := os.WriteFile(markerPath, []byte(time.Now().Format(time.RFC3339)), 0644); err != nil {
		return fmt.Errorf("consolidate: write marker: %w", err)
	}

	return nil
}

// splitMemory separates MEMORY.md into user-written and auto-generated sections.
func splitMemory(content string) (userContent, autoContent string) {
	lines := strings.Split(content, "\n")
	var user, auto []string
	inAuto := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## Auto-persisted") || strings.HasPrefix(trimmed, "## Auto-consolidated") {
			inAuto = true
			auto = append(auto, line)
			continue
		}
		if strings.HasPrefix(trimmed, "- [") && strings.Contains(trimmed, "See [auto-") && strings.HasSuffix(trimmed, "for details") {
			auto = append(auto, line)
			continue
		}
		if inAuto && (strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "# ")) {
			inAuto = false
		}

		if inAuto {
			auto = append(auto, line)
		} else {
			user = append(user, line)
		}
	}

	return strings.TrimSpace(strings.Join(user, "\n")),
		strings.TrimSpace(strings.Join(auto, "\n"))
}
