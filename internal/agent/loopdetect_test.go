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
		ld.Record("grep", `{"pattern":"test"}`, false, "", "", false, false)
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
	ld.Record("grep", `{"pattern":"test"}`, false, "", "", false, false)
	ld.Record("grep", `{"pattern":"test"}`, false, "", "", false, false)

	action, _ := ld.Check("grep")
	if action != LoopNudge {
		t.Errorf("Expected LoopNudge for 2 consecutive duplicates, got %v", action)
	}
}

func TestLoopDetector_NoLoop(t *testing.T) {
	ld := NewLoopDetector()

	// Different tools between calls — no loop
	ld.Record("file_read", `{"path":"a.go"}`, false, "", "", false, false)
	ld.Record("grep", `{"pattern":"test"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"b.go"}`, false, "", "", false, false) // different file

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
		ld.Record("bash", `{"command":"invalid"}`, true, "command not found: invalid", "", false, false)
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
		ld.Record("file_read", `{"path":"/tmp/file`+string(rune('a'+i))+`.go"}`, false, "", "", false, false)
	}

	action, _ := ld.Check("file_read")
	if action != LoopForceStop {
		t.Errorf("Expected LoopForceStop for no-progress, got %v", action)
	}
}

func TestLoopDetector_SleepDetection(t *testing.T) {
	ld := NewLoopDetector()

	// 2 sleep commands -> nudge
	ld.Record("bash", `{"command":"sleep 5"}`, false, "", "", false, false)
	ld.Record("bash", `{"command":"sleep 3 && echo done"}`, false, "", "", false, false)

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
		ld.Record("file_read", `{"path":"main.go"}`, false, "", "", false, false)
		// Insert different calls between
		if i < 3 {
			ld.Record("grep", `{"pattern":"different"}`, false, "", "", false, false)
		}
	}

	action, _ := ld.Check("file_read")
	if action != LoopForceStop {
		t.Errorf("Expected LoopForceStop for spread duplicates, got %v", action)
	}
}

func TestLoopDetector_SingleCall(t *testing.T) {
	ld := NewLoopDetector()

	ld.Record("file_read", `{"path":"main.go"}`, false, "", "", false, false)

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
		ld.Record("bash", `{"command":"echo`+string(rune('a'+i%26))+`"}`, false, "", "", false, false)
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

// ---- UnproductiveStreak Tests ----
//
// Use a non-family, non-repeatable tool with unique args each time to avoid
// interference from ConsecutiveDup or ExactDup detectors.

func TestLoopDetector_UnproductiveStreak_Nudge(t *testing.T) {
	ld := NewLoopDetector()

	// 5 consecutive non-actionable calls -> nudge. Use unique args to avoid consecutive dup.
	for i := 0; i < 5; i++ {
		ld.Record("custom_tool", `{"arg":"x`+string(rune('0'+i))+`"}`, false, "", "", true, false)
	}

	action, msg := ld.Check("custom_tool")
	if action != LoopNudge {
		t.Errorf("Expected LoopNudge for 5 unproductive calls, got %v", action)
	}
	if msg == "" {
		t.Error("Expected nudge message")
	}
}

func TestLoopDetector_UnproductiveStreak_ForceStop(t *testing.T) {
	ld := NewLoopDetector()

	// 8 consecutive non-actionable calls -> force stop
	for i := 0; i < 8; i++ {
		ld.Record("custom_tool", `{"arg":"x`+string(rune('0'+i))+`"}`, false, "", "", true, false)
	}

	action, msg := ld.Check("custom_tool")
	if action != LoopForceStop {
		t.Errorf("Expected LoopForceStop for 8 unproductive calls, got %v", action)
	}
	if msg == "" {
		t.Error("Expected stop message")
	}
}

func TestLoopDetector_UnproductiveStreak_ResetByProductive(t *testing.T) {
	ld := NewLoopDetector()

	// 4 non-actionable calls, then 1 productive resets the streak
	for i := 0; i < 4; i++ {
		ld.Record("custom_tool", `{"arg":"x`+string(rune('0'+i))+`"}`, false, "", "", true, false)
	}
	// Productive call resets the streak
	ld.Record("custom_tool", `{"arg":"found"}`, false, "", "", false, false)

	// Now 2 more non-actionable calls (streak = 2, not enough to trigger)
	ld.Record("custom_tool", `{"arg":"z1"}`, false, "", "", true, false)
	ld.Record("custom_tool", `{"arg":"z2"}`, false, "", "", true, false)

	action, _ := ld.Check("custom_tool")
	if action != LoopContinue {
		t.Errorf("Expected LoopContinue after productive call interrupts streak, got %v", action)
	}
}

func TestLoopDetector_UnproductiveStreak_DifferentToolResets(t *testing.T) {
	ld := NewLoopDetector()

	// 4 non-actionable custom_tool calls
	for i := 0; i < 4; i++ {
		ld.Record("custom_tool", `{"arg":"x`+string(rune('0'+i))+`"}`, false, "", "", true, false)
	}
	// A productive call of the same tool resets the streak
	ld.Record("custom_tool", `{"arg":"found"}`, false, "", "", false, false)

	action, _ := ld.Check("custom_tool")
	if action != LoopContinue {
		t.Errorf("Expected LoopContinue after productive call, got %v", action)
	}
}

func TestLoopDetector_UnproductiveStreak_SearchSkipped(t *testing.T) {
	ld := NewLoopDetector()

	// Search-family tools are handled by SearchEscalation (>=3 nudge, >=5 force stop).
	// 5 unproductive grep calls should trigger SearchEscalation ForceStop.
	for i := 0; i < 5; i++ {
		ld.Record("grep", `{"pattern":"nothing"}`, false, "", "", true, false)
	}

	action, _ := ld.Check("grep")
	if action != LoopForceStop {
		t.Errorf("SearchEscalation should force-stop at 5 consecutive search calls, got %v", action)
	}
}

// ---- FileReadRepeat Tests ----
//
// Use intervening different-arg calls to avoid ConsecutiveDup.
// Use different file paths to avoid ExactDup.

func TestLoopDetector_FileReadRepeat(t *testing.T) {
	ld := NewLoopDetector()

	// Read the same file 3 times with different intervening calls to avoid
	// triggering consecutive duplicate detector.
	ld.Record("file_read", `{"path":"main.go"}`, false, "", "", false, false)
	ld.Record("grep", `{"pattern":"x"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"main.go"}`, false, "", "", false, false)
	ld.Record("grep", `{"pattern":"y"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"main.go"}`, false, "", "", false, false)

	action, msg := ld.Check("file_read")
	if action != LoopNudge {
		t.Errorf("Expected LoopNudge for 3 same-file reads, got %v", action)
	}
	if msg == "" {
		t.Error("Expected nudge message")
	}
}

func TestLoopDetector_FileReadRepeat_DifferentFilesNoNudge(t *testing.T) {
	ld := NewLoopDetector()

	ld.Record("file_read", `{"path":"a.go"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"b.go"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"c.go"}`, false, "", "", false, false)

	action, _ := ld.Check("file_read")
	if action != LoopContinue {
		t.Errorf("Expected LoopContinue for different file reads, got %v", action)
	}
}

func TestLoopDetector_FileReadRepeat_TwoReadsNoNudge(t *testing.T) {
	ld := NewLoopDetector()

	// Only 2 reads of same file, not 3. With a different-args call between to
	// avoid consecutive duplicate detection.
	ld.Record("file_read", `{"path":"main.go"}`, false, "", "", false, false)
	ld.Record("grep", `{"pattern":"x"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"main.go"}`, false, "", "", false, false)

	action, _ := ld.Check("file_read")
	if action != LoopContinue {
		t.Errorf("Expected LoopContinue for 2 reads of same file, got %v", action)
	}
}

// ---- ToolModeFlipFlop Tests ----
//
// Use unique args for each call to avoid ExactDup and ConsecutiveDup.

func TestLoopDetector_ToolModeFlipFlop_Nudge(t *testing.T) {
	ld := NewLoopDetector()

	// A,B,A,B,A,B,A pattern = 3 cycles (A->B->A = 1 cycle, repeated 3x).
	// Use different args for each call to avoid triggering ExactDup.
	tools := []string{"file_read", "file_edit"}
	for i := 0; i < 7; i++ {
		arg := `{"path":"m` + string(rune('0'+i)) + `.go"}`
		if i%2 == 1 {
			arg = `{"path":"m` + string(rune('0'+i)) + `.go","new":"v` + string(rune('0'+i)) + `"}`
		}
		ld.Record(tools[i%2], arg, false, "", "", false, false)
	}

	action, msg := ld.Check("file_read")
	if action != LoopNudge {
		t.Errorf("Expected LoopNudge for 3 flip-flop cycles, got %v", action)
	}
	if msg == "" {
		t.Error("Expected nudge message")
	}
}

func TestLoopDetector_ToolModeFlipFlop_ForceStop(t *testing.T) {
	ld := NewLoopDetector()

	// 11 alternating calls = 5 cycles. Use unique args to avoid ExactDup.
	tools := []string{"file_read", "file_edit"}
	for i := 0; i < 11; i++ {
		arg := `{"path":"m` + string(rune('0'+i)) + `.go"}`
		if i%2 == 1 {
			arg = `{"path":"m` + string(rune('0'+i)) + `.go","new":"v` + string(rune('0'+i)) + `"}`
		}
		ld.Record(tools[i%2], arg, false, "", "", false, false)
	}

	action, msg := ld.Check("file_read")
	if action != LoopForceStop {
		t.Errorf("Expected LoopForceStop for 5 flip-flop cycles, got %v", action)
	}
	if msg == "" {
		t.Error("Expected stop message")
	}
}

func TestLoopDetector_ToolModeFlipFlop_SameToolBreaks(t *testing.T) {
	ld := NewLoopDetector()

	// file_read + file_edit twice, then file_read twice (breaks alternation),
	// then file_edit + file_read. The consecutive file_read should prevent
	// flip-flop from seeing a long enough alternating streak.
	// All args unique to avoid other detectors.
	ld.Record("file_read", `{"path":"a1.go"}`, false, "", "", false, false)
	ld.Record("file_edit", `{"path":"a2.go","new":"v1"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"a3.go"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"a4.go"}`, false, "", "", false, false) // same tool breaks
	ld.Record("file_edit", `{"path":"a5.go","new":"v2"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"a6.go"}`, false, "", "", false, false)

	action, _ := ld.Check("file_read")
	if action != LoopContinue {
		t.Errorf("Expected LoopContinue when same tool breaks alternation, got %v", action)
	}
}

func TestLoopDetector_ToolModeFlipFlop_ThreeToolCycle(t *testing.T) {
	ld := NewLoopDetector()

	// Three-way cycle file_read-grep-file_read-grep-file_read-grep-file_read.
	// This alternates between file_read and grep, BUT every other file_read
	// has a different path, so the flip-flop pattern involves the same two
	// tools alternating—wait, that IS a 2-tool flip-flop. Need 3 unique tools.
	//
	// Use bash, grep, file_read cycling: bash->grep->file_read->bash->...
	// This should NOT trigger flip-flop (needs strict 2-tool alternation).
	tools := []string{"bash", "grep", "file_read"}
	args := []string{
		`{"command":"echo a"}`,
		`{"pattern":"x"}`,
		`{"path":"a.go"}`,
		`{"command":"echo b"}`,
		`{"pattern":"y"}`,
		`{"path":"b.go"}`,
		`{"command":"echo c"}`,
	}
	for i, a := range args {
		ld.Record(tools[i%3], a, false, "", "", false, false)
	}

	action, _ := ld.Check("bash")
	if action != LoopContinue {
		t.Errorf("Expected LoopContinue for 3-way cycle (not a flip-flop), got %v", action)
	}
}

func TestLoopDetector_ToolModeFlipFlop_MinimumCalls(t *testing.T) {
	ld := NewLoopDetector()

	// Just 2 alternations (4 calls) should not trigger 3-cycle threshold.
	ld.Record("bash", `{"command":"echo a"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"a.go"}`, false, "", "", false, false)
	ld.Record("bash", `{"command":"echo b"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"a.go"}`, false, "", "", false, false)

	action, _ := ld.Check("file_read")
	if action != LoopContinue {
		t.Errorf("Expected LoopContinue for 1.5 cycles (below 3 threshold), got %v", action)
	}
}

// ---- IdentityCycle Tests ----

func TestLoopDetector_IdentityCycle(t *testing.T) {
	ld := NewLoopDetector()

	// First call is normal
	ld.Record("bash", `{"command":"echo hello"}`, false, "", "", false, false)
	// Second call has output == input
	ld.Record("bash", `{"command":"echo hello"}`, false, "", "", false, true)

	action, msg := ld.Check("bash")
	if action != LoopNudge {
		t.Errorf("Expected LoopNudge for identity cycle, got %v", action)
	}
	if msg == "" {
		t.Error("Expected nudge message")
	}
}

func TestLoopDetector_IdentityCycle_NotTriggered(t *testing.T) {
	ld := NewLoopDetector()

	// Normal calls with different args should not trigger identity cycle or
	// any other detector.
	ld.Record("file_read", `{"path":"main.go"}`, false, "", "", false, false)
	ld.Record("file_read", `{"path":"other.go"}`, false, "", "", false, false)

	action, _ := ld.Check("file_read")
	if action != LoopContinue {
		t.Errorf("Expected LoopContinue for normal calls with different args, got %v", action)
	}
}

// ---- Edge Case Tests ----

func TestLoopDetector_EmptyHistory(t *testing.T) {
	ld := NewLoopDetector()

	// Check with empty history should return LoopContinue
	if len(ld.history) != 0 {
		t.Fatal("Expected empty history")
	}

	action, msg := ld.Check("bash")
	if action != LoopContinue {
		t.Errorf("Empty history should produce LoopContinue, got %v", action)
	}
	if msg != "" {
		t.Errorf("Empty history should produce empty message, got %q", msg)
	}
}

func TestLoopDetector_SingleRecord(t *testing.T) {
	ld := NewLoopDetector()
	ld.Record("bash", `{"command":"echo hi"}`, false, "", "", false, false)

	if len(ld.history) != 1 {
		t.Fatal("Expected 1 record")
	}

	action, msg := ld.Check("bash")
	if action != LoopContinue {
		t.Errorf("Single record should produce LoopContinue, got %v", action)
	}
	if msg != "" {
		t.Errorf("Single record should produce empty message, got %q", msg)
	}
}

// ---- Utility Tests ----

func TestExtractFilePath(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		argsJSON string
		expected string
	}{
		{"file_read with path", "file_read", `{"path":"main.go"}`, "main.go"},
		{"file_read with empty path", "file_read", `{}`, ""},
		{"non-file tool", "bash", `{"path":"main.go"}`, ""},
		{"invalid JSON", "file_read", "not-json", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractFilePath(tt.toolName, tt.argsJSON); got != tt.expected {
				t.Errorf("extractFilePath(%q, %q) = %q, want %q", tt.toolName, tt.argsJSON, got, tt.expected)
			}
		})
	}
}
