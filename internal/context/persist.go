package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/client"
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
		// Truncate long tool results in transcript
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
	defer f.Close()
	_, err = f.WriteString(entry)
	return err
}
