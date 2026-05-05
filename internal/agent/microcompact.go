package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/starclaw/starclaw/internal/client"
)

const (
	microCompactMarker = "[micro-compact] "

	// microCompactMinChars is the minimum content length for LLM summarization.
	microCompactMinChars = 2000

	// microCompactMaxPerPass caps LLM attempts per invocation.
	microCompactMaxPerPass = 2
)

// microCompactSkipTools lists tools whose results should never be micro-compacted.
var microCompactSkipTools = map[string]bool{
	"think":          true,
	"cloud_delegate": true,
}

const microCompactPrompt = `Summarize this tool result in 1-2 sentences. Preserve exact error strings, file paths, URLs, IDs, and numbers when present. Focus on the final outcome or conclusion.

Tool: %s
Result:
%s`

// SmallCompleter is a minimal interface for LLM completion used by micro-compact.
type SmallCompleter interface {
	Complete(ctx context.Context, req client.CompletionRequest) (*client.CompletionResponse, error)
}

// microCompactResult uses a small LLM tier to produce a 1-2 sentence semantic
// summary of a tool result. Returns ("", false) if summarization fails or is
// skipped, signaling the caller to fall back to mechanical truncation.
func microCompactResult(ctx context.Context, c SmallCompleter, toolName, content string) (string, bool) {
	if c == nil {
		return "", false
	}

	prompt := fmt.Sprintf(microCompactPrompt, toolName, content)

	resp, err := c.Complete(ctx, client.CompletionRequest{
		Messages: []client.Message{
			{Role: "user", Content: client.NewTextContent(prompt)},
		},
		ModelTier:   "small",
		Temperature: 0.0,
		MaxTokens:   200,
	})
	if err != nil || resp.OutputText == "" {
		return "", false
	}

	summary := strings.TrimSpace(resp.OutputText)
	if summary == "" {
		return "", false
	}

	return microCompactMarker + summary, true
}

// isMicroCompacted returns true if the content was already summarized by micro-compact.
func isMicroCompacted(content string) bool {
	return strings.HasPrefix(content, microCompactMarker)
}
