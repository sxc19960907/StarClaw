package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
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
	if strings.HasPrefix(url, "/") {
		url = c.baseURL + url
	}

	ch := make(chan SSEEvent, 64)

	go c.run(ctx, url, ch)

	return ch, nil
}

// run connects to the SSE stream and loops, reconnecting on failure.
func (c *SSEClient) run(ctx context.Context, url string, ch chan<- SSEEvent) {
	defer close(ch)

	delay := c.reconnectDelay

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := c.connectOnce(ctx, url, ch)
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

		// Wait with backoff before reconnecting.
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		delay *= 2
		if delay > c.maxReconnectDelay {
			delay = c.maxReconnectDelay
		}
	}
}

// connectOnce makes a single SSE connection and processes events until the
// stream ends or an error occurs.
func (c *SSEClient) connectOnce(ctx context.Context, url string, ch chan<- SSEEvent) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create SSE request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("SSE connect failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE returned %d", resp.StatusCode)
	}

	return c.readEvents(ctx, resp.Body, ch)
}

// readEvents reads SSE events from the response body and sends them to the
// channel. Returns nil when the stream ends normally or the context is done.
func (c *SSEClient) readEvents(ctx context.Context, body io.Reader, ch chan<- SSEEvent) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var current SSEEvent

	for scanner.Scan() {
		line := scanner.Text()

		// Comment lines (heartbeats).
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Empty line = event boundary.
		if line == "" {
			if current.Type != "" || current.Data != "" {
				c.sendEvent(ctx, ch, current)
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
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("SSE read error: %w", err)
	}

	return nil
}

// sendEvent sends an SSEEvent to the channel, respecting context cancellation.
func (c *SSEClient) sendEvent(ctx context.Context, ch chan<- SSEEvent, evt SSEEvent) {
	select {
	case ch <- evt:
	case <-ctx.Done():
	}
}
