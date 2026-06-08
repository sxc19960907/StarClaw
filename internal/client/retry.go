package client

import (
	"errors"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RetryConfig controls retry behavior for LLM API calls.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns sensible defaults for LLM API retries.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
	}
}

// IsRetryableError returns true for transient errors that may succeed on retry.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrStreamIdleTimeout) {
		return false
	}
	s := err.Error()
	if strings.Contains(s, "context canceled") || strings.Contains(s, "context deadline exceeded") {
		return false
	}
	if strings.Contains(s, "401") || strings.Contains(s, "403") {
		return false
	}
	if strings.Contains(s, "429") || strings.Contains(s, "rate limit") {
		return true
	}
	if strings.Contains(s, "500") || strings.Contains(s, "502") || strings.Contains(s, "503") || strings.Contains(s, "504") {
		return true
	}
	if strings.Contains(s, "timeout") || strings.Contains(s, "deadline exceeded") {
		return true
	}
	if strings.Contains(s, "connection reset") || strings.Contains(s, "connection refused") {
		return true
	}
	if strings.Contains(s, "EOF") || strings.Contains(s, "broken pipe") {
		return true
	}
	return false
}

// BackoffDelay calculates the delay for a given attempt with jitter.
// attempt is 0-indexed.
func BackoffDelay(attempt int, cfg RetryConfig) time.Duration {
	delay := cfg.BaseDelay * (1 << attempt)
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}
	jitter := time.Duration(rand.Int64N(int64(delay) / 2))
	return delay + jitter
}

// ParseRetryAfter extracts a delay from an HTTP response's Retry-After header.
// Returns 0 if the header is absent or unparseable.
func ParseRetryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	val := resp.Header.Get("Retry-After")
	if val == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(val); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(val); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}
