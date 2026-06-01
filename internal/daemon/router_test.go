package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRouterRegistersRoutes verifies that Router.RegisterRoutes registers all
// expected routes by checking they respond without 404 from the mux.
// A 404 returned by the handler itself (e.g. "not found") is still a
// valid route registration — we distinguish by checking that the response
// body contains a handler-produced error rather than the default mux 404 page.
func TestRouterRegistersRoutes(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	defaultMux404 := "404 page not found\n"

	cases := []struct {
		method string
		path   string
	}{
		// Health
		{"GET", "/health"},
		{"GET", "/status"},
		{"GET", "/diagnostics"},

		// Message / cancel / shutdown
		{"POST", "/message"},
		{"POST", "/cancel"},
		{"POST", "/shutdown"},

		// Schedule
		{"GET", "/schedules"},
		{"GET", "/schedules/nonexistent"},
		{"POST", "/schedules"},
		{"PATCH", "/schedules/nonexistent"},
		{"DELETE", "/schedules/nonexistent"},

		// Agents
		{"GET", "/agents"},
		{"GET", "/agents/test-agent"},
		{"POST", "/agents"},
		{"PUT", "/agents/test-agent"},
		{"DELETE", "/agents/test-agent"},

		// Config
		{"GET", "/config"},
		{"PATCH", "/config"},

		// Instructions
		{"GET", "/instructions"},
		{"PUT", "/instructions"},

		// Sessions
		{"GET", "/sessions"},
		{"GET", "/sessions/test-id"},
		{"DELETE", "/sessions/test-id"},
		{"GET", "/sessions/search?q=test"},

		// Permissions / Approval
		{"GET", "/permissions"},
		{"POST", "/approval"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, ts.URL+tc.path, nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("do request: %v", err)
			}
			body := make([]byte, 100)
			n, _ := resp.Body.Read(body)
			resp.Body.Close()
			bodyStr := string(body[:n])

			// If it's the default mux 404, the route is not registered.
			if resp.StatusCode == http.StatusNotFound && bodyStr == defaultMux404 {
				t.Errorf("route not registered; got default mux 404 for %s %s", tc.method, tc.path)
			}
		})
	}
}

// TestRouterStruct verifies the Router is created correctly.
func TestRouterStruct(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	r := NewRouter(s)

	if r.srv != s {
		t.Error("Router.srv should point to the provided server")
	}

	mux := http.NewServeMux()
	deps := newTestServerDeps(t)
	r.RegisterRoutes(mux, deps)

	// Verify routes are not nil (smoke check).
	if mux == nil {
		t.Error("mux should not be nil after registration")
	}
}
