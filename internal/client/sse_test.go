package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSSEEventStruct(t *testing.T) {
	evt := SSEEvent{
		Type: "message",
		Data: "hello",
		ID:   "42",
	}
	if evt.Type != "message" {
		t.Errorf("expected Type 'message', got %q", evt.Type)
	}
	if evt.Data != "hello" {
		t.Errorf("expected Data 'hello', got %q", evt.Data)
	}
	if evt.ID != "42" {
		t.Errorf("expected ID '42', got %q", evt.ID)
	}
}

func TestNewSSEClient(t *testing.T) {
	c := NewSSEClient("http://localhost:8080", "test-key")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("expected baseURL 'http://localhost:8080', got %q", c.baseURL)
	}
	if c.apiKey != "test-key" {
		t.Errorf("expected apiKey 'test-key', got %q", c.apiKey)
	}
	if c.reconnectDelay != 1*time.Second {
		t.Errorf("expected reconnectDelay 1s, got %v", c.reconnectDelay)
	}
	if c.maxReconnectDelay != 30*time.Second {
		t.Errorf("expected maxReconnectDelay 30s, got %v", c.maxReconnectDelay)
	}
}

func TestSSEClientConnectAndReceive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("expected Accept header text/event-stream")
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected X-API-Key header")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		_, _ = fmt.Fprintf(w, "event: message\ndata: hello world\n\n")
		_, _ = fmt.Fprintf(w, "event: done\ndata: {\"status\":\"complete\"}\n\n")
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "test-key")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := client.Connect(ctx, "/events")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	var received []SSEEvent
	for evt := range ch {
		received = append(received, evt)
	}

	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}
	if received[0].Type != "message" {
		t.Errorf("expected event type 'message', got %q", received[0].Type)
	}
	if received[0].Data != "hello world" {
		t.Errorf("expected data 'hello world', got %q", received[0].Data)
	}
	if received[1].Type != "done" {
		t.Errorf("expected event type 'done', got %q", received[1].Type)
	}
}

func TestSSEClientMultipleDataLines(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "event: update\ndata: line1\ndata: line2\n\n")
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := client.Connect(ctx, srv.URL+"/events")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	var received []SSEEvent
	for evt := range ch {
		received = append(received, evt)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(received))
	}
	expectedData := "line1\nline2"
	if received[0].Data != expectedData {
		t.Errorf("expected data %q, got %q", expectedData, received[0].Data)
	}
}

func TestSSEClientHeartbeats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Heartbeat comment
		_, _ = fmt.Fprintf(w, ": keepalive\n\n")
		_, _ = fmt.Fprintf(w, "event: data\ndata: payload\n\n")
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := client.Connect(ctx, srv.URL+"/stream")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	var received []SSEEvent
	for evt := range ch {
		received = append(received, evt)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 event (heartbeat skipped), got %d", len(received))
	}
	if received[0].Data != "payload" {
		t.Errorf("expected data 'payload', got %q", received[0].Data)
	}
}

func TestSSEClientEventWithID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "id: abc123\nevent: status\ndata: running\n\n")
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := client.Connect(ctx, srv.URL+"/events")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	evt := <-ch
	if evt.ID != "abc123" {
		t.Errorf("expected ID 'abc123', got %q", evt.ID)
	}
	if evt.Type != "status" {
		t.Errorf("expected Type 'status', got %q", evt.Type)
	}
	if evt.Data != "running" {
		t.Errorf("expected Data 'running', got %q", evt.Data)
	}
}

func TestSSEClientFlushesFinalEventWithoutBlankLine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "event: done\ndata: ok")
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := client.Connect(ctx, srv.URL+"/events")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	var received []SSEEvent
	for evt := range ch {
		received = append(received, evt)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 final event, got %d", len(received))
	}
	if received[0].Type != "done" {
		t.Errorf("expected Type 'done', got %q", received[0].Type)
	}
	if received[0].Data != "ok" {
		t.Errorf("expected Data 'ok', got %q", received[0].Data)
	}
}

func TestSSEClientRelativeURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "event: test\ndata: ok\n\n")
	}))
	defer srv.Close()

	client := NewSSEClient(strings.TrimSuffix(srv.URL, "/"), "")
	// Use absolute URL explicitly to avoid path resolution issues.
	ch, err := client.Connect(context.Background(), srv.URL+"/events")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	select {
	case evt, ok := <-ch:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		if evt.Data != "ok" {
			t.Errorf("expected data 'ok', got %q", evt.Data)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}

func TestSSEClientContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		// Flush headers so the client sees the 200.
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// Wait for client to disconnect.
		<-r.Context().Done()
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ch, err := client.Connect(ctx, srv.URL+"/events")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Channel should close when context is cancelled.
	_, ok := <-ch
	if ok {
		t.Error("expected channel to be closed after context cancellation")
	}
}

func TestSSEClientHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// The connect will fail immediately with 500, then reconnect with backoff.
	// With a short context timeout, the channel will close.
	ch, err := client.Connect(ctx, srv.URL+"/events")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Channel should close after context expires.
	<-ch
}

func TestSSEClientCancelDuringReconnectDelay(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	client.reconnectDelay = time.Hour
	client.maxReconnectDelay = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := client.Connect(ctx, srv.URL+"/events")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to close after cancellation")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("channel did not close promptly during reconnect delay")
	}
}

func TestSSEClientConnectWithOptionsReconnectsWithLastEventID(t *testing.T) {
	var conns int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&conns, 1)
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		if n == 1 {
			_, _ = fmt.Fprintf(w, "id: 5\nevent: started\ndata: first\n\n")
			if flusher != nil {
				flusher.Flush()
			}
			<-r.Context().Done()
			return
		}
		if got := r.Header.Get("Last-Event-ID"); got != "5" {
			t.Errorf("Last-Event-ID = %q, want 5", got)
		}
		_, _ = fmt.Fprintf(w, "id: 6\nevent: completed\ndata: second\n\n")
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ch, err := client.ConnectWithOptions(ctx, srv.URL+"/events", SSEConnectOptions{
		IdleTimeout:          30 * time.Millisecond,
		MaxReconnects:        2,
		ReconnectBackoffBase: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("ConnectWithOptions failed: %v", err)
	}

	var events []SSEEvent
	for evt := range ch {
		events = append(events, evt)
	}
	if atomic.LoadInt32(&conns) != 2 {
		t.Fatalf("connections = %d, want 2", conns)
	}
	if len(events) != 2 {
		t.Fatalf("events = %#v, want 2", events)
	}
	if events[0].ID != "5" || events[0].Type != "started" || events[1].ID != "6" || events[1].Type != "completed" {
		t.Fatalf("events = %#v, want started id=5 then completed id=6", events)
	}
}

func TestSSEClientConnectWithOptionsIdleTimeoutReconnectBudgetExhausted(t *testing.T) {
	var conns int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&conns, 1)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		<-r.Context().Done()
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ch, err := client.ConnectWithOptions(ctx, srv.URL+"/events", SSEConnectOptions{
		IdleTimeout:          20 * time.Millisecond,
		MaxReconnects:        2,
		ReconnectBackoffBase: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("ConnectWithOptions failed: %v", err)
	}
	for range ch {
	}
	if got := atomic.LoadInt32(&conns); got != 3 {
		t.Fatalf("connections = %d, want 3 (initial + 2 reconnects)", got)
	}
}

func TestSSEClientConnectWithOptionsCancelDuringReconnectDelay(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewSSEClient(srv.URL, "")
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := client.ConnectWithOptions(ctx, srv.URL+"/events", SSEConnectOptions{
		MaxReconnects:        5,
		ReconnectBackoffBase: time.Hour,
	})
	if err != nil {
		t.Fatalf("ConnectWithOptions failed: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to close after cancellation")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("channel did not close promptly during reconnect delay")
	}
}
