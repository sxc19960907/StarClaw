package agent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/client"
)

// TestExecuteParallel executes multiple tools and verifies all results are collected.
func TestExecuteParallel(t *testing.T) {
	pc := NewPartitionConcurrency()
	tools := []ToolCall{
		{Name: "tool_a", Args: "a", Exec: okExec("result_a")},
		{Name: "tool_b", Args: "b", Exec: okExec("result_b")},
		{Name: "tool_c", Args: "c", Exec: okExec("result_c")},
	}

	results := pc.ExecuteParallel(tools, nil)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, r := range results {
		if r.IsError {
			t.Errorf("result %d should not be an error", i)
		}
	}
}

// TestExecuteParallelOrder verifies that results are returned in input order.
func TestExecuteParallelOrder(t *testing.T) {
	pc := NewPartitionConcurrency()
	tools := []ToolCall{
		{Name: "slow", Args: "", Exec: delayedExec("first", 50*time.Millisecond)},
		{Name: "fast", Args: "", Exec: okExec("second")},
	}

	results := pc.ExecuteParallel(tools, nil)
	if results[0].Content != "first" {
		t.Errorf("expected result[0]=%q, got %q", "first", results[0].Content)
	}
	if results[1].Content != "second" {
		t.Errorf("expected result[1]=%q, got %q", "second", results[1].Content)
	}
}

// TestExecuteParallelError verifies that tool errors are captured properly.
func TestExecuteParallelError(t *testing.T) {
	pc := NewPartitionConcurrency()
	execErr := errors.New("execution failure")
	tools := []ToolCall{
		{Name: "ok", Args: "", Exec: okExec("good")},
		{Name: "fail", Args: "", Exec: func(_ context.Context, _ string) (ToolResult, error) {
			return ToolResult{}, execErr
		}},
	}

	results := pc.ExecuteParallel(tools, nil)
	if results[0].IsError {
		t.Errorf("result 0 should not be an error")
	}
	if !results[1].IsError {
		t.Errorf("result 1 should be an error")
	}
	if results[1].Content != execErr.Error() {
		t.Errorf("expected error message %q, got %q", execErr.Error(), results[1].Content)
	}
}

// TestExecuteParallelPanic verifies that panicking tools are recovered and
// reported as errors.
func TestExecuteParallelPanic(t *testing.T) {
	pc := NewPartitionConcurrency()
	tools := []ToolCall{
		{Name: "panic", Args: "", Exec: func(_ context.Context, _ string) (ToolResult, error) {
			panic("something went wrong")
		}},
		{Name: "ok", Args: "", Exec: okExec("survivor")},
	}

	results := pc.ExecuteParallel(tools, nil)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].IsError {
		t.Errorf("panicking tool should produce an error result")
	}
	if results[1].IsError {
		t.Errorf("surviving tool should not be an error, got %q", results[1].Content)
	}
}

// TestExecuteParallelEmpty verifies that an empty tool slice returns nil.
func TestExecuteParallelEmpty(t *testing.T) {
	pc := NewPartitionConcurrency()
	results := pc.ExecuteParallel(nil, nil)
	if results != nil {
		t.Errorf("expected nil for empty input, got %v", results)
	}

	results = pc.ExecuteParallel([]ToolCall{}, nil)
	if results != nil {
		t.Errorf("expected nil for empty slice, got %v", results)
	}
}

// TestExecuteParallelHandler verifies that OnToolCall is called for each tool.
func TestExecuteParallelHandler(t *testing.T) {
	pc := NewPartitionConcurrency()
	var mu sync.Mutex
	var calls []string
	handler := &spyHandler{
		onToolCall: func(name, args string) {
			mu.Lock()
			calls = append(calls, name)
			mu.Unlock()
		},
	}

	tools := []ToolCall{
		{Name: "t1", Args: "a", Exec: okExec("r1")},
		{Name: "t2", Args: "b", Exec: okExec("r2")},
	}

	_ = pc.ExecuteParallel(tools, handler)

	mu.Lock()
	defer mu.Unlock()
	if len(calls) != 2 {
		t.Fatalf("expected 2 handler calls, got %d", len(calls))
	}
	if calls[0] != "t1" || calls[1] != "t2" {
		t.Errorf("expected calls [t1, t2], got %v", calls)
	}
}

// TestExecuteParallelBoundedConcurrency verifies that concurrency does not
// exceed the configured limit. We launch N tools where each takes a short
// time, and track how many are running concurrently.
func TestExecuteParallelBoundedConcurrency(t *testing.T) {
	pc := NewPartitionConcurrency()
	var maxConcurrent atomic.Int64
	var curConcurrent atomic.Int64
	var wg sync.WaitGroup

	n := 20
	tools := make([]ToolCall, n)
	for i := range tools {
		tools[i] = ToolCall{
			Name: "tool",
			Args: "",
			Exec: func(_ context.Context, _ string) (ToolResult, error) {
				wg.Add(1)
				defer wg.Done()
				v := curConcurrent.Add(1)
				for {
					current := maxConcurrent.Load()
					if v > current {
						if maxConcurrent.CompareAndSwap(current, v) {
							break
						}
					} else {
						break
					}
				}
				time.Sleep(20 * time.Millisecond)
				curConcurrent.Add(-1)
				return ToolResult{Content: "ok"}, nil
			},
		}
	}
	wg.Add(n)

	_ = pc.ExecuteParallel(tools, nil)

	if m := maxConcurrent.Load(); m > 5 {
		t.Errorf("max concurrency was %d, expected at most 5", m)
	}
}

// TestExecuteParallelNilExec verifies tools with nil Exec do not panic.
func TestExecuteParallelNilExec(t *testing.T) {
	pc := NewPartitionConcurrency()
	tools := []ToolCall{
		{Name: "nil_exec", Args: "", Exec: nil},
	}

	results := pc.ExecuteParallel(tools, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].IsError {
		t.Errorf("nil Exec should produce error")
	}
}

// okExec returns a function that always succeeds with the given content.
func okExec(content string) func(context.Context, string) (ToolResult, error) {
	return func(_ context.Context, _ string) (ToolResult, error) {
		return ToolResult{Content: content}, nil
	}
}

// delayedExec returns content after the specified delay.
func delayedExec(content string, d time.Duration) func(context.Context, string) (ToolResult, error) {
	return func(_ context.Context, _ string) (ToolResult, error) {
		time.Sleep(d)
		return ToolResult{Content: content}, nil
	}
}

// spyHandler records EventHandler callbacks for testing.
type spyHandler struct {
	onToolCall    func(name, args string)
	onToolResult  func(name string, result ToolResult)
	onText        func(text string)
	onUsage       func(usage struct{ InputTokens, OutputTokens int })
	onStreamDelta func(delta string)
	onPreamble    func(preamble string)
}

func (s *spyHandler) OnToolCall(name, args string) {
	if s.onToolCall != nil {
		s.onToolCall(name, args)
	}
}
func (s *spyHandler) OnToolResult(name string, result ToolResult) {
	if s.onToolResult != nil {
		s.onToolResult(name, result)
	}
}
func (s *spyHandler) OnText(text string) {
	if s.onText != nil {
		s.onText(text)
	}
}
func (s *spyHandler) OnUsage(usage client.Usage) {
	if s.onUsage != nil {
		s.onUsage(struct{ InputTokens, OutputTokens int }{usage.InputTokens, usage.OutputTokens})
	}
}
func (s *spyHandler) OnStreamDelta(delta string) {
	if s.onStreamDelta != nil {
		s.onStreamDelta(delta)
	}
}
func (s *spyHandler) OnPreamble(preamble string) {
	if s.onPreamble != nil {
		s.onPreamble(preamble)
	}
}
