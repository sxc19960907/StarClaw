// Package images provides a provider-gated HTTP client for image generation
// and editing endpoints. StarClaw does not register image provider tools by
// default; callers must explicitly wire this client into the tool registry.
package images

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GenerateRequest struct {
	Prompt     string `json:"prompt"`
	Size       string `json:"size,omitempty"`
	Quality    string `json:"quality,omitempty"`
	N          int    `json:"n,omitempty"`
	Background string `json:"background,omitempty"`
}

type EditRequest struct {
	Prompt     string   `json:"prompt"`
	ImageURLs  []string `json:"image_urls"`
	Size       string   `json:"size,omitempty"`
	Quality    string   `json:"quality,omitempty"`
	N          int      `json:"n,omitempty"`
	Background string   `json:"background,omitempty"`
}

type Image struct {
	URL         string `json:"url"`
	Key         string `json:"key,omitempty"`
	SizeBytes   int64  `json:"size_bytes,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type GenerateResponse struct {
	Images []Image          `json:"images"`
	Model  string           `json:"model,omitempty"`
	Size   string           `json:"size,omitempty"`
	Usage  *json.RawMessage `json:"usage,omitempty"`
}

type errorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

var (
	ErrUnauthorized     = errors.New("image: unauthorized")
	ErrBadRequest       = errors.New("image: bad request")
	ErrRequestTooLarge  = errors.New("image: request too large")
	ErrEndpointNotFound = errors.New("image: endpoint not deployed")
	ErrUpstreamTimeout  = errors.New("image: upstream timeout")
	ErrContentRejected  = errors.New("image: no images returned")
	ErrServerConfig     = errors.New("image: server misconfigured")
	ErrTransient        = errors.New("image: transient")
	ErrInvalidImageURL  = errors.New("image: invalid source URL")
	ErrSourceTooLarge   = errors.New("image: source too large")
)

type Client struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	maxAttempts int
	backoff     func(attempt int) time.Duration
}

func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 600 * time.Second}
	}
	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		apiKey:      strings.TrimSpace(apiKey),
		httpClient:  httpClient,
		maxAttempts: 3,
		backoff:     defaultBackoff,
	}
}

func defaultBackoff(attempt int) time.Duration {
	if attempt <= 1 {
		return 0
	}
	delay := time.Second
	for i := 1; i < attempt-1; i++ {
		delay *= 2
	}
	return delay
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	return c.doWithRetry(ctx, func() (*GenerateResponse, error) {
		return c.attempt(ctx, "/api/v1/images/generations", req)
	})
}

func (c *Client) Edit(ctx context.Context, req EditRequest) (*GenerateResponse, error) {
	return c.doWithRetry(ctx, func() (*GenerateResponse, error) {
		return c.attempt(ctx, "/api/v1/images/edits", req)
	})
}

func (c *Client) doWithRetry(ctx context.Context, do func() (*GenerateResponse, error)) (*GenerateResponse, error) {
	var lastErr error
	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		if delay := c.backoff(attempt); delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		resp, err := do()
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !errors.Is(err, ErrTransient) {
			return nil, err
		}
	}
	return nil, lastErr
}

func (c *Client) attempt(ctx context.Context, endpoint string, req any) (*GenerateResponse, error) {
	if c == nil {
		return nil, fmt.Errorf("%w: image client is nil", ErrBadRequest)
	}
	if strings.TrimSpace(c.baseURL) == "" {
		return nil, fmt.Errorf("%w: provider endpoint is required", ErrBadRequest)
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("image: marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("image: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: network: %v", ErrTransient, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var out GenerateResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, fmt.Errorf("image: parse response: %w", err)
		}
		if len(out.Images) == 0 {
			return nil, fmt.Errorf("image: response missing images")
		}
		for i := range out.Images {
			if strings.TrimSpace(out.Images[i].URL) == "" {
				return nil, fmt.Errorf("image: response image[%d] missing url", i)
			}
		}
		return &out, nil
	}
	return nil, classifyError(resp.StatusCode, respBody)
}

func classifyError(status int, body []byte) error {
	var parsed errorBody
	_ = json.Unmarshal(body, &parsed)
	code := parsed.Error
	suffix := func() string {
		if parsed.Message != "" {
			return ": " + parsed.Message
		}
		if code == "" {
			if s := strings.TrimSpace(string(body)); s != "" {
				return ": " + s
			}
		}
		return ""
	}

	switch status {
	case http.StatusUnauthorized:
		return fmt.Errorf("%w (status %d, code %q)%s", ErrUnauthorized, status, code, suffix())
	case http.StatusBadRequest:
		if code == "invalid_image_url" {
			return fmt.Errorf("%w (status %d, code %q)%s", ErrInvalidImageURL, status, code, suffix())
		}
		return fmt.Errorf("%w (status %d, code %q)%s", ErrBadRequest, status, code, suffix())
	case http.StatusNotFound:
		return fmt.Errorf("%w (status %d)%s", ErrEndpointNotFound, status, suffix())
	case http.StatusRequestEntityTooLarge:
		if code == "source_too_large" {
			return fmt.Errorf("%w (status %d, code %q)%s", ErrSourceTooLarge, status, code, suffix())
		}
		return fmt.Errorf("%w (status %d, code %q)%s", ErrRequestTooLarge, status, code, suffix())
	case http.StatusGatewayTimeout:
		return fmt.Errorf("%w (status %d, code %q)%s", ErrUpstreamTimeout, status, code, suffix())
	case http.StatusBadGateway:
		if code == "no_images_returned" {
			return fmt.Errorf("%w (status %d, code %q)%s", ErrContentRejected, status, code, suffix())
		}
		return fmt.Errorf("%w (status %d, code %q)%s", ErrTransient, status, code, suffix())
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%w (status %d, code %q)%s", ErrTransient, status, code, suffix())
	case http.StatusInternalServerError:
		switch code {
		case "server_misconfigured", "s3_unconfigured":
			return fmt.Errorf("%w (status %d, code %q)%s", ErrServerConfig, status, code, suffix())
		default:
			return fmt.Errorf("%w (status %d, code %q)%s", ErrTransient, status, code, suffix())
		}
	default:
		if status >= 500 {
			return fmt.Errorf("%w (status %d, code %q)%s", ErrTransient, status, code, suffix())
		}
		return fmt.Errorf("%w (status %d, code %q)%s", ErrBadRequest, status, code, suffix())
	}
}
