package tools

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestWaitTool_Info(t *testing.T) {
	tool := &WaitTool{}
	info := tool.Info()
	if info.Name != "wait" {
		t.Errorf("Name = %q, want 'wait'", info.Name)
	}
}

func TestWaitTool_DefaultDuration(t *testing.T) {
	tool := &WaitTool{}
	start := time.Now()
	result, err := tool.Run(context.Background(), `{}`)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if elapsed < 4*time.Second || elapsed > 6*time.Second {
		t.Errorf("Default wait should be ~5s, took %v", elapsed)
	}
	if !strings.Contains(result.Content, "Waited") {
		t.Errorf("Expected 'Waited' message, got: %s", result.Content)
	}
}

func TestWaitTool_CustomDuration(t *testing.T) {
	tool := &WaitTool{}
	start := time.Now()
	result, err := tool.Run(context.Background(), `{"seconds":1.5}`)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if elapsed > 2*time.Second {
		t.Errorf("1.5s wait took too long: %v", elapsed)
	}
	if !strings.Contains(result.Content, "1.5") {
		t.Errorf("Should show wait duration: %s", result.Content)
	}
}

func TestWaitTool_MaxExceeded(t *testing.T) {
	tool := &WaitTool{}
	result, _ := tool.Run(context.Background(), `{"seconds":60}`)
	if !result.IsError {
		t.Error("Should reject > 30 seconds")
	}
}

func TestWaitTool_RequiresApproval(t *testing.T) {
	tool := &WaitTool{}
	if tool.RequiresApproval() {
		t.Error("wait should not require approval")
	}
}

func TestWaitTool_ContextCancel(t *testing.T) {
	tool := &WaitTool{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	result, _ := tool.Run(ctx, `{"seconds":5}`)
	if !result.IsError {
		t.Error("Cancelled context should return error")
	}
}
