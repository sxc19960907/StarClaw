package context

import (
	"encoding/json"
	"strings"

	"github.com/starclaw/starclaw/internal/client"
)

// SanitizeHistory repairs malformed message history that would cause API errors.
// Specifically:
//   - tool role messages with plain text → dropped
//   - empty assistant/user messages → dropped
//   - consecutive same-role messages → merged
//
// Returns a new slice; the original is not modified.
func SanitizeHistory(messages []client.Message) []client.Message {
	if len(messages) == 0 {
		return messages
	}

	// First pass: drop invalid messages
	var cleaned []client.Message
	for _, msg := range messages {
		if shouldDrop(msg) {
			continue
		}
		cleaned = append(cleaned, msg)
	}

	// Second pass: merge consecutive same-role messages
	merged := mergeConsecutiveRoles(cleaned)

	// Third pass: strip orphaned tool_use/tool_result pairs.
	// StarClaw encodes tool calls as JSON strings within Content.
	// Tool_use blocks have {"type":"tool_use","id":"...","name":"..."}
	// Tool_result blocks have {"type":"tool_result","tool_use_id":"..."}
	stripped := stripOrphanedStringTools(merged)

	// Final pass: re-merge after stripping
	return mergeConsecutiveRoles(stripped)
}

// shouldDrop returns true for messages that would cause API errors.
func shouldDrop(msg client.Message) bool {
	switch msg.Role {
	case "tool":
		return true
	case "assistant":
		if strings.TrimSpace(msg.Content) == "" {
			return true
		}
		if strings.HasPrefix(strings.TrimSpace(msg.Content), "[tool_call:") {
			return true
		}
	case "user":
		if strings.TrimSpace(msg.Content) == "" {
			return true
		}
	}
	return false
}

// mergeConsecutiveRoles collapses consecutive same-role messages.
func mergeConsecutiveRoles(messages []client.Message) []client.Message {
	var out []client.Message
	for i, msg := range messages {
		if i > 0 && msg.Role == messages[i-1].Role {
			switch msg.Role {
			case "assistant", "user":
				out[len(out)-1] = msg
				continue
			}
		}
		out = append(out, msg)
	}
	return out
}

// stripOrphanedStringTools removes unpaired tool_use and tool_result blocks
// from StarClaw's string-based message content. Tool calls are encoded as
// JSON objects within the content string separated by newlines.
func stripOrphanedStringTools(messages []client.Message) []client.Message {
	if len(messages) < 2 {
		return messages
	}

	// Parse tool IDs from assistant and user messages
	type toolInfo struct {
		id        string
		isToolUse bool // true for tool_use, false for tool_result
	}
	msgTools := make([][]toolInfo, len(messages))

	// Extract tool blocks from each message
	for i, msg := range messages {
		blocks := parseJSONBlocks(msg.Content)
		for _, b := range blocks {
			if b.typ == "tool_use" && b.id != "" {
				msgTools[i] = append(msgTools[i], toolInfo{id: b.id, isToolUse: true})
			}
			if b.typ == "tool_result" && b.toolUseID != "" {
				msgTools[i] = append(msgTools[i], toolInfo{id: b.toolUseID, isToolUse: false})
			}
		}
	}

	// Build valid ID set from adjacent pairs (assistant[i] + user[i+1])
	validAt := make([]map[string]bool, len(messages))
	for i := 0; i+1 < len(messages); i++ {
		if messages[i].Role != "assistant" || messages[i+1].Role != "user" {
			continue
		}
		useIDs := make(map[string]bool)
		for _, t := range msgTools[i] {
			if t.isToolUse {
				useIDs[t.id] = true
			}
		}
		for _, t := range msgTools[i+1] {
			if !t.isToolUse && useIDs[t.id] {
				if validAt[i] == nil {
					validAt[i] = make(map[string]bool)
				}
				if validAt[i+1] == nil {
					validAt[i+1] = make(map[string]bool)
				}
				validAt[i][t.id] = true
				validAt[i+1][t.id] = true
			}
		}
	}

	var out []client.Message
	for i, msg := range messages {
		blocks := parseJSONBlocks(msg.Content)
		if len(blocks) == 0 {
			out = append(out, msg)
			continue
		}

		// Filter orphaned blocks. validAt[i] may be nil (no valid pairs at this position).
		var kept []string
		hasToolBlock := false
		for _, b := range blocks {
			switch b.typ {
			case "tool_use":
				hasToolBlock = true
				if b.id != "" && (validAt[i] == nil || !validAt[i][b.id]) {
					continue
				}
			case "tool_result":
				hasToolBlock = true
				if b.toolUseID != "" && (validAt[i] == nil || !validAt[i][b.toolUseID]) {
					continue
				}
			}
			kept = append(kept, b.raw)
		}

		if len(kept) == 0 && (msg.Role == "assistant" || msg.Role == "user") {
			// Drop messages that become empty after stripping tool blocks
			// unless they have non-tool text content
			if !hasNonToolContent(blocks) {
				continue
			}
		}

		if hasToolBlock {
			out = append(out, client.Message{Role: msg.Role, Content: strings.Join(kept, "\n")})
		} else {
			out = append(out, msg)
		}
	}
	return out
}

// jsonBlock represents a parsed JSON block from message content.
type jsonBlock struct {
	typ        string // "tool_use", "tool_result", "text", "unknown"
	id         string // tool_use ID
	toolUseID  string // tool_result's tool_use_id
	raw        string // original JSON string
}

// parseJSONBlocks splits message content by newlines and tries to parse each
// line as a JSON block to extract tool_use/tool_result metadata.
func parseJSONBlocks(content string) []jsonBlock {
	lines := strings.Split(content, "\n")
	var blocks []jsonBlock
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			// Non-JSON text
			blocks = append(blocks, jsonBlock{typ: "text", raw: line})
			continue
		}

		typ, _ := obj["type"].(string)
		switch typ {
		case "tool_use":
			id, _ := obj["id"].(string)
			blocks = append(blocks, jsonBlock{typ: "tool_use", id: id, raw: line})
		case "tool_result":
			toolUseID, _ := obj["tool_use_id"].(string)
			blocks = append(blocks, jsonBlock{typ: "tool_result", toolUseID: toolUseID, raw: line})
		default:
			blocks = append(blocks, jsonBlock{typ: "unknown", raw: line})
		}
	}
	return blocks
}

// hasNonToolContent returns true if blocks contain text or unknown content (not just tool blocks).
func hasNonToolContent(blocks []jsonBlock) bool {
	for _, b := range blocks {
		if b.typ == "text" || b.typ == "unknown" {
			return true
		}
	}
	return false
}
