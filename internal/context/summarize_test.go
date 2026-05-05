package context

import (
	"context"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
)

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "both tags",
			input:    "<analysis>some analysis here</analysis>\n<summary>This is the summary</summary>",
			expected: "This is the summary",
		},
		{
			name:     "summary only",
			input:    "<summary>Just a summary</summary>",
			expected: "Just a summary",
		},
		{
			name:     "no tags",
			input:    "Just plain text summary",
			expected: "Just plain text summary",
		},
		{
			name:     "analysis only",
			input:    "<analysis>lots of analysis</analysis>",
			expected: "",
		},
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \n  ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSummary(tt.input)
			if result != tt.expected {
				t.Errorf("extractSummary(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMessageText(t *testing.T) {
	m := client.Message{Role: "user", Content: "Hello, world!"}
	text := messageText(m)
	if text != "Hello, world!" {
		t.Errorf("messageText = %q, want 'Hello, world!'", text)
	}

	// Empty content
	m.Content = ""
	text = messageText(m)
	if text != "" {
		t.Error("Empty content should return empty string")
	}
}

// mockCompleter for testing
type mockCompleter struct {
	output string
	err    error
}

func (m *mockCompleter) Complete(ctx context.Context, req client.CompletionRequest) (*client.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &client.CompletionResponse{OutputText: m.output}, nil
}

func TestGenerateSummary_WithMock(t *testing.T) {
	mock := &mockCompleter{output: "<summary>This is a test summary.</summary>"}
	msgs := []client.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi there"},
	}
	summary, err := GenerateSummary(context.Background(), mock, msgs)
	if err != nil {
		t.Fatalf("GenerateSummary failed: %v", err)
	}
	if summary != "This is a test summary." {
		t.Errorf("Summary = %q, want 'This is a test summary.'", summary)
	}
}

func TestMessageText_Truncation(t *testing.T) {
	// Content longer than 1000 runes should be truncated
	longContent := ""
	for i := 0; i < 2000; i++ {
		longContent += "x"
	}
	m := client.Message{Role: "user", Content: longContent}
	text := messageText(m)
	if len([]rune(text)) > 1100 { // 1000 + "..."
		t.Errorf("Long content should be truncated, got %d runes", len([]rune(text)))
	}
	if len([]rune(text)) < 1000 {
		t.Error("Short content should NOT be truncated")
	}
}
