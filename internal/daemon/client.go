package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/schedule"
)

// defaultTimeout is applied when the caller's context has no deadline.
const defaultTimeout = 30 * time.Second

// Client is an HTTP client for communicating with the daemon server.
// All methods construct HTTP requests to {baseURL}/{endpoint}, handle JSON
// marshaling/unmarshaling, and return errors for non-2xx responses.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Client with the given base URL.
// The baseURL should be the scheme + host + optional port, e.g. "http://localhost:8080".
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

// request performs an HTTP request and decodes the response. If the response
// status is non-2xx, an error containing the response body is returned.
func (c *Client) request(ctx context.Context, method, path string, body, result interface{}) error {
	// Enforce a default timeout when the caller has not set one.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body: %w", err)
		}
	}

	return nil
}

// Health checks whether the daemon server is reachable and healthy.
func (c *Client) Health(ctx context.Context) error {
	return c.request(ctx, http.MethodGet, "/health", nil, nil)
}

// Status returns the current daemon server status.
func (c *Client) Status(ctx context.Context) (*StatusResponse, error) {
	var status StatusResponse
	if err := c.request(ctx, http.MethodGet, "/status", nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// Message sends a RunAgentRequest and returns the RunAgentResponse.
func (c *Client) Message(ctx context.Context, req RunAgentRequest) (*RunAgentResponse, error) {
	var resp RunAgentResponse
	if err := c.request(ctx, http.MethodPost, "/message", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Cancel cancels a running agent execution identified by session or request ID.
func (c *Client) Cancel(ctx context.Context, req CancelRequest) error {
	return c.request(ctx, http.MethodPost, "/cancel", req, nil)
}

// Shutdown requests the daemon server to gracefully shut down.
func (c *Client) Shutdown(ctx context.Context) error {
	return c.request(ctx, http.MethodPost, "/shutdown", nil, nil)
}

// ListSchedules returns all scheduled tasks.
func (c *Client) ListSchedules(ctx context.Context) ([]schedule.Schedule, error) {
	var schedules []schedule.Schedule
	if err := c.request(ctx, http.MethodGet, "/schedules", nil, &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

// GetSchedule returns a single scheduled task by ID.
func (c *Client) GetSchedule(ctx context.Context, id string) (*schedule.Schedule, error) {
	var s schedule.Schedule
	if err := c.request(ctx, http.MethodGet, "/schedules/"+id, nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// CreateSchedule creates a new scheduled task and returns its ID.
func (c *Client) CreateSchedule(ctx context.Context, req CreateScheduleRequest) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.request(ctx, http.MethodPost, "/schedules", req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

// PatchSchedule updates one or more fields on an existing schedule.
func (c *Client) PatchSchedule(ctx context.Context, id string, req PatchScheduleRequest) error {
	return c.request(ctx, http.MethodPatch, "/schedules/"+id, req, nil)
}

// DeleteSchedule removes a scheduled task by ID.
func (c *Client) DeleteSchedule(ctx context.Context, id string) error {
	return c.request(ctx, http.MethodDelete, "/schedules/"+id, nil, nil)
}
