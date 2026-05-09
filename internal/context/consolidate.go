package context

import (
	"encoding/json"
	"strings"

	"github.com/starclaw/starclaw/internal/client"
)

// toolCallEntry represents a parsed tool call in the message history.
type toolCallEntry struct {
	assistantIdx int
	userIdx      int
	toolUseID    string
	toolName     string
	path         string // file path for file_read
	pattern      string // regex pattern for grep
	glob         string // glob filter for grep
	content      string // tool result content
	isError      bool
}

// collectToolCalls extracts tool call entries from message history.
// Walks through assistant-user message pairs and parses tool_use + tool_result blocks.
func collectToolCalls(messages []client.Message) []toolCallEntry {
	var entries []toolCallEntry

	for i := 0; i+1 < len(messages); i++ {
		if messages[i].Role != "assistant" || messages[i+1].Role != "user" {
			continue
		}

		assistantBlocks := parseJSONBlocks(messages[i].Content)
		userBlocks := parseJSONBlocks(messages[i+1].Content)

		// Build result map: toolUseID -> index in user blocks
		resultMap := make(map[string]int)
		for j, b := range userBlocks {
			if b.typ == "tool_result" && b.toolUseID != "" {
				resultMap[b.toolUseID] = j
			}
		}

		for _, ab := range assistantBlocks {
			if ab.typ != "tool_use" || ab.id == "" {
				continue
			}

			resultIdx, hasResult := resultMap[ab.id]
			if !hasResult {
				continue
			}

			// Parse tool_use for name and input
			var toolUseObj map[string]any
			if err := json.Unmarshal([]byte(ab.raw), &toolUseObj); err != nil {
				continue
			}
			name, _ := toolUseObj["name"].(string)

			// Extract relevant input parameters for consolidation
			var path, pattern, glob string
			if input, ok := toolUseObj["input"].(map[string]any); ok {
				if p, ok := input["path"].(string); ok {
					path = p
				}
				if p, ok := input["pattern"].(string); ok {
					pattern = p
				}
				if g, ok := input["glob"].(string); ok {
					glob = g
				}
			}

			// Parse tool result
			ub := userBlocks[resultIdx]
			var resultObj map[string]any
			if err := json.Unmarshal([]byte(ub.raw), &resultObj); err != nil {
				continue
			}
			content, _ := resultObj["content"].(string)
			isError, _ := resultObj["is_error"].(bool)

			entries = append(entries, toolCallEntry{
				assistantIdx: i,
				userIdx:      i + 1,
				toolUseID:    ab.id,
				toolName:     name,
				path:         path,
				pattern:      pattern,
				glob:         glob,
				content:      content,
				isError:      isError,
			})
		}
	}

	return entries
}

// isSameConsolidationTarget returns true if two consecutive tool entries
// should be considered for consolidation — same tool with the same identifying
// parameter (file path for file_read, pattern for grep).
func isSameConsolidationTarget(a, b toolCallEntry) bool {
	if a.toolName != b.toolName {
		return false
	}
	switch a.toolName {
	case "file_read":
		return a.path != "" && a.path == b.path
	case "grep":
		return a.pattern != "" && a.pattern == b.pattern
	default:
		return false
	}
}

// ConsolidateRedundant merges consecutive similar tool results.
//
// Rules:
//   - Consecutive file_read results for the same file: keep only the latest result.
//   - Consecutive grep results for the same pattern: merge results into the latest.
//
// The original message slice is not modified; a new slice is returned.
func ConsolidateRedundant(messages []client.Message) []client.Message {
	if len(messages) < 2 {
		return messages
	}

	entries := collectToolCalls(messages)
	if len(entries) < 2 {
		return messages
	}

	// Identify redundant groups and build operation plans
	toRemove := make(map[string]bool)       // toolUseID -> remove entirely
	mergedContent := make(map[string]string) // toolUseID -> new consolidated content

	i := 0
	for i < len(entries) {
		// Find end of consecutive group with same parameters
		j := i + 1
		for j < len(entries) && isSameConsolidationTarget(entries[j-1], entries[j]) {
			j++
		}

		groupSize := j - i
		if groupSize > 1 {
			switch entries[i].toolName {
			case "file_read":
				// Keep only the last entry in the group
				for k := i; k < j-1; k++ {
					toRemove[entries[k].toolUseID] = true
				}
			case "grep":
				// Merge content into the last entry
				var combined strings.Builder
				for k := i; k < j-1; k++ {
					toRemove[entries[k].toolUseID] = true
					if entries[k].content != "" {
						combined.WriteString(entries[k].content)
						combined.WriteString("\n")
					}
				}
				// Append last entry's content
				combined.WriteString(entries[j-1].content)
				mergedContent[entries[j-1].toolUseID] = combined.String()
			}
		}

		i = j
	}

	if len(toRemove) == 0 && len(mergedContent) == 0 {
		return messages
	}

	// Apply operations to messages
	result := make([]client.Message, 0, len(messages))

	for _, msg := range messages {
		blocks := parseJSONBlocks(msg.Content)
		if len(blocks) == 0 {
			result = append(result, msg)
			continue
		}

		var kept []string
		modified := false

		for _, b := range blocks {
			switch b.typ {
			case "tool_use":
				if toRemove[b.id] {
					modified = true
					continue // drop this tool_use
				}
				kept = append(kept, b.raw)

			case "tool_result":
				if toRemove[b.toolUseID] {
					modified = true
					continue // drop this tool_result
				}
				// Check if we need to update the content (grep merge)
				if newContent, needsMerge := mergedContent[b.toolUseID]; needsMerge {
					var obj map[string]any
					if err := json.Unmarshal([]byte(b.raw), &obj); err == nil {
						obj["content"] = newContent
						if newJSON, err := json.Marshal(obj); err == nil {
							kept = append(kept, string(newJSON))
							modified = true
							continue
						}
					}
				}
				kept = append(kept, b.raw)

			default:
				// Non-tool blocks (text, unknown) always pass through
				kept = append(kept, b.raw)
			}
		}

		if !modified {
			result = append(result, msg)
			continue
		}

		// Drop messages that become empty after removing tool blocks
		if len(kept) == 0 && (msg.Role == "assistant" || msg.Role == "user") {
			if !hasNonToolContent(blocks) {
				continue
			}
		}

		result = append(result, client.Message{
			Role:    msg.Role,
			Content: strings.Join(kept, "\n"),
		})
	}

	// Re-merge consecutive same-role messages after stripping
	return mergeConsecutiveRoles(result)
}
