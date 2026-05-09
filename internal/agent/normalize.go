package agent

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// fillerWords are common query padding that don't affect semantic meaning.
var fillerWords = map[string]bool{
	"today": true, "yesterday": true, "latest": true, "recent": true,
	"top": true, "major": true, "breaking": true, "headlines": true,
	"news": true, "current": true, "update": true, "updates": true,
}

var (
	isoDatePattern       = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`)
	standaloneYearPattern = regexp.MustCompile(`\b20\d{2}\b`)
	urlPattern           = regexp.MustCompile(`(?i)https?://[^\s"'<>\])\},]+`)
)

// NormalizeWebQuery produces a canonical form of a search query for loop detection.
// Strips dates, filler words, punctuation, lowercases, and sorts remaining tokens.
// Two queries about the same topic with different date/filler noise produce the
// same normalized string.
func NormalizeWebQuery(query string) string {
	query = strings.ToLower(query)

	// Strip dates.
	query = isoDatePattern.ReplaceAllString(query, " ")
	query = standaloneYearPattern.ReplaceAllString(query, " ")

	// Strip URLs.
	query = urlPattern.ReplaceAllString(query, "")

	// Tokenize, strip punctuation, filter filler and short tokens.
	tokens := strings.Fields(query)
	var cleaned []string
	for _, tok := range tokens {
		tok = strings.TrimFunc(tok, func(r rune) bool {
			return unicode.IsPunct(r) || unicode.IsSymbol(r)
		})
		if len(tok) < 2 {
			continue
		}
		if fillerWords[tok] {
			continue
		}
		cleaned = append(cleaned, tok)
	}

	sort.Strings(cleaned)
	if len(cleaned) == 0 {
		// All tokens were filler — return a sentinel so all-filler queries
		// match each other (prevents bypassing topic detection).
		return "[empty]"
	}
	return strings.Join(cleaned, " ")
}

// TruncateErrSig truncates an error signature to at most n runes.
// Rune-aware so multi-byte characters are not split.
func TruncateErrSig(sig string, n int) string {
	if n < 0 {
		return ""
	}
	r := []rune(sig)
	if len(r) <= n {
		return sig
	}
	return string(r[:n])
}

// ExtractResultSignature extracts unique URLs (domain+path) from text content,
// strips query strings and fragments, deduplicates, sorts, and returns as
// comma-separated. More granular than domain-only; reuters.com/climate does
// not equal reuters.com/economics.
func ExtractResultSignature(content string) string {
	matches := urlPattern.FindAllString(content, -1)
	if len(matches) == 0 {
		return ""
	}

	seen := make(map[string]bool)
	var urls []string
	for _, u := range matches {
		// Strip query string.
		if idx := strings.IndexByte(u, '?'); idx != -1 {
			u = u[:idx]
		}
		// Strip fragment.
		if idx := strings.IndexByte(u, '#'); idx != -1 {
			u = u[:idx]
		}
		// Trim trailing punctuation that leaked in.
		u = strings.TrimRight(u, ".,;:!)")
		u = strings.ToLower(u)
		if !seen[u] {
			seen[u] = true
			urls = append(urls, u)
		}
	}

	sort.Strings(urls)
	return strings.Join(urls, ",")
}
