package session

import (
	"strings"
	"unicode"
)

// GenerateTitle creates a title from the first user message
func GenerateTitle(firstMessage string) string {
	// Truncate to 50 chars
	title := firstMessage
	if len(title) > 50 {
		title = title[:50]
	}

	// Clean up whitespace - replace all whitespace sequences with single space
	var cleaned strings.Builder
	inWhitespace := false
	for _, r := range title {
		if unicode.IsSpace(r) {
			if !inWhitespace {
				cleaned.WriteRune(' ')
				inWhitespace = true
			}
		} else {
			cleaned.WriteRune(r)
			inWhitespace = false
		}
	}

	result := strings.TrimSpace(cleaned.String())

	// If empty after cleanup, use default
	if result == "" {
		result = "New session"
	}

	return result
}
