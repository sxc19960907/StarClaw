package desktop_rpc

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestBrokerNotConnected(t *testing.T) {
	t.Parallel()
	b := NewBroker()
	res, err := b.Request(context.Background(), &RPCRequest{Method: MethodSystemPing})
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("error = %v, want ErrNotConnected", err)
	}
	if res != nil {
		t.Fatalf("result = %#v, want nil", res)
	}
}

func TestBrokerRequestResultCorrelation(t *testing.T) {
	t.Parallel()
	b := NewBroker()
	want := json.RawMessage(`{"pong":"hi"}`)
	b.SetSendFn(func(req *RPCRequest) error {
		go b.Resolve(req.RequestID, &RPCResult{RequestID: req.RequestID, OK: true, Result: want})
		return nil
	})
	res, err := b.Request(context.Background(), &RPCRequest{
		Method:    MethodSystemPing,
		Params:    json.RawMessage(`{"echo":"hi"}`),
		TimeoutMs: 1000,
	})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if !res.OK || string(res.Result) != string(want) {
		t.Fatalf("result = %#v, want %s", res, want)
	}
	if b.PendingCount() != 0 {
		t.Fatalf("pending = %d, want 0", b.PendingCount())
	}
}

func TestBrokerTimeout(t *testing.T) {
	t.Parallel()
	b := NewBroker()
	b.SetSendFn(func(req *RPCRequest) error { return nil })
	res, err := b.Request(context.Background(), &RPCRequest{Method: MethodSystemPing, TimeoutMs: 25})
	if err != nil {
		t.Fatalf("Request returned Go error: %v", err)
	}
	if res == nil || res.OK || res.Error == nil || res.Error.Code != ErrCodeTimeout {
		t.Fatalf("result = %#v, want timeout RPC error", res)
	}
	if b.PendingCount() != 0 {
		t.Fatalf("pending = %d, want 0", b.PendingCount())
	}
}

func TestBrokerCancelAllUnblocksPending(t *testing.T) {
	t.Parallel()
	b := NewBroker()
	b.SetSendFn(func(req *RPCRequest) error { return nil })

	const count = 4
	var wg sync.WaitGroup
	errs := make([]error, count)
	results := make([]*RPCResult, count)
	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		go func() {
			defer wg.Done()
			results[i], errs[i] = b.Request(context.Background(), &RPCRequest{
				Method:    MethodSystemPing,
				TimeoutMs: 5000,
			})
		}()
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && b.PendingCount() != count {
		time.Sleep(5 * time.Millisecond)
	}
	if b.PendingCount() != count {
		t.Fatalf("pending = %d, want %d", b.PendingCount(), count)
	}

	b.CancelAll()
	wg.Wait()
	for i := 0; i < count; i++ {
		if errs[i] != nil {
			t.Fatalf("request %d error = %v", i, errs[i])
		}
		if results[i] == nil || results[i].OK || results[i].Error == nil || results[i].Error.Code != ErrCodeDesktopDisconnected {
			t.Fatalf("request %d result = %#v, want desktop_disconnected", i, results[i])
		}
	}
	if b.IsConnected() {
		t.Fatal("broker still connected after CancelAll")
	}
}

func TestBrokerContextCancel(t *testing.T) {
	t.Parallel()
	b := NewBroker()
	b.SetSendFn(func(req *RPCRequest) error { return nil })
	ctx, cancel := context.WithCancel(context.Background())
	resultCh := make(chan error, 1)
	go func() {
		_, err := b.Request(ctx, &RPCRequest{Method: MethodSystemPing, TimeoutMs: 5000})
		resultCh <- err
	}()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && b.PendingCount() != 1 {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	select {
	case err := <-resultCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Request did not return after context cancel")
	}
}
