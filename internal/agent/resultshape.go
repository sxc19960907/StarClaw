package agent

import (
	"regexp"
	"strings"
)

// ShapeType represents the category of an LLM response shape.
type ShapeType string

const (
	ShapePureText     ShapeType = "pure_text"
	ShapeToolCalls    ShapeType = "tool_calls"
	ShapeMixed        ShapeType = "mixed"
	ShapeErrorPattern ShapeType = "error_pattern"
)

// ResultShape describes the analytical shape of an LLM response text.
type ResultShape struct {
	Type    ShapeType
	TextLen int
	Details string
}

var (
	toolUseXMLPattern  = regexp.MustCompile(`<tool_use>\s*<tool_name>[^<]+</tool_name>`)
	toolUseJSONPattern = regexp.MustCompile(`"type"\s*:\s*"tool_use"`)
	toolBlockPattern   = regexp.MustCompile(`(?i)\btool_use\b.*\btool_name\b`)
	errorPattern       = regexp.MustCompile(`(?i)\berror\b|\bfailed\b|\bpanic\b|exception|\[error\]|stack trace`)
)

// AnalyzeShape inspects the text and classifies its shape.
func AnalyzeShape(text string) ResultShape {
	text = strings.TrimSpace(text)
	if text == "" {
		return ResultShape{Type: ShapePureText, TextLen: 0}
	}

	rs := ResultShape{TextLen: len(text)}

	hasToolCall := detectToolCall(text)
	hasError := detectErrorPattern(text)

	if hasError {
		rs.Type = ShapeErrorPattern
		rs.Details = extractErrorDetail(text)
		return rs
	}

	switch {
	case hasToolCall && hasSubstantialText(text):
		rs.Type = ShapeMixed
	case hasToolCall:
		rs.Type = ShapeToolCalls
	default:
		rs.Type = ShapePureText
	}

	return rs
}

func detectToolCall(text string) bool {
	return toolUseXMLPattern.MatchString(text) ||
		toolUseJSONPattern.MatchString(text) ||
		toolBlockPattern.MatchString(text)
}

func detectErrorPattern(text string) bool {
	return errorPattern.MatchString(text)
}

// hasSubstantialText returns true when the text contains non-trivial prose
// beyond just tool call markup.
func hasSubstantialText(text string) bool {
	// Strip tool call markup and check remaining content
	cleaned := toolUseXMLPattern.ReplaceAllString(text, "")
	cleaned = toolUseJSONPattern.ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)
	return len(cleaned) > 60
}

func extractErrorDetail(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if errorPattern.MatchString(line) {
			trimmed := strings.TrimSpace(line)
			if len([]rune(trimmed)) > 120 {
				trimmed = string([]rune(trimmed)[:120]) + "..."
			}
			return trimmed
		}
	}
	return ""
}
