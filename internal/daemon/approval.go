package daemon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// DefaultApprovalTimeout is the maximum time to wait for an approval response.
const DefaultApprovalTimeout = 5 * time.Minute

// ApprovalBroker mediates tool approval requests between the agent and the
// remote approval client. It blocks the caller until a decision arrives or
// the context is cancelled.
type ApprovalBroker struct {
	mu      sync.Mutex
	pending map[string]chan ApprovalDecision
	timeout time.Duration
}

// NewApprovalBroker creates a new ApprovalBroker with the default timeout.
func NewApprovalBroker() *ApprovalBroker {
	return &ApprovalBroker{
		pending: make(map[string]chan ApprovalDecision),
		timeout: DefaultApprovalTimeout,
	}
}

// NewApprovalBrokerWithTimeout creates a new ApprovalBroker with a custom timeout.
func NewApprovalBrokerWithTimeout(timeout time.Duration) *ApprovalBroker {
	return &ApprovalBroker{
		pending: make(map[string]chan ApprovalDecision),
		timeout: timeout,
	}
}

// WaitForApproval registers a pending approval and blocks until the request
// is resolved by a call to Resolve, the timeout expires, or ctx is cancelled.
// Returns the approval decision or an error if ctx was cancelled.
func (b *ApprovalBroker) WaitForApproval(ctx context.Context, req ApprovalRequest) (ApprovalDecision, error) {
	ch := make(chan ApprovalDecision, 1)

	b.mu.Lock()
	b.pending[req.RequestID] = ch
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.pending, req.RequestID)
		b.mu.Unlock()
	}()

	timer := time.NewTimer(b.timeout)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	select {
	case decision := <-ch:
		return decision, nil
	case <-timer.C:
		return DecisionDeny, nil
	case <-ctx.Done():
		return DecisionDeny, ctx.Err()
	}
}

// Resolve delivers a decision to a pending approval request identified by
// payload.RequestID. If the request ID is not found (already resolved, timed
// out, or unknown), the call is a no-op.
func (b *ApprovalBroker) Resolve(payload ApprovalResolvedPayload) {
	b.mu.Lock()
	ch, ok := b.pending[payload.RequestID]
	if ok {
		delete(b.pending, payload.RequestID)
	}
	b.mu.Unlock()

	if ok {
		select {
		case ch <- payload.Decision:
		default:
		}
	}
}

// NewApprovalRequestID generates a unique ID for an approval request.
func NewApprovalRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("apr_%d", time.Now().UnixNano())
	}
	return "apr_" + hex.EncodeToString(b)
}
