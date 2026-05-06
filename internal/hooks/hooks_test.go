package hooks

import (
	"bytes"
	"context"
	"testing"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner(Config{})
	if r == nil {
		t.Fatal("NewRunner returned nil")
	}
	if r.timeout != defaultTimeout {
		t.Errorf("timeout = %v, want %v", r.timeout, defaultTimeout)
	}
}

func TestRunner_NilRunner(t *testing.T) {
	var r *Runner
	// All methods should be safe on nil
	if decision, _ := r.RunPreToolUse(context.Background(), "bash", `{}`, "sess"); decision != "" {
		t.Error("nil runner should return empty decision")
	}
	r.RunPostToolUse(context.Background(), "bash", `{}`, "ok", "sess")
	r.RunSessionStart(context.Background(), "sess")
	r.RunStop(context.Background(), "sess")
}

func TestRunner_EmptyConfig(t *testing.T) {
	r := NewRunner(Config{})
	decision, reason := r.RunPreToolUse(context.Background(), "bash", `{}`, "sess")
	if decision != "" || reason != "" {
		t.Errorf("empty config: decision=%q reason=%q", decision, reason)
	}
}

func TestRunner_PreToolUse_NoMatch(t *testing.T) {
	cfg := Config{
		PreToolUse: []Entry{
			{Matcher: "file_write", Command: "./test-hook.sh"},
		},
	}
	r := NewRunner(cfg)
	// file_read doesn't match file_write matcher
	d, _ := r.RunPreToolUse(context.Background(), "file_read", `{}`, "sess")
	if d != "" {
		t.Errorf("non-matching matcher should not trigger: got %q", d)
	}
}

func TestRunner_PreToolUse_MatchAll(t *testing.T) {
	cfg := Config{
		PreToolUse: []Entry{
			{Command: "/nonexistent/path"},
		},
	}
	r := NewRunner(cfg)
	// Empty matcher matches everything, but path doesn't exist → logs warning, doesn't deny
	d, _ := r.RunPreToolUse(context.Background(), "bash", `{}`, "sess")
	if d != "" {
		t.Errorf("failed hook should not deny: got %q", d)
	}
}

func TestRunner_Match(t *testing.T) {
	r := NewRunner(Config{})
	entries := []Entry{
		{Matcher: "file_.*", Command: "./a"},
		{Matcher: "", Command: "./b"}, // matches all
		{Matcher: "bash", Command: "./c"},
	}

	matched := r.match(entries, "file_read")
	if len(matched) != 2 {
		t.Errorf("file_read should match 2 entries (file_.* + empty), got %d", len(matched))
	}

	matched = r.match(entries, "bash")
	if len(matched) != 2 {
		t.Errorf("bash should match 2 entries (empty + bash), got %d", len(matched))
	}

	matched = r.match(entries, "think")
	if len(matched) != 1 {
		t.Errorf("think should match 1 entry (empty only), got %d", len(matched))
	}
}

func TestResolveCommand(t *testing.T) {
	// Empty
	_, err := resolveCommand("")
	if err == nil {
		t.Error("empty command should fail")
	}

	// Bare command
	_, err = resolveCommand("python3")
	if err == nil {
		t.Error("bare command should fail")
	}

	// Relative
	path, err := resolveCommand("./script.sh")
	if err != nil {
		t.Errorf("./script.sh should succeed: %v", err)
	}
	if path != "./script.sh" {
		t.Errorf("path = %q", path)
	}

	// Absolute outside ~/.starclaw
	_, err = resolveCommand("/usr/bin/python3")
	if err == nil {
		t.Error("absolute outside ~/.starclaw should fail")
	}
}

func TestToRawJSON(t *testing.T) {
	// Empty string
	if string(toRawJSON("")) != "null" {
		t.Error("empty should be null")
	}
	// Valid JSON
	if string(toRawJSON(`{"key":"value"}`)) != `{"key":"value"}` {
		t.Error("valid JSON should pass through")
	}
	// Plain string
	if string(toRawJSON("hello")) != `"hello"` {
		t.Errorf("plain string should be quoted: got %s", toRawJSON("hello"))
	}
}

func TestLimitedWriter(t *testing.T) {
	var b bytes.Buffer
	w := &limitedWriter{buf: &b, limit: 5}
	n, err := w.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Errorf("Write(hello): n=%d err=%v", n, err)
	}
	if b.String() != "hello" {
		t.Errorf("buf = %q", b.String())
	}
	// Write more than limit
	n, err = w.Write([]byte(" world"))
	if err != nil || n != 6 {
		t.Errorf("Write( world): n=%d err=%v", n, err)
	}
	// Buffer should not have grown beyond limit
	if b.Len() > 5 {
		t.Errorf("buffer exceeded limit: %d", b.Len())
	}
}

func TestRunner_RecursionGuard(t *testing.T) {
	r := NewRunner(Config{
		PreToolUse: []Entry{{Command: "/nonexistent"}},
	})

	// enterHook should return true on second call
	if r.enterHook() {
		t.Fatal("first enterHook should succeed")
	}
	if !r.enterHook() {
		t.Error("second enterHook should return true (already in hook)")
	}
	r.exitHook()
	if r.enterHook() {
		t.Error("should be able to enterHook after exitHook")
	}
	r.exitHook()
}
