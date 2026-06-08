package daemon

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebUIRoutes(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	tests := []struct {
		name             string
		path             string
		wantStatus       int
		wantLocation     string
		wantContentType  string
		wantBodyContains []string
	}{
		{
			name:         "root redirects to app",
			path:         "/",
			wantStatus:   http.StatusFound,
			wantLocation: "/app/",
		},
		{
			name:         "app redirects to canonical slash",
			path:         "/app",
			wantStatus:   http.StatusFound,
			wantLocation: "/app/",
		},
		{
			name:            "app serves html shell",
			path:            "/app/",
			wantStatus:      http.StatusOK,
			wantContentType: "text/html",
			wantBodyContains: []string{
				"Astria",
				"/app/assets/styles.css",
				"/app/assets/app.js",
			},
		},
		{
			name:             "styles asset is served",
			path:             "/app/assets/styles.css",
			wantStatus:       http.StatusOK,
			wantContentType:  "text/css",
			wantBodyContains: []string{".shell", ".runtime-recovery-grid", ".runtime-badge"},
		},
		{
			name:             "script asset is served",
			path:             "/app/assets/app.js",
			wantStatus:       http.StatusOK,
			wantContentType:  "text/javascript",
			wantBodyContains: []string{"refreshAll", "renderRuntimeRecovery", "/trace", "Recovered"},
		},
		{
			name:       "unknown path remains not found",
			path:       "/missing",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(ts.URL + tt.path)
			if err != nil {
				t.Fatalf("GET %s: %v", tt.path, err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
			if tt.wantLocation != "" && resp.Header.Get("Location") != tt.wantLocation {
				t.Fatalf("expected Location %q, got %q", tt.wantLocation, resp.Header.Get("Location"))
			}
			if tt.wantContentType != "" {
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, tt.wantContentType) {
					t.Fatalf("expected Content-Type containing %q, got %q", tt.wantContentType, contentType)
				}
			}

			if len(tt.wantBodyContains) == 0 {
				return
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			bodyText := string(body)
			for _, want := range tt.wantBodyContains {
				if !strings.Contains(bodyText, want) {
					t.Fatalf("expected body to contain %q", want)
				}
			}
		})
	}
}

func TestWebUITraceAndRecoveryRenderersSanitizePayloads(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/app/assets/app.js")
	if err != nil {
		t.Fatalf("GET app.js: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	body := string(data)
	for _, want := range []string{
		"function safeRenderPayload(value)",
		"function unsafePayloadKey(key)",
		"function secretLikeValue(value)",
		"formatToolPayload(safeRenderPayload(step.metadata))",
		"formatToolPayload(safeRenderPayload(item.attributes))",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("app.js missing safe-render marker %q", want)
		}
	}
	for _, forbidden := range []string{
		"formatToolPayload(step.metadata)",
		"formatToolPayload(item.attributes)",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("app.js renders unsafe metadata path %q", forbidden)
		}
	}
}

func TestWebUIPhase5RuntimeLayoutAndNavigationHooks(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	script, err := fetchWebUIAsset(t, ts.URL+"/app/assets/app.js")
	if err != nil {
		t.Fatalf("fetch app.js: %v", err)
	}
	for _, want := range []string{
		`data-data-insight="${escapeHTML(card.id)}"`,
		`data-data-select="${escapeHTML(card.id)}"`,
		`data-data-draft="${escapeHTML(card.id)}"`,
		`event.target.closest("[data-data-select]")`,
		`event.target.closest("[data-data-draft]")`,
		`event.target.closest("[data-data-insight]")`,
		`renderRunTrace(state.currentRunTrace, state.currentRunTraceError)`,
		`Trace unavailable: ${escapeHTML(error)}`,
		`runID: latestRun?.id || ""`,
		`runID: completedRun.id || ""`,
		`data-run-open="${escapeHTML(card.runID)}"`,
		`data-run-open="${escapeHTML(asset.runID)}"`,
		`Prompt available in the explicit Prompt section.`,
		`[REDACTED: use Copy prompt for local operator review]`,
		`step.id === "runtime-recovery"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("app.js missing Phase 5 hook %q", want)
		}
	}
	for _, forbidden := range []string{
		`return status === "running" && run.id !== state.activeRequestID;`,
		`detail: prompt,`,
		"`Prompt: ${runPrompt(run) || \"-\"}`",
		`formatToolPayload(entry.data || {})`,
	} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("app.js still contains unsafe Phase 5 marker %q", forbidden)
		}
	}

	styles, err := fetchWebUIAsset(t, ts.URL+"/app/assets/styles.css")
	if err != nil {
		t.Fatalf("fetch styles.css: %v", err)
	}
	for _, want := range []string{
		"grid-template-columns: repeat(auto-fit, minmax(118px, max-content));",
		"white-space: normal;",
		"grid-template-columns: minmax(0, 1fr) max-content;",
		"overflow-wrap: anywhere;",
	} {
		if !strings.Contains(styles, want) {
			t.Fatalf("styles.css missing runtime layout guard %q", want)
		}
	}
}

func fetchWebUIAsset(t *testing.T, url string) (string, error) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status = %d, want 200", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
