package agent

import (
	"testing"
)

func TestAnalyzeShape_PureText(t *testing.T) {
	rs := AnalyzeShape("Hello, how can I help you today?")
	if rs.Type != ShapePureText {
		t.Errorf("expected pure_text, got %s", rs.Type)
	}
	if rs.TextLen == 0 {
		t.Error("expected non-zero TextLen")
	}
}

func TestAnalyzeShape_EmptyText(t *testing.T) {
	rs := AnalyzeShape("")
	if rs.Type != ShapePureText {
		t.Errorf("expected pure_text for empty, got %s", rs.Type)
	}
	if rs.TextLen != 0 {
		t.Errorf("expected TextLen 0, got %d", rs.TextLen)
	}
}

func TestAnalyzeShape_WhitespaceOnly(t *testing.T) {
	rs := AnalyzeShape("   \n\t  ")
	if rs.Type != ShapePureText {
		t.Errorf("expected pure_text for whitespace, got %s", rs.Type)
	}
}

func TestAnalyzeShape_ToolCallsXML(t *testing.T) {
	text := `<tool_use>
  <tool_name>read_file</tool_name>
  <input>{"path": "/tmp/test.txt"}</input>
</tool_use>`
	rs := AnalyzeShape(text)
	if rs.Type != ShapeToolCalls {
		t.Errorf("expected tool_calls, got %s", rs.Type)
	}
}

func TestAnalyzeShape_ToolCallsJSON(t *testing.T) {
	text := `{"type": "tool_use", "name": "read_file", "input": {"path": "/tmp/test.txt"}}`
	rs := AnalyzeShape(text)
	if rs.Type != ShapeToolCalls {
		t.Errorf("expected tool_calls, got %s; text=%q", rs.Type, text)
	}
}

func TestAnalyzeShape_Mixed(t *testing.T) {
	text := `I will read the file now.
<tool_use>
  <tool_name>read_file</tool_name>
  <input>{"path": "/tmp/test.txt"}</input>
</tool_use>
Let me show you the results.`
	rs := AnalyzeShape(text)
	if rs.Type != ShapeMixed {
		t.Errorf("expected mixed, got %s", rs.Type)
	}
}

func TestAnalyzeShape_ErrorPattern(t *testing.T) {
	text := `[error] The operation failed with status code 500.`
	rs := AnalyzeShape(text)
	if rs.Type != ShapeErrorPattern {
		t.Errorf("expected error_pattern, got %s", rs.Type)
	}
	if rs.Details == "" {
		t.Error("expected non-empty Details for error pattern")
	}
}

func TestAnalyzeShape_ErrorMixedWithToolCall(t *testing.T) {
	// When error is present, error_pattern takes priority
	text := `Error: connection refused
<tool_use>
  <tool_name>read_file</tool_name>
  <input>{"path": "/tmp/test.txt"}</input>
</tool_use>`
	rs := AnalyzeShape(text)
	if rs.Type != ShapeErrorPattern {
		t.Errorf("expected error_pattern (takes priority), got %s", rs.Type)
	}
}

func TestAnalyzeShape_ErrorDetailExtraction(t *testing.T) {
	text := `Some operation started.
Error: connection refused to database server at 10.0.0.1:5432
Proceeding with fallback.`
	rs := AnalyzeShape(text)
	if rs.Details != "Error: connection refused to database server at 10.0.0.1:5432" {
		t.Errorf("unexpected detail: %q", rs.Details)
	}
}

func TestAnalyzeShape_ErrorDetailTruncation(t *testing.T) {
	longLine := "Error: " + string(make([]byte, 200)) + " suffix"
	rs := AnalyzeShape(longLine)
	if len(rs.Details) > 123 { // 120 chars + "..." = 123
		t.Errorf("details too long: %d chars", len(rs.Details))
	}
}

func TestAnalyzeShape_Patterns(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected ShapeType
	}{
		{"failed", "The request failed with status 503", ShapeErrorPattern},
		{"exception", "NullPointerException occurred", ShapeErrorPattern},
		{"panic", "panic: runtime error", ShapeErrorPattern},
		{"stack trace", "stack trace:\n  at main.go:42", ShapeErrorPattern},
		{"case insensitive error", "ERROR: something broke", ShapeErrorPattern},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rs := AnalyzeShape(tc.text)
			if rs.Type != tc.expected {
				t.Errorf("AnalyzeShape(%q) = %s; want %s", tc.text, rs.Type, tc.expected)
			}
		})
	}
}

func TestAnalyzeShape_TextLenAccuracy(t *testing.T) {
	text := "hello world"
	rs := AnalyzeShape(text)
	if rs.TextLen != len(text) {
		t.Errorf("TextLen = %d; want %d", rs.TextLen, len(text))
	}
}
