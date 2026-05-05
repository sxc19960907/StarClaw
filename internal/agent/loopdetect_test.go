package agent

import (
	"testing"
)

func TestNewLoopDetector(t *testing.T) {
	ld := NewLoopDetector()
	if ld == nil {
		t.Fatal("NewLoopDetector() returned nil")
	}
	if ld.historySize != 20 {
		t.Errorf("historySize = %d, want 20", ld.historySize)
	}
	if ld.consecDupThreshold != 2 {
		t.Errorf("consecDupThreshold = %d, want 2", ld.consecDupThreshold)
	}
}

func TestLoopDetector_ConsecutiveDuplicate(t *testing.T) {
	ld := NewLoopDetector()

	// Record 3 identical calls
	for i := 0; i < 3; i++ {
		ld.Record("grep", `{"pattern":"test"}`, false, "", "")
	}

	action, msg := ld.Check("grep")
	if action != LoopForceStop {
		t.Errorf("Expected LoopForceStop for 3 consecutive duplicates, got %v", action)
	}
	if msg == "" {
		t.Error("Expected nudge/stop message")
	}
}

func TestLoopDetector_NudgeConsecutive(t *testing.T) {
	ld := NewLoopDetector()

	// 2 identical calls should nudge (not stop)
	ld.Record("grep", `{"pattern":"test"}`, false, "", "")
	ld.Record("grep", `{"pattern":"test"}`, false, "", "")

	action, _ := ld.Check("grep")
	if action != LoopNudge {
		t.Errorf("Expected LoopNudge for 2 consecutive duplicates, got %v", action)
	}
}

func TestLoopDetector_NoLoop(t *testing.T) {
	ld := NewLoopDetector()

	// Different tools between calls — no loop
	ld.Record("file_read", `{"path":"a.go"}`, false, "", "")
	ld.Record("grep", `{"pattern":"test"}`, false, "", "")
	ld.Record("file_read", `{"path":"b.go"}`, false, "", "") // different file

	action, _ := ld.Check("file_read")
	if action != LoopContinue {
		t.Errorf("Expected LoopContinue for non-loop pattern, got %v", action)
	}
}

func TestLoopDetector_SameToolErrors(t *testing.T) {
	ld := NewLoopDetector()
	ld.sameToolErrThreshold = 2 // lower threshold for test

	// 4 errors on same tool -> force stop (threshold*2)
	for i := 0; i < 4; i++ {
		ld.Record("bash", `{"command":"invalid"}`, true, "command not found: invalid", "")
	}

	action, msg := ld.Check("bash")
	if action != LoopForceStop {
		t.Errorf("Expected LoopForceStop for persistent errors, got %v", action)
	}
	if msg == "" {
		t.Error("Expected stop message")
	}
}

func TestLoopDetector_NoProgress(t *testing.T) {
	ld := NewLoopDetector()
	ld.noProgressThreshold = 2 // lower threshold for test

	// 4 calls of same tool with different args -> force stop
	// Use a tool NOT in any family to avoid family-based checks
	for i := 0; i < 4; i++ {
		ld.Record("file_read", `{"path":"/tmp/file`+string(rune('a'+i))+`.go"}`, false, "", "")
	}

	action, _ := ld.Check("file_read")
	if action != LoopForceStop {
		t.Errorf("Expected LoopForceStop for no-progress, got %v", action)
	}
}

func TestLoopDetector_SleepDetection(t *testing.T) {
	ld := NewLoopDetector()

	// 2 sleep commands -> nudge
	ld.Record("bash", `{"command":"sleep 5"}`, false, "", "")
	ld.Record("bash", `{"command":"sleep 3 && echo done"}`, false, "", "")

	action, _ := ld.Check("bash")
	if action != LoopNudge {
		t.Errorf("Expected LoopNudge for sleep detection, got %v", action)
	}
}

func TestLoopDetector_ExactDuplicateSpread(t *testing.T) {
	ld := NewLoopDetector()
	ld.exactDupThreshold = 2 // lower for test

	// Same args appearing 4 times spread across window -> force stop
	for i := 0; i < 4; i++ {
		ld.Record("file_read", `{"path":"main.go"}`, false, "", "")
		// Insert different calls between
		if i < 3 {
			ld.Record("grep", `{"pattern":"different"}`, false, "", "")
		}
	}

	action, _ := ld.Check("file_read")
	if action != LoopForceStop {
		t.Errorf("Expected LoopForceStop for spread duplicates, got %v", action)
	}
}

func TestLoopDetector_SingleCall(t *testing.T) {
	ld := NewLoopDetector()

	ld.Record("file_read", `{"path":"main.go"}`, false, "", "")

	action, _ := ld.Check("file_read")
	if action != LoopContinue {
		t.Errorf("Single call should be LoopContinue, got %v", action)
	}
}

func TestLoopDetector_HistoryClears(t *testing.T) {
	ld := NewLoopDetector()
	ld.historySize = 5

	// Fill beyond history size
	for i := 0; i < 10; i++ {
		ld.Record("bash", `{"command":"echo`+string(rune('a'+i%26))+`"}`, false, "", "")
	}

	if len(ld.history) > ld.historySize {
		t.Errorf("History size %d exceeds max %d", len(ld.history), ld.historySize)
	}
}

func TestIsSleepCommand(t *testing.T) {
	tests := []struct {
		args     string
		expected bool
	}{
		{`{"command":"sleep 5"}`, true},
		{`{"command":"sleep 1 && echo done"}`, true},
		{`{"command":"echo hello"}`, false},
		{`{"command":"cat sleep.log"}`, false},
		{`{"command":""}`, false},
		{`{}`, false},
		{`invalid`, false},
	}

	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			if got := isSleepCommand(tt.args); got != tt.expected {
				t.Errorf("isSleepCommand(%q) = %v, want %v", tt.args, got, tt.expected)
			}
		})
	}
}

func TestNormalizeWebQuery(t *testing.T) {
	// Basic extraction
	result := normalizeWebQuery(`{"query":"how to test Go code"}`)
	if result == "" {
		t.Error("Expected non-empty normalized query")
	}

	// Date stripping
	result = normalizeWebQuery(`{"query":"latest news 2026-01-15 today"}`)
	if result != "" {
		t.Logf("Normalized: %s", result)
	}

	// Empty input
	if normalizeWebQuery(`{}`) != "" {
		t.Error("Expected empty for no query")
	}
	if normalizeWebQuery(`{"command":"echo"}`) != "" {
		t.Error("Expected empty for bash command")
	}
}

func TestTruncateErrSig(t *testing.T) {
	short := "short error"
	if truncateErrSig(short, 100) != short {
		t.Error("Short string should not be truncated")
	}

	long := ""
	for i := 0; i < 200; i++ {
		long += "x"
	}
	truncated := truncateErrSig(long, 100)
	if len([]rune(truncated)) != 100 {
		t.Errorf("Expected 100 runes, got %d", len([]rune(truncated)))
	}
}
