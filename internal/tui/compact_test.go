package tui

import (
	"strings"
	"testing"
	"time"
)

func TestNewCompactView(t *testing.T) {
	cv := NewCompactView()
	if cv == nil {
		t.Fatal("NewCompactView() returned nil")
	}
	if cv.Summary() != "" {
		t.Error("New CompactView should have empty summary")
	}
}

func TestCompactView_Add(t *testing.T) {
	cv := NewCompactView()
	cv.Add("bash", `{"command":"ls"}`, "exit 0", false, 100*time.Millisecond)
	cv.Add("file_read", `{"path":"test.txt"}`, "content", false, 50*time.Millisecond)

	summary := cv.Summary()
	if summary == "" {
		t.Fatal("Summary should not be empty after adding results")
	}
	if !strings.Contains(summary, "2 tools") {
		t.Errorf("Summary should mention '2 tools', got: %s", summary)
	}
}

func TestCompactView_AddEntry(t *testing.T) {
	cv := NewCompactView()
	cv.AddEntry(CompactResult{
		Name:    "bash",
		Args:    `{"command":"ls"}`,
		Content: "file1\nfile2",
		IsError: false,
	})

	summary := cv.Summary()
	if summary == "" {
		t.Fatal("Summary should not be empty after AddEntry")
	}
}

func TestCompactView_Clear(t *testing.T) {
	cv := NewCompactView()
	cv.Add("bash", "{}", "ok", false, 0)
	cv.Clear()

	if cv.Summary() != "" {
		t.Error("Summary should be empty after Clear")
	}
}

func TestCompactView_ErrorCount(t *testing.T) {
	cv := NewCompactView()
	cv.Add("bash", "{}", "fail", true, 0)
	cv.Add("read", "{}", "ok", false, 0)
	cv.Add("write", "{}", "ok", false, 0)

	summary := cv.Summary()
	if summary == "" {
		t.Fatal("Summary should not be empty")
	}
}

func TestCompactView_AllErrors(t *testing.T) {
	cv := NewCompactView()
	cv.Add("bash", "{}", "err1", true, 0)
	cv.Add("read", "{}", "err2", true, 0)

	summary := cv.Summary()
	if summary == "" {
		t.Fatal("Summary should not be empty")
	}
}

func TestCompactView_MultipleEntries(t *testing.T) {
	cv := NewCompactView()
	for i := 0; i < 10; i++ {
		cv.Add("tool", "{}", "ok", false, 0)
	}

	summary := cv.Summary()
	if !strings.Contains(summary, "10 tools") {
		t.Errorf("Summary should mention '10 tools', got: %s", summary)
	}
}

func TestCompactView_ConsecutiveClear(t *testing.T) {
	cv := NewCompactView()
	cv.Add("bash", "{}", "ok", false, 0)
	cv.Clear()
	cv.Clear() // Should not panic
	if cv.Summary() != "" {
		t.Error("Summary should be empty after consecutive Clear")
	}
}

func TestCompactResult_Fields(t *testing.T) {
	now := time.Now()
	cr := CompactResult{
		Name:    "bash",
		Args:    `{"command":"echo hi"}`,
		Content: "hi",
		IsError: false,
		Elapsed: now.Sub(time.Now().Add(-time.Second)),
	}

	if cr.Name != "bash" {
		t.Errorf("CompactResult.Name = %q, want %q", cr.Name, "bash")
	}
	if cr.IsError {
		t.Error("CompactResult.IsError should be false")
	}
}
