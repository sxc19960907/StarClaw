package desktop_rpc

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListenerSystemMethodsAndOutgoingRPC(t *testing.T) {
	t.Parallel()
	sockPath, broker, cleanup := startTestListener(t)
	defer cleanup()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	req := &RPCRequest{
		RequestID: "desktop_ping",
		Method:    MethodSystemPing,
		Params:    json.RawMessage(`{"echo":"hello"}`),
	}
	frame, err := EncodeRequestFrame(req)
	if err != nil {
		t.Fatalf("EncodeRequestFrame: %v", err)
	}
	if err := WriteFrame(conn, frame); err != nil {
		t.Fatalf("write ping: %v", err)
	}
	res := readRPCResult(t, reader)
	if !res.OK || res.RequestID != req.RequestID {
		t.Fatalf("ping result = %#v", res)
	}
	var pong SystemPingResult
	if err := json.Unmarshal(res.Result, &pong); err != nil {
		t.Fatalf("decode pong: %v", err)
	}
	if pong.Pong != "hello" || pong.ServerTime == "" {
		t.Fatalf("pong = %#v", pong)
	}

	capsReq := &RPCRequest{RequestID: "desktop_caps", Method: MethodSystemCapabilities}
	capsFrame, _ := EncodeRequestFrame(capsReq)
	if err := WriteFrame(conn, capsFrame); err != nil {
		t.Fatalf("write caps: %v", err)
	}
	capsRes := readRPCResult(t, reader)
	if !capsRes.OK {
		t.Fatalf("caps result = %#v", capsRes)
	}
	var caps SystemCapabilitiesResult
	if err := json.Unmarshal(capsRes.Result, &caps); err != nil {
		t.Fatalf("decode caps: %v", err)
	}
	if caps.Version != ProtocolVersion || caps.Platform.AppVersion != "test-version" {
		t.Fatalf("caps = %#v", caps)
	}
	if len(caps.Methods) != len(ProtocolMethods) {
		t.Fatalf("methods = %#v, want %#v", caps.Methods, ProtocolMethods)
	}

	waitForConnected(t, broker)
	done := make(chan *RPCResult, 1)
	go func() {
		got, err := broker.Request(context.Background(), &RPCRequest{
			Method:    "native.future_method",
			Params:    json.RawMessage(`{"x":1}`),
			TimeoutMs: 1000,
		})
		if err != nil {
			t.Errorf("broker request: %v", err)
			return
		}
		done <- got
	}()
	outgoing, err := ReadFrame(reader)
	if err != nil {
		t.Fatalf("read outgoing request: %v", err)
	}
	if outgoing.Type != FrameDesktopRPCRequest {
		t.Fatalf("outgoing type = %q", outgoing.Type)
	}
	var outgoingReq RPCRequest
	if err := json.Unmarshal(outgoing.Payload, &outgoingReq); err != nil {
		t.Fatalf("decode outgoing request: %v", err)
	}
	resultFrame, _ := EncodeResultFrame(&RPCResult{
		RequestID: outgoingReq.RequestID,
		OK:        true,
		Result:    json.RawMessage(`{"ok":true}`),
	})
	if err := WriteFrame(conn, resultFrame); err != nil {
		t.Fatalf("write outgoing result: %v", err)
	}
	select {
	case got := <-done:
		if !got.OK || string(got.Result) != `{"ok":true}` {
			t.Fatalf("outgoing result = %#v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("broker request did not resolve")
	}
}

func TestListenerDisconnectCancelsPending(t *testing.T) {
	t.Parallel()
	sockPath, broker, cleanup := startTestListener(t)
	defer cleanup()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	waitForConnected(t, broker)
	resultCh := make(chan *RPCResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := broker.Request(context.Background(), &RPCRequest{Method: "native.wait", TimeoutMs: 5000})
		resultCh <- res
		errCh <- err
	}()
	reader := bufio.NewReader(conn)
	if _, err := ReadFrame(reader); err != nil {
		t.Fatalf("fake desktop read request: %v", err)
	}
	_ = conn.Close()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("request err = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("request did not return after disconnect")
	}
	res := <-resultCh
	if res == nil || res.OK || res.Error == nil || res.Error.Code != ErrCodeDesktopDisconnected {
		t.Fatalf("result = %#v, want desktop_disconnected", res)
	}
}

func TestListenerDispatchesDesktopEvents(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("/tmp", "starclaw-drpc-event")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	sockPath := filepath.Join(dir, "daemon.sock")
	readyCh := make(chan struct{})
	events := make(chan *DesktopEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener := NewListener(ListenerConfig{
		SockPath:    sockPath,
		PidfilePath: filepath.Join(dir, "daemon.pid"),
		Platform:    Platform{OS: "test-os", AppVersion: "test-version"},
		Broker:      NewBroker(),
		EventSink: func(evt *DesktopEvent) {
			events <- evt
		},
		ReadyCh: readyCh,
	})
	done := make(chan error, 1)
	go func() {
		done <- listener.Run(ctx)
	}()
	select {
	case <-readyCh:
	case err := <-done:
		t.Fatalf("listener exited before ready: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("listener did not become ready")
	}

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	payload, err := json.Marshal(&DesktopEvent{Event: EventDesktopOnline, TS: "2026-06-09T01:00:00Z"})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	if err := WriteFrame(conn, &Frame{Type: FrameDesktopEvent, Payload: payload}); err != nil {
		t.Fatalf("write event: %v", err)
	}

	select {
	case evt := <-events:
		if evt.Event != EventDesktopOnline || evt.TS != "2026-06-09T01:00:00Z" {
			t.Fatalf("event = %#v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("event sink was not called")
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("listener shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("listener did not shut down")
	}
}

func startTestListener(t *testing.T) (string, *Broker, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "starclaw-drpc")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	sockPath := filepath.Join(dir, "daemon.sock")
	broker := NewBroker()
	readyCh := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	listener := NewListener(ListenerConfig{
		SockPath:    sockPath,
		PidfilePath: filepath.Join(dir, "daemon.pid"),
		Platform:    Platform{OS: "test-os", AppVersion: "test-version"},
		Broker:      broker,
		ReadyCh:     readyCh,
	})
	done := make(chan error, 1)
	go func() {
		done <- listener.Run(ctx)
	}()
	select {
	case <-readyCh:
	case err := <-done:
		cancel()
		t.Fatalf("listener exited before ready: %v", err)
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("listener did not become ready")
	}
	return sockPath, broker, func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("listener shutdown: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("listener did not shut down")
		}
		_ = os.RemoveAll(dir)
	}
}

func readRPCResult(t *testing.T, reader *bufio.Reader) RPCResult {
	t.Helper()
	frame, err := ReadFrame(reader)
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	if frame.Type != FrameDesktopRPCResult {
		t.Fatalf("frame type = %q", frame.Type)
	}
	var res RPCResult
	if err := json.Unmarshal(frame.Payload, &res); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return res
}

func waitForConnected(t *testing.T, broker *Broker) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if broker.IsConnected() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("broker did not become connected")
}
