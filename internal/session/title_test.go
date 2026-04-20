package session

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTitle_TruncatesLongText(t *testing.T) {
	longText := strings.Repeat("a", 100)
	title := GenerateTitle(longText)
	assert.Len(t, title, 50)
}

func TestGenerateTitle_CleansWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello  World", "Hello World"},
		{"Hello\tWorld", "Hello World"},
		{"Hello\nWorld", "Hello World"},
		{"  Hello World  ", "Hello World"},
		{"Hello   World   Test", "Hello World Test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GenerateTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateTitle_DefaultForEmpty(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"   "},
		{"\t\n\r"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GenerateTitle(tt.input)
			assert.Equal(t, "New session", result)
		})
	}
}

func TestGenerateTitle_NormalText(t *testing.T) {
	input := "Help me refactor the database code"
	result := GenerateTitle(input)
	assert.Equal(t, input, result)
}
