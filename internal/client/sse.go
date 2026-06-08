package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// SSEEvent represents a single Server-Sent Event.
type SSEEvent struct {
	Type string
	Data string
	ID   string
}

// SSEClient consumes Server-Sent Events from a daemon event stream.
type SSEClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	// reconnectDelay is the initial backoff duration for reconnection.
	reconnectDelay time.Duration
	// maxReconnectDelay caps the backoff duration.
	maxReconnectDelay time.Duration
}

// SSEConnectOptions controls reconnect and idle-watchdog behavior for SSE
// streams. The zero value preserves the legacy Connect behavior: no idle
// watchdog and unbounded reconnects on connect/read failures.
type SSEConnectOptions struct {
	IdleTimeout          time.Duration
	MaxReconnects        int
	ReconnectBackoffBase time.Duration
}

var errSSEIdleTimeout = fmt.Errorf("sse idle timeout")

// NewSSEClient creates a new SSEClient.
func NewSSEClient(baseURL, apiKey string) *SSEClient {
	return &SSEClient{
		baseURL:           baseURL,
		apiKey:            apiKey,
		httpClient:        &http.Client{Timeout: 0}, // no timeout for long-lived SSE
		reconnectDelay:    1 * time.Second,
		maxReconnectDelay: 30 * time.Second,
	}
}

// Connect connects to the given SSE endpoint and returns a channel of events.
// The channel is closed when the context is cancelled or the connection is
// permanently lost. On disconnect, the client automatically reconnects with
// exponential backoff. The url parameter can be a relative path (e.g.
// "/events") which will be resolved against the base URL.
func (c *SSEClient) Connect(ctx context.Context, url string) (<-chan SSEEvent, error) {
	return c.ConnectWithOptions(ctx, url, SSEConnectOptions{})
}

// ConnectWithOptions connects to the given SSE endpoint using explicit
// reconnect and idle-watchdog settings.
func (c *SSEClient) ConnectWithOptions(ctx context.Context, url string, opts SSEConnectOptions) (<-chan SSEEvent, error) {
	if strings.HasPrefix(url, "/") {
		url = c.baseURL + url
	}

	ch := make(chan SSEEvent, 64)

	go c.run(ctx, url, opts, ch)

	return ch, nil
}

// run connects to the SSE stream and loops, reconnecting on failure.
func (c *SSEClient) run(ctx context.Context, url string, opts SSEConnectOptions, ch chan<- SSEEvent) {
	defer close(ch)

	delay := opts.ReconnectBackoffBase
	if delay <= 0 {
		delay = c.reconnectDelay
	}
	if delay <= 0 {
		delay = time.Second
	}
	maxDelay := c.maxReconnectDelay
	if maxDelay <= 0 {
		maxDelay = 30 * time.Second
	}
	attempts := 0
	lastEventID := ""

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := c.connectOnce(ctx, url, lastEventID, opts.IdleTimeout, ch, func(id string) {
			lastEventID = id
		})
		if err == nil {
			// Clean exit (stream completed).
			return
		}

		// Check if context was cancelled during the connection attempt.
		select {
		case <-ctx.Done():
			return
		default:
		}

		if opts.MaxReconnects > 0 && attempts >= opts.MaxReconnects {
			return
		}
		attempts++

		// Wait with backoff before reconnecting.
		log.Printf("client: SSE reconnect %d after %s (last_event_id=%q): %v", attempts, delay, lastEventID, err)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
		}

		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}

// connectOnce makes a single SSE connection and processes events until the
// stream ends or an error occurs.
func (c *SSEClient) connectOnce(ctx context.Context, url string, lastEventID string, idleTimeout time.Duration, ch chan<- SSEEvent, onEventID func(string)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create SSE request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	if lastEventID != "" {
		req.Header.Set("Last-Event-ID", lastEventID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("SSE connect failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE returned %d", resp.StatusCode)
	}

	return c.readEvents(ctx, resp.Body, idleTimeout, ch, onEventID)
}

// readEvents reads SSE events from the response body and sends them to the
// channel. Returns nil when the stream ends normally or the context is done.
func (c *SSEClient) readEvents(ctx context.Context, body io.Reader, idleTimeout time.Duration, ch chan<- SSEEvent, onEventID func(string)) error {
	lineCh := make(chan sseLine, 64)
	scanCtx, scanCancel := context.WithCancel(ctx)
	defer scanCancel()
	go scanSSELines(scanCtx, body, lineCh)

	var current SSEEvent
	var idleC <-chan time.Time
	var idleTimer *time.Timer
	if idleTimeout > 0 {
		idleTimer = time.NewTimer(idleTimeout)
		defer idleTimer.Stop()
		idleC = idleTimer.C
	}

	for {
		select {
		case msg := <-lineCh:
			if msg.err != nil {
				return fmt.Errorf("SSE read error: %w", msg.err)
			}
			if msg.eof {
				if current.Type != "" || current.Data != "" {
					c.sendEvent(ctx, ch, current, onEventID)
				}
				return nil
			}
			if idleTimer != nil {
				resetSSETimer(idleTimer, idleTimeout)
			}
			line := msg.line

			// Comment lines (heartbeats).
			if strings.HasPrefix(line, ":") {
				continue
			}

			// Empty line = event boundary.
			if line == "" {
				if current.Type != "" || current.Data != "" {
					c.sendEvent(ctx, ch, current, onEventID)
					current = SSEEvent{}
				}
				continue
			}

			// Split on the first ":" to get field name and value.
			field, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			value = strings.TrimPrefix(value, " ")

			switch field {
			case "id":
				current.ID = value
			case "event":
				current.Type = value
			case "data":
				if current.Data != "" {
					current.Data += "\n" + value
				} else {
					current.Data = value
				}
			}
		case <-idleC:
			scanCancel()
			if closer, ok := body.(io.Closer); ok {
				_ = closer.Close()
			}
			return errSSEIdleTimeout
		}
	}
}

// sendEvent sends an SSEEvent to the channel, respecting context cancellation.
func (c *SSEClient) sendEvent(ctx context.Context, ch chan<- SSEEvent, evt SSEEvent, onEventID func(string)) {
	if evt.ID != "" && onEventID != nil {
		onEventID(evt.ID)
	}
	select {
	case ch <- evt:
	case <-ctx.Done():
	}
}

type sseLine struct {
	line string
	eof  bool
	err  error
}

func scanSSELines(ctx context.Context, body io.Reader, lineCh chan<- sseLine) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		select {
		case lineCh <- sseLine{line: scanner.Text()}:
		case <-ctx.Done():
			return
		}
	}
	if err := scanner.Err(); err != nil {
		select {
		case lineCh <- sseLine{err: err}:
		case <-ctx.Done():
		}
		return
	}
	select {
	case lineCh <- sseLine{eof: true}:
	case <-ctx.Done():
	}
}

func resetSSETimer(timer *time.Timer, timeout time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(timeout)
}
