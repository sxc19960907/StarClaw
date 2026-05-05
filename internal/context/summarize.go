package context

import (
	"context"
	"fmt"
	"strings"

	"github.com/starclaw/starclaw/internal/client"
)

const summarizePrompt = `Compress the following conversation into a concise summary using a two-phase approach.

Phase 1 — Write a chronological analysis inside <analysis> tags:
- Walk through the conversation in order
- Note every user correction, decision, or preference change
- Track files read, modified, or created
- Record errors, blockers, and their resolutions

Phase 2 — Write the final summary inside <summary> tags:
- Distill the analysis into what a continuation needs to know
- Preserve user corrections and decisions (these are highest priority)
- Include current task state and next steps
- Be factual and brief

Format your response as:
<analysis>
[chronological walkthrough]
</analysis>
<summary>
[concise summary for continuation]
</summary>`

// Completer is the interface for making LLM completion calls.
type Completer interface {
	Complete(ctx context.Context, req client.CompletionRequest) (*client.CompletionResponse, error)
}

// GenerateSummary calls the LLM (small tier) to summarize a conversation.
func GenerateSummary(ctx context.Context, c Completer, messages []client.Message) (string, error) {
	var transcript strings.Builder
	for _, m := range messages {
		if m.Role == "system" {
			continue
		}
		text := messageText(m)
		if text == "" {
			continue
		}
		fmt.Fprintf(&transcript, "[%s]: %s\n\n", m.Role, text)
	}

	req := client.CompletionRequest{
		Messages: []client.Message{
			{Role: "system", Content: summarizePrompt},
			{Role: "user", Content: transcript.String()},
		},
		ModelTier:   "small",
		Temperature: 0.2,
		MaxTokens:   2000,
	}

	resp, err := c.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("summarization failed: %w", err)
	}

	return extractSummary(resp.OutputText), nil
}

// extractSummary extracts <summary> content from a two-phase response.
func extractSummary(raw string) string {
	raw = strings.TrimSpace(raw)

	if start := strings.Index(raw, "<summary>"); start >= 0 {
		after := raw[start+len("<summary>"):]
		if end := strings.Index(after, "</summary>"); end >= 0 {
			return strings.TrimSpace(after[:end])
		}
		return strings.TrimSpace(after)
	}

	// No summary tags — strip any analysis block and return remainder
	result := raw
	for {
		start := strings.Index(result, "<analysis>")
		if start < 0 {
			break
		}
		end := strings.Index(result, "</analysis>")
		if end < 0 {
			result = result[:start]
			break
		}
		result = result[:start] + result[end+len("</analysis>"):]
	}

	result = strings.TrimSpace(result)
	if result == "" {
		return ""
	}
	return result
}

// messageText extracts readable text from a message string.
// For StarClaw's string-based Content, truncates long tool results.
func messageText(m client.Message) string {
	text := m.Content
	if text == "" {
		return ""
	}
	// Truncate long content for summarization
	if r := []rune(text); len(r) > 1000 {
		text = string(r[:1000]) + "..."
	}
	return text
}
