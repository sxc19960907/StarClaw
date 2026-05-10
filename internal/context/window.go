// Package context provides context window management for the agent loop.
package context

import (
	"math"

	"github.com/starclaw/starclaw/internal/client"
)

const (
	// charsPerToken is the conservative estimation ratio.
	// 3.5 chars/token handles mixed English/code/CJK better than 4.
	charsPerToken = 3.5

	// overheadPerMessage accounts for role, formatting, and separator tokens.
	overheadPerMessage = 4

	// compactThreshold is the fraction of context window that triggers compaction.
	compactThreshold = 0.85

	// defaultKeepLast is the default number of recent turn pairs to keep.
	defaultKeepLast = 20

	// minKeepLast is the minimum recent turn pairs to keep under budget pressure.
	minKeepLast = 3
)

// MinShapeable returns the minimum messages needed for shaping to have any effect.
func MinShapeable() int {
	return 1 + minKeepLast*2 // first message + N turn pairs
}

// EstimateTokens returns a heuristic token count for messages.
// Uses chars/3.5 + 4 overhead per message. For CJK text, estimate is ~2x conservative.
func EstimateTokens(messages []client.Message) int {
	total := 0
	for _, m := range messages {
		chars := len([]rune(m.Content))
		tokens := int(math.Ceil(float64(chars) / charsPerToken))
		total += tokens + overheadPerMessage
	}
	return total
}

// ShouldCompact returns true if total tokens exceed 85% of the context window.
func ShouldCompact(inputTokens, outputTokens, contextWindow int) bool {
	if contextWindow <= 0 {
		return false
	}
	threshold := int(float64(contextWindow) * compactThreshold)
	return inputTokens+outputTokens >= threshold
}

// ShapeHistory builds a sliding window: [first user] + [summary] + [last N turns].
// If summary is non-empty, it's injected as a user message between the first user and recent turns.
// If the history fits without shaping, it's returned as-is.
func ShapeHistory(messages []client.Message, summary string, contextWindow int) []client.Message {
	if len(messages) <= 1+minKeepLast*2 {
		return messages
	}
	if len(messages) <= 1+defaultKeepLast*2 && (contextWindow <= 0 || EstimateTokens(messages) < contextWindow) {
		return messages
	}

	// Use the first message as anchor regardless of role
	firstMsg := messages[0]
	rest := messages[1:]

	for keep := defaultKeepLast; keep >= minKeepLast; keep-- {
		keepMsgs := keep * 2
		if keepMsgs > len(rest) {
			keepMsgs = len(rest)
		}
		shaped := make([]client.Message, 0, 3+keepMsgs)
		shaped = append(shaped, firstMsg)
		if summary != "" {
			shaped = append(shaped, client.Message{
				Role:    "user",
				Content: "Previous context summary: " + summary,
			})
		}
		shaped = append(shaped, rest[len(rest)-keepMsgs:]...)

		if contextWindow <= 0 || EstimateTokens(shaped) < contextWindow {
			return shaped
		}
	}

	// Floor: minKeepLast even if over budget
	keepMsgs := minKeepLast * 2
	if keepMsgs > len(rest) {
		keepMsgs = len(rest)
	}
	shaped := make([]client.Message, 0, 3+keepMsgs)
	shaped = append(shaped, firstMsg)
	if summary != "" {
		shaped = append(shaped, client.Message{
			Role:    "user",
			Content: "Previous context summary: " + summary,
		})
	}
	shaped = append(shaped, rest[len(rest)-keepMsgs:]...)
	return shaped
}

// CompressOldToolResults truncates tool results older than compressAfter turns
// from the end of the message list. Recent turns keep full results for context.
// oldTurns is the number of turn pairs (assistant+user) to skip from the end.
func CompressOldToolResults(messages []client.Message, oldTurns, maxResultChars int) {
	if len(messages) < 2 || oldTurns <= 0 {
		return
	}

	// Count turn pairs from the end.
	// A turn pair is assistant→user; count only that transition to avoid double-counting.
	pairs := 0
	cutoff := len(messages)
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "assistant" && i+1 < len(messages) && messages[i+1].Role == "user" {
			pairs++
		}
		if pairs >= oldTurns {
			cutoff = i
			break
		}
	}

	// Truncate tool result messages before cutoff
	for i := 0; i < cutoff; i++ {
		m := &messages[i]
		if m.Role == "user" {
			r := []rune(m.Content)
			if len(r) > maxResultChars {
				m.Content = string(r[:maxResultChars]) + "... [truncated]"
			}
		}
	}
}
