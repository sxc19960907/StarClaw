package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/client"
)

// defaultKeepRecent is the number of most recent tool results to retain
// during time-based compaction.
const defaultKeepRecent = 3

// toolResultPlaceholderTemplate is the format string used to replace
// compacted tool results. The %d is replaced with maxAge in minutes.
const toolResultPlaceholderTemplate = "[tool result from %d minutes ago omitted]"

// TimeBasedCompactor replaces old tool results with a summary placeholder.
// Old tool outputs are less relevant and get replaced to save context space.
type TimeBasedCompactor struct {
	maxAge        time.Duration
	lastCompactAt time.Time
}

// NewTimeBasedCompactor creates a new TimeBasedCompactor.
// maxAge controls how often compaction runs. If maxAge has not elapsed since
// the last compaction, Compact is a no-op.
func NewTimeBasedCompactor(maxAge time.Duration) *TimeBasedCompactor {
	return &TimeBasedCompactor{
		maxAge:        maxAge,
		lastCompactAt: time.Now(),
	}
}

// Compact replaces old tool results with a summary placeholder.
// Tool results in positions beyond the most recent defaultKeepRecent are
// replaced with "[tool result from X minutes ago omitted]" when enough
// time has elapsed since the last compaction.
func (c *TimeBasedCompactor) Compact(messages []client.Message) []client.Message {
	if len(messages) == 0 {
		return messages
	}

	// Only compact if enough time has passed since the last compaction
	if time.Since(c.lastCompactAt) < c.maxAge {
		return messages
	}

	result := make([]client.Message, len(messages))
	copy(result, messages)

	// Collect tool result IDs in order from user messages
	var ids []string
	for _, msg := range result {
		if msg.Role != "user" {
			continue
		}
		for _, line := range strings.Split(msg.Content, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			var parsed map[string]any
			if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
				continue
			}
			typ, _ := parsed["type"].(string)
			if typ == "tool_result" {
				if toolUseID, ok := parsed["tool_use_id"].(string); ok && toolUseID != "" {
					ids = append(ids, toolUseID)
				}
			}
		}
	}

	if len(ids) <= defaultKeepRecent {
		return result
	}

	// All but the last defaultKeepRecent IDs should be compacted
	compactSet := make(map[string]bool, len(ids)-defaultKeepRecent)
	for _, id := range ids[:len(ids)-defaultKeepRecent] {
		compactSet[id] = true
	}

	placeholder := fmt.Sprintf(toolResultPlaceholderTemplate, int(c.maxAge.Minutes()))

	// Compact tool results by replacing their content field
	for i, msg := range result {
		if msg.Role != "user" {
			continue
		}

		lines := strings.Split(msg.Content, "\n")
		if len(lines) == 0 {
			continue
		}

		newLines := make([]string, 0, len(lines))
		modified := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				newLines = append(newLines, line)
				continue
			}

			var parsed map[string]any
			if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
				newLines = append(newLines, line)
				continue
			}

			typ, _ := parsed["type"].(string)
			toolUseID, _ := parsed["tool_use_id"].(string)

			if typ == "tool_result" && compactSet[toolUseID] {
				existingContent, _ := parsed["content"].(string)
				if existingContent != placeholder {
					parsed["content"] = placeholder
					newJSON, err := json.Marshal(parsed)
					if err == nil {
						newLines = append(newLines, string(newJSON))
						modified = true
						continue
					}
				}
			}
			newLines = append(newLines, line)
		}

		if modified {
			result[i].Content = strings.Join(newLines, "\n")
		}
	}

	c.lastCompactAt = time.Now()
	return result
}
