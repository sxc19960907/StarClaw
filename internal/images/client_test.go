package images

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newImageTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := NewClient(srv.URL, "sk_test", srv.Client())
	c.backoff = func(int) time.Duration { return 0 }
	return c, srv
}

func TestGeneratePostsJSONWithAPIKey(t *testing.T) {
	var gotPath, gotKey, gotType string
	var gotBody []byte
	c, srv := newImageTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("X-API-Key")
		gotType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"images":[{"url":"https://cdn.example/a.png","content_type":"image/png","size_bytes":12}],"model":"img","size":"1024x1024"}`))
	}))
	defer srv.Close()

	res, err := c.Generate(context.Background(), GenerateRequest{Prompt: "a cat", Quality: "low"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if gotPath != "/api/v1/images/generations" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotKey != "sk_test" {
		t.Fatalf("api key header = %q", gotKey)
	}
	if !strings.HasPrefix(gotType, "application/json") {
		t.Fatalf("content type = %q", gotType)
	}
	var body map[string]any
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatalf("body JSON: %v", err)
	}
	if body["prompt"] != "a cat" || body["quality"] != "low" {
		t.Fatalf("body = %s", gotBody)
	}
	if len(res.Images) != 1 || res.Images[0].URL != "https://cdn.example/a.png" {
		t.Fatalf("response = %+v", res)
	}
}

func TestEditPostsToEditEndpoint(t *testing.T) {
	var gotPath string
	c, srv := newImageTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"images":[{"url":"https://cdn.example/edit.png"}]}`))
	}))
	defer srv.Close()

	if _, err := c.Edit(context.Background(), EditRequest{Prompt: "add hat", ImageURLs: []string{"https://cdn.example/src.png"}}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if gotPath != "/api/v1/images/edits" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestClientDoesNotSendEmptyAPIKey(t *testing.T) {
	var gotKey string
	c, srv := newImageTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"images":[{"url":"https://cdn.example/a.png"}]}`))
	}))
	defer srv.Close()
	c.apiKey = ""

	if _, err := c.Generate(context.Background(), GenerateRequest{Prompt: "x"}); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if gotKey != "" {
		t.Fatalf("unexpected API key header %q", gotKey)
	}
}

func TestClientRejectsMissingEndpoint(t *testing.T) {
	c := NewClient("", "", nil)
	_, err := c.Generate(context.Background(), GenerateRequest{Prompt: "x"})
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestImageClientErrorClassificationNoRetry(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"unauthorized", http.StatusUnauthorized, `{"error":"unauthorized"}`, ErrUnauthorized},
		{"bad request", http.StatusBadRequest, `{"error":"prompt_too_long"}`, ErrBadRequest},
		{"invalid image url", http.StatusBadRequest, `{"error":"invalid_image_url"}`, ErrInvalidImageURL},
		{"not found", http.StatusNotFound, `not found`, ErrEndpointNotFound},
		{"too large", http.StatusRequestEntityTooLarge, `{"error":"request_too_large"}`, ErrRequestTooLarge},
		{"source too large", http.StatusRequestEntityTooLarge, `{"error":"source_too_large"}`, ErrSourceTooLarge},
		{"timeout", http.StatusGatewayTimeout, `{"error":"upstream_timeout"}`, ErrUpstreamTimeout},
		{"content rejected", http.StatusBadGateway, `{"error":"no_images_returned"}`, ErrContentRejected},
		{"server config", http.StatusInternalServerError, `{"error":"server_misconfigured"}`, ErrServerConfig},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var calls int32
			c, srv := newImageTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&calls, 1)
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			_, err := c.Generate(context.Background(), GenerateRequest{Prompt: "x"})
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
			if got := atomic.LoadInt32(&calls); got != 1 {
				t.Fatalf("calls = %d, want 1", got)
			}
		})
	}
}

func TestImageClientRetriesTransient(t *testing.T) {
	var calls int32
	c, srv := newImageTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"busy"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"images":[{"url":"https://cdn.example/ok.png"}]}`))
	}))
	defer srv.Close()

	res, err := c.Generate(context.Background(), GenerateRequest{Prompt: "x"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
	if res.Images[0].URL != "https://cdn.example/ok.png" {
		t.Fatalf("response = %+v", res)
	}
}

func TestImageClientSuccessValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"invalid json", `not json`},
		{"missing images", `{}`},
		{"missing url", `{"images":[{}]}`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c, srv := newImageTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			if _, err := c.Generate(context.Background(), GenerateRequest{Prompt: "x"}); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
