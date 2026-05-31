package daemon

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
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

func TestDaemonApprovalRequesterAllow(t *testing.T) {
	broker := NewApprovalBroker()
	bus := NewEventBus()
	ch := bus.Subscribe("test")
	requester := NewDaemonApprovalRequester(broker, bus, ChannelHTTP, "run-123", "helper")

	done := make(chan agent.ApprovalDecision, 1)
	errCh := make(chan error, 1)
	go func() {
		decision, err := requester.RequestApproval(context.Background(), agent.ApprovalRequest{
			Tool:   "bash",
			Args:   `{"command":"touch file"}`,
			Reason: "requires approval",
		})
		done <- decision
		errCh <- err
	}()

	needed := readApprovalEvent(t, ch)
	if needed.Type != EventApprovalNeeded {
		t.Fatalf("event type = %q, want %q", needed.Type, EventApprovalNeeded)
	}
	var neededPayload ApprovalRequest
	if err := json.Unmarshal([]byte(needed.Data), &neededPayload); err != nil {
		t.Fatalf("decode approval_needed: %v", err)
	}
	if neededPayload.RequestID == "" {
		t.Fatal("approval request id is empty")
	}
	if neededPayload.ThreadID != "run-123" || neededPayload.Tool != "bash" || neededPayload.Agent != "helper" {
		t.Fatalf("unexpected approval payload: %#v", neededPayload)
	}

	broker.Resolve(ApprovalResolvedPayload{RequestID: neededPayload.RequestID, Decision: DecisionAllow})

	resolved := readApprovalEvent(t, ch)
	if resolved.Type != EventApprovalResolved {
		t.Fatalf("event type = %q, want %q", resolved.Type, EventApprovalResolved)
	}
	var resolvedPayload ApprovalResolvedPayload
	if err := json.Unmarshal([]byte(resolved.Data), &resolvedPayload); err != nil {
		t.Fatalf("decode approval_resolved: %v", err)
	}
	if resolvedPayload.Decision != DecisionAllow {
		t.Fatalf("resolved decision = %q, want allow", resolvedPayload.Decision)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("RequestApproval error: %v", err)
	}
	if decision := <-done; decision != agent.ApprovalAllow {
		t.Fatalf("decision = %q, want allow", decision)
	}
}

func TestDaemonApprovalRequesterDeny(t *testing.T) {
	broker := NewApprovalBroker()
	bus := NewEventBus()
	ch := bus.Subscribe("test")
	requester := NewDaemonApprovalRequester(broker, bus, ChannelHTTP, "run-123", "")

	done := make(chan agent.ApprovalDecision, 1)
	go func() {
		decision, _ := requester.RequestApproval(context.Background(), agent.ApprovalRequest{
			Tool:   "file_write",
			Args:   `{"path":"secret.txt"}`,
			Reason: "write operations require approval",
		})
		done <- decision
	}()

	needed := readApprovalEvent(t, ch)
	var neededPayload ApprovalRequest
	if err := json.Unmarshal([]byte(needed.Data), &neededPayload); err != nil {
		t.Fatalf("decode approval_needed: %v", err)
	}
	broker.Resolve(ApprovalResolvedPayload{RequestID: neededPayload.RequestID, Decision: DecisionDeny})
	_ = readApprovalEvent(t, ch)

	if decision := <-done; decision != agent.ApprovalDeny {
		t.Fatalf("decision = %q, want deny", decision)
	}
}

func readApprovalEvent(t *testing.T, ch <-chan Event) Event {
	t.Helper()
	select {
	case evt := <-ch:
		return evt
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for approval event")
		return Event{}
	}
}
