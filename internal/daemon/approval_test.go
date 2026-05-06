package daemon

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestApprovalBrokerWaitAndResolve(t *testing.T) {
	broker := NewApprovalBroker()
	req := ApprovalRequest{
		RequestID: "test-1",
		Tool:      "bash",
		Args:      "echo hello",
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var decision ApprovalDecision
	var err error

	go func() {
		defer wg.Done()
		decision, err = broker.WaitForApproval(context.Background(), req)
	}()

	time.Sleep(10 * time.Millisecond)

	broker.Resolve(ApprovalResolvedPayload{
		RequestID: "test-1",
		Decision:  DecisionAllow,
	})

	wg.Wait()

	if err != nil {
		t.Fatalf("WaitForApproval returned error: %v", err)
	}
	if decision != DecisionAllow {
		t.Errorf("expected DecisionAllow, got %q", decision)
	}
}

func TestApprovalBrokerResolveDenied(t *testing.T) {
	broker := NewApprovalBroker()
	req := ApprovalRequest{
		RequestID: "test-deny",
		Tool:      "file_write",
		Args:      "/etc/passwd",
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var decision ApprovalDecision
	var err error

	go func() {
		defer wg.Done()
		decision, err = broker.WaitForApproval(context.Background(), req)
	}()

	time.Sleep(10 * time.Millisecond)

	broker.Resolve(ApprovalResolvedPayload{
		RequestID: "test-deny",
		Decision:  DecisionDeny,
	})

	wg.Wait()

	if err != nil {
		t.Fatalf("WaitForApproval returned error: %v", err)
	}
	if decision != DecisionDeny {
		t.Errorf("expected DecisionDeny, got %q", decision)
	}
}

func TestApprovalBrokerTimeout(t *testing.T) {
	broker := NewApprovalBrokerWithTimeout(10 * time.Millisecond)
	req := ApprovalRequest{
		RequestID: "test-timeout",
		Tool:      "http",
		Args:      "https://example.com",
	}

	decision, err := broker.WaitForApproval(context.Background(), req)
	if err != nil {
		t.Fatalf("WaitForApproval returned error: %v", err)
	}
	if decision != DecisionDeny {
		t.Errorf("expected DecisionDeny on timeout, got %q", decision)
	}
}

func TestApprovalBrokerContextCancelled(t *testing.T) {
	broker := NewApprovalBroker()
	req := ApprovalRequest{
		RequestID: "test-cancel",
		Tool:      "read",
		Args:      "secret.txt",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	decision, err := broker.WaitForApproval(ctx, req)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if decision != DecisionDeny {
		t.Errorf("expected DecisionDeny on cancel, got %q", decision)
	}
}

func TestApprovalBrokerResolveUnknown(t *testing.T) {
	broker := NewApprovalBroker()
	broker.Resolve(ApprovalResolvedPayload{
		RequestID: "nonexistent",
		Decision:  DecisionAllow,
	})
}

func TestApprovalBrokerConcurrentResolve(t *testing.T) {
	broker := NewApprovalBroker()
	req := ApprovalRequest{
		RequestID: "test-concurrent",
		Tool:      "bash",
		Args:      "echo concurrent",
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, _ = broker.WaitForApproval(context.Background(), req)
	}()

	time.Sleep(10 * time.Millisecond)

	broker.Resolve(ApprovalResolvedPayload{
		RequestID: "test-concurrent",
		Decision:  DecisionAllow,
	})
	broker.Resolve(ApprovalResolvedPayload{
		RequestID: "test-concurrent",
		Decision:  DecisionDeny,
	})

	wg.Wait()
}

func TestNewApprovalRequestID(t *testing.T) {
	id1 := NewApprovalRequestID()
	id2 := NewApprovalRequestID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
	if len(id1) < 10 {
		t.Errorf("ID too short: %q", id1)
	}
}
