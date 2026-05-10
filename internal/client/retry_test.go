package client

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil", nil, false},
		{"rate limit 429", fmt.Errorf("API error (429): rate limited"), true},
		{"server error 500", fmt.Errorf("API error (500): internal"), true},
		{"server error 502", fmt.Errorf("API error (502): bad gateway"), true},
		{"server error 503", fmt.Errorf("API error (503): unavailable"), true},
		{"timeout", fmt.Errorf("request timeout"), true},
		{"connection reset", fmt.Errorf("connection reset by peer"), true},
		{"EOF", fmt.Errorf("unexpected EOF"), true},
		{"auth 401", fmt.Errorf("API error (401): unauthorized"), false},
		{"forbidden 403", fmt.Errorf("API error (403): forbidden"), false},
		{"context canceled", fmt.Errorf("context canceled"), false},
		{"bad request", fmt.Errorf("API error (400): bad request"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryableError(tt.err)
			if got != tt.expected {
				t.Errorf("IsRetryableError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestBackoffDelay(t *testing.T) {
	cfg := DefaultRetryConfig()

	d0 := BackoffDelay(0, cfg)
	if d0 < cfg.BaseDelay || d0 > cfg.BaseDelay*2 {
		t.Errorf("attempt 0: delay %v out of expected range [%v, %v]", d0, cfg.BaseDelay, cfg.BaseDelay*2)
	}

	d1 := BackoffDelay(1, cfg)
	if d1 < 2*cfg.BaseDelay || d1 > 3*cfg.BaseDelay {
		t.Errorf("attempt 1: delay %v out of expected range [%v, %v]", d1, 2*cfg.BaseDelay, 3*cfg.BaseDelay)
	}

	// Should cap at MaxDelay + jitter
	d10 := BackoffDelay(10, cfg)
	if d10 > cfg.MaxDelay*2 {
		t.Errorf("attempt 10: delay %v exceeds 2x max %v", d10, cfg.MaxDelay)
	}
}

func TestParseRetryAfter_Seconds(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Retry-After", "5")

	d := ParseRetryAfter(resp)
	if d != 5*time.Second {
		t.Errorf("ParseRetryAfter = %v, want 5s", d)
	}
}

func TestParseRetryAfter_Missing(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	d := ParseRetryAfter(resp)
	if d != 0 {
		t.Errorf("ParseRetryAfter = %v, want 0", d)
	}
}

func TestParseRetryAfter_Nil(t *testing.T) {
	d := ParseRetryAfter(nil)
	if d != 0 {
		t.Errorf("ParseRetryAfter(nil) = %v, want 0", d)
	}
}
