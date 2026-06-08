package desktop_rpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

const DefaultTimeoutMs = 30000

type SendFn func(req *RPCRequest) error

var ErrNotConnected = errors.New("desktop_rpc: no desktop client connected")

type pending struct {
	ch chan *RPCResult
}

type Broker struct {
	mu      sync.Mutex
	pending map[string]*pending
	sendFn  SendFn
}

func NewBroker() *Broker {
	return &Broker{pending: make(map[string]*pending)}
}

func (b *Broker) SetSendFn(fn SendFn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sendFn = fn
}

func (b *Broker) IsConnected() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sendFn != nil
}

func (b *Broker) PendingCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.pending)
}

func (b *Broker) Request(ctx context.Context, req *RPCRequest) (*RPCResult, error) {
	if req == nil {
		return nil, errors.New("desktop_rpc: nil RPCRequest")
	}
	if req.Method == "" {
		return nil, errors.New("desktop_rpc: empty method")
	}
	if req.TimeoutMs <= 0 {
		req.TimeoutMs = DefaultTimeoutMs
	}
	req.RequestID = generateRequestID()

	b.mu.Lock()
	sendFn := b.sendFn
	if sendFn == nil {
		b.mu.Unlock()
		return nil, ErrNotConnected
	}
	pa := &pending{ch: make(chan *RPCResult, 1)}
	b.pending[req.RequestID] = pa
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.pending, req.RequestID)
		b.mu.Unlock()
	}()

	if err := sendFn(req); err != nil {
		return nil, fmt.Errorf("desktop_rpc: send failed: %w", err)
	}

	timer := time.NewTimer(time.Duration(req.TimeoutMs) * time.Millisecond)
	defer timer.Stop()
	select {
	case res := <-pa.ch:
		return res, nil
	case <-timer.C:
		return &RPCResult{
			RequestID: req.RequestID,
			OK:        false,
			Error: &RPCError{
				Code:      ErrCodeTimeout,
				Message:   fmt.Sprintf("RPC timed out after %dms", req.TimeoutMs),
				Retriable: false,
			},
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (b *Broker) Resolve(requestID string, res *RPCResult) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	pa, ok := b.pending[requestID]
	if !ok {
		return false
	}
	select {
	case pa.ch <- res:
	default:
	}
	delete(b.pending, requestID)
	return true
}

func (b *Broker) CancelAll() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, pa := range b.pending {
		res := &RPCResult{
			RequestID: id,
			OK:        false,
			Error: &RPCError{
				Code:      ErrCodeDesktopDisconnected,
				Message:   "Desktop disconnected",
				Retriable: false,
			},
		}
		select {
		case pa.ch <- res:
		default:
		}
		delete(b.pending, id)
	}
	b.sendFn = nil
}

func MarshalParams(v any) (json.RawMessage, error) {
	return json.Marshal(v)
}

func generateRequestID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "drpc_" + hex.EncodeToString(b[:])
}
