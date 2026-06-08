package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/client"
)

type testPauseController struct {
	mu       sync.Mutex
	paused   bool
	resumeCh chan struct{}
	entered  chan struct{}
}

func newTestPauseController(paused bool) *testPauseController {
	return &testPauseController{
		paused:   paused,
		resumeCh: make(chan struct{}),
		entered:  make(chan struct{}, 1),
	}
}

func (p *testPauseController) WaitIfPaused(ctx context.Context) error {
	p.mu.Lock()
	if !p.paused {
		p.mu.Unlock()
		return nil
	}
	resumeCh := p.resumeCh
	p.mu.Unlock()

	select {
	case p.entered <- struct{}{}:
	default:
	}

	select {
	case <-resumeCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *testPauseController) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.paused {
		return
	}
	p.paused = false
	close(p.resumeCh)
	p.resumeCh = make(chan struct{})
}

type pauseCountingClient struct {
	calls chan struct{}
}

func (c *pauseCountingClient) Chat(ctx context.Context, _ string, _ []client.Message, _ []client.ToolDef, _ int, _ *client.ChatOptions) (*client.Response, error) {
	select {
	case c.calls <- struct{}{}:
	default:
	}
	return &client.Response{Content: "done"}, nil
}

func TestAgentLoop_PauseBlocksBeforeModelCallAndResumes(t *testing.T) {
	pause := newTestPauseController(true)
	llm := &pauseCountingClient{calls: make(chan struct{}, 1)}
	loop := NewAgentLoop(llm, NewToolRegistry())
	loop.SetPauseController(pause)

	done := make(chan error, 1)
	go func() {
		_, err := loop.Run(context.Background(), "hello")
		done <- err
	}()

	select {
	case <-pause.entered:
	case <-time.After(time.Second):
		t.Fatal("agent loop did not enter pause wait")
	}
	select {
	case <-llm.calls:
		t.Fatal("LLM was called while paused")
	default:
	}

	pause.Resume()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("agent loop did not resume")
	}
	select {
	case <-llm.calls:
	default:
		t.Fatal("LLM was not called after resume")
	}
}

func TestAgentLoop_PauseReturnsContextErrorOnCancel(t *testing.T) {
	pause := newTestPauseController(true)
	llm := &pauseCountingClient{calls: make(chan struct{}, 1)}
	loop := NewAgentLoop(llm, NewToolRegistry())
	loop.SetPauseController(pause)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		_, err := loop.Run(ctx, "hello")
		done <- err
	}()

	select {
	case <-pause.entered:
	case <-time.After(time.Second):
		t.Fatal("agent loop did not enter pause wait")
	}
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected cancellation error")
		}
	case <-time.After(time.Second):
		t.Fatal("agent loop did not exit after cancellation")
	}
	select {
	case <-llm.calls:
		t.Fatal("LLM was called after cancellation while paused")
	default:
	}
}
