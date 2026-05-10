package mcp

import (
	"context"
	"testing"
	"time"
)

func TestNewReadinessChecker(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	if rc == nil {
		t.Fatal("NewReadinessChecker returned nil")
	}
	if rc.manager != cm {
		t.Error("ReadinessChecker should use the provided ClientManager")
	}
}

func TestReadinessChecker_IsReady_NotConnected(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	if rc.IsReady("unknown-server") {
		t.Error("IsReady should return false for unknown server")
	}
}

func TestReadinessChecker_IsReady_Connected(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	// Directly inject a mock client
	cm.clients["test-server"] = &mockMCPClient{}

	if !rc.IsReady("test-server") {
		t.Error("IsReady should return true for connected server")
	}
}

func TestReadinessChecker_IsReady_AfterDisconnect(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	cm.clients["test-server"] = &mockMCPClient{}
	if !rc.IsReady("test-server") {
		t.Error("IsReady should return true for connected server")
	}

	cm.Disconnect("test-server")
	if rc.IsReady("test-server") {
		t.Error("IsReady should return false after disconnect")
	}
}

func TestReadinessChecker_WaitForReady_AlreadyReady(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	cm.clients["ready-server"] = &mockMCPClient{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := rc.WaitForReady(ctx, "ready-server")
	if err != nil {
		t.Errorf("WaitForReady should return immediately if already ready: %v", err)
	}
}

func TestReadinessChecker_WaitForReady_BecomesReady(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// In another goroutine, make the server ready after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cm.mu.Lock()
			cm.clients["slow-server"] = &mockMCPClient{}
			cm.mu.Unlock()
	}()

	err := rc.WaitForReady(ctx, "slow-server")
	if err != nil {
		t.Errorf("WaitForReady should succeed when server becomes ready: %v", err)
	}
}

func TestReadinessChecker_WaitForReady_Timeout(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := rc.WaitForReady(ctx, "never-ready")
	if err == nil {
		t.Error("WaitForReady should return error on timeout")
	}
}

func TestReadinessChecker_WaitForReady_Cancel(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	err := rc.WaitForReady(ctx, "cancelled-server")
	if err == nil {
		t.Error("WaitForReady should return error when context is cancelled")
	}
}

func TestReadinessChecker_MultipleServers(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	cm.clients["server-a"] = &mockMCPClient{}

	if !rc.IsReady("server-a") {
		t.Error("server-a should be ready")
	}
	if rc.IsReady("server-b") {
		t.Error("server-b should not be ready")
	}

	cm.clients["server-b"] = &mockMCPClient{}

	if !rc.IsReady("server-b") {
		t.Error("server-b should now be ready")
	}
}

func TestReadinessChecker_ConcurrentAccess(t *testing.T) {
	cm := NewClientManager()
	rc := NewReadinessChecker(cm)

	// Seed with mock clients
	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		cm.clients[name] = &mockMCPClient{}
	}

	// Concurrent IsReady calls should not panic
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			_ = rc.IsReady("a")
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cm.Disconnect("a")
			cm.mu.Lock()
			cm.clients["a"] = &mockMCPClient{}
			cm.mu.Unlock()
		}
		done <- struct{}{}
	}()

	<-done
	<-done
	// Should not panic or deadlock
}
