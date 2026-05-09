package agent

import (
	"testing"
)

func TestNormalizeWebQuery_Basic(t *testing.T) {
	result := NormalizeWebQuery("how to test Go code")
	if result == "" {
		t.Fatal("expected non-empty normalized query")
	}
	// All tokens present, sorted alphabetically.
	expected := "code go how test to"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestNormalizeWebQuery_DateStripping(t *testing.T) {
	// "latest", "news", "today" are all filler words; "2026-01-15" is a date.
	// After stripping all of those, only filler words remain -> "[empty]".
	result := NormalizeWebQuery("latest news 2026-01-15 today")
	if result != "[empty]" {
		t.Errorf("got %q, want %q", result, "[empty]")
	}

	// Mixed date + meaningful content.
	result = NormalizeWebQuery("golang errors 2026-01-15 today")
	if result != "errors golang" {
		t.Errorf("got %q, want %q", result, "errors golang")
	}
}

func TestNormalizeWebQuery_StandaloneYear(t *testing.T) {
	result := NormalizeWebQuery("python 2026 framework")
	// "2026" should be stripped.
	if result != "framework python" {
		t.Errorf("got %q, want %q", result, "framework python")
	}
}

func TestNormalizeWebQuery_FillerWordsRemoved(t *testing.T) {
	result := NormalizeWebQuery("top breaking headlines news today")
	if result != "[empty]" {
		t.Errorf("all-filler query should return sentinel, got %q", result)
	}
}

func TestNormalizeWebQuery_EmptyInput(t *testing.T) {
	result := NormalizeWebQuery("")
	if result != "[empty]" {
		t.Errorf("empty input should return sentinel, got %q", result)
	}
}

func TestNormalizeWebQuery_Whitespace(t *testing.T) {
	result := NormalizeWebQuery("   ")
	if result != "[empty]" {
		t.Errorf("whitespace-only should return sentinel, got %q", result)
	}
}

func TestNormalizeWebQuery_TokenSorting(t *testing.T) {
	// Order-independent.
	a := NormalizeWebQuery("world hello")
	b := NormalizeWebQuery("hello world")
	if a != b {
		t.Errorf("order-independent queries should match: %q vs %q", a, b)
	}
}

func TestNormalizeWebQuery_PunctuationTrimmed(t *testing.T) {
	result := NormalizeWebQuery("hello, world!!!")
	if result != "hello world" {
		t.Errorf("got %q, want %q", result, "hello world")
	}
}

func TestNormalizeWebQuery_ShortTokensRemoved(t *testing.T) {
	result := NormalizeWebQuery("a bb c dd e")
	if result != "bb dd" {
		t.Errorf("got %q, want %q", result, "bb dd")
	}
}

func TestNormalizeWebQuery_UrlStripped(t *testing.T) {
	result := NormalizeWebQuery("check https://example.com/test page")
	if result != "check page" {
		t.Errorf("got %q, want %q", result, "check page")
	}
}

func TestNormalizeWebQuery_Lowercase(t *testing.T) {
	result := NormalizeWebQuery("HELLO WORLD")
	if result != "hello world" {
		t.Errorf("got %q, want %q", result, "hello world")
	}
}

func TestTruncateErrSig_Short(t *testing.T) {
	short := "short error"
	if TruncateErrSig(short, 100) != short {
		t.Error("short string should not be truncated")
	}
}

func TestTruncateErrSig_Long(t *testing.T) {
	long := ""
	for i := 0; i < 200; i++ {
		long += "x"
	}
	truncated := TruncateErrSig(long, 100)
	if len([]rune(truncated)) != 100 {
		t.Errorf("expected 100 runes, got %d", len([]rune(truncated)))
	}
}

func TestTruncateErrSig_Exact(t *testing.T) {
	// Build a string of exactly 50 runes.
	s := ""
	for i := 0; i < 50; i++ {
		s += "a"
	}
	got := TruncateErrSig(s, 50)
	if got != s {
		t.Errorf("exact-length string should not be truncated, got len=%d", len([]rune(got)))
	}
}

func TestTruncateErrSig_Empty(t *testing.T) {
	if TruncateErrSig("", 10) != "" {
		t.Error("empty string should return empty")
	}
}

func TestTruncateErrSig_Zero(t *testing.T) {
	if TruncateErrSig("hello", 0) != "" {
		t.Error("zero limit should return empty")
	}
}

func TestTruncateErrSig_Negative(t *testing.T) {
	if TruncateErrSig("hello", -1) != "" {
		t.Error("negative limit should return empty")
	}
}

func TestTruncateErrSig_Unicode(t *testing.T) {
	s := "hello world"
	got := TruncateErrSig(s, 5)
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestExtractResultSignature_Empty(t *testing.T) {
	if ExtractResultSignature("") != "" {
		t.Error("empty content should return empty")
	}
	if ExtractResultSignature("no urls here") != "" {
		t.Error("content without URLs should return empty")
	}
}

func TestExtractResultSignature_SingleURL(t *testing.T) {
	content := "visit https://example.com/page"
	result := ExtractResultSignature(content)
	if result != "https://example.com/page" {
		t.Errorf("got %q, want %q", result, "https://example.com/page")
	}
}

func TestExtractResultSignature_MultipleURLs(t *testing.T) {
	content := "site1: https://example.com/a site2: https://example.org/b"
	result := ExtractResultSignature(content)
	expected := "https://example.com/a,https://example.org/b"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestExtractResultSignature_QueryStripped(t *testing.T) {
	content := "see https://example.com/page?foo=bar&baz=1"
	result := ExtractResultSignature(content)
	if result != "https://example.com/page" {
		t.Errorf("query string not stripped, got %q", result)
	}
}

func TestExtractResultSignature_FragmentStripped(t *testing.T) {
	content := "see https://example.com/page#section"
	result := ExtractResultSignature(content)
	if result != "https://example.com/page" {
		t.Errorf("fragment not stripped, got %q", result)
	}
}

func TestExtractResultSignature_Deduplication(t *testing.T) {
	content := "a: https://example.com/page b: https://example.com/page"
	result := ExtractResultSignature(content)
	if result != "https://example.com/page" {
		t.Errorf("duplicates should be deduplicated, got %q", result)
	}
}

func TestExtractResultSignature_TrailingPunctuation(t *testing.T) {
	content := "visit https://example.com/page."
	result := ExtractResultSignature(content)
	if result != "https://example.com/page" {
		t.Errorf("trailing punctuation should be trimmed, got %q", result)
	}
}

func TestExtractResultSignature_Sorted(t *testing.T) {
	content := "b: https://z.example.com a: https://a.example.com"
	result := ExtractResultSignature(content)
	expected := "https://a.example.com,https://z.example.com"
	if result != expected {
		t.Errorf("URLs should be sorted, got %q", result)
	}
}

func TestExtractResultSignature_Lowercase(t *testing.T) {
	content := "HTTPS://EXAMPLE.COM/PAGE"
	result := ExtractResultSignature(content)
	if result != "https://example.com/page" {
		t.Errorf("URLs should be lowercased, got %q", result)
	}
}
