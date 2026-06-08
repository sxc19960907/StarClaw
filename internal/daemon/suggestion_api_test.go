package daemon

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSuggestionAPIDefaultSessionGetAcceptConsumes(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	s.suggestions.Set("sess-1", "continue implementation", time.Date(2026, 6, 8, 13, 0, 0, 0, time.UTC))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var got suggestionResponse
	getJSON(t, ts.URL+"/sessions/sess-1/suggestion", http.StatusOK, &got)
	if got.Text != "continue implementation" || got.SuggestedAtUnix == 0 {
		t.Fatalf("suggestion = %#v", got)
	}

	var accepted suggestionResponse
	postJSON(t, ts.URL+"/sessions/sess-1/suggestion/accept", `{}`, http.StatusOK, &accepted)
	if accepted.Suggestion != "continue implementation" || accepted.Text != "continue implementation" {
		t.Fatalf("accepted = %#v", accepted)
	}
	getJSON(t, ts.URL+"/sessions/sess-1/suggestion", http.StatusNotFound, &map[string]string{})
	postJSON(t, ts.URL+"/sessions/sess-1/suggestion/accept", `{}`, http.StatusNotFound, &map[string]string{})
}

func TestSuggestionAPINamedAgentRoute(t *testing.T) {
	deps := newTestServerDeps(t)
	agentDir := filepath.Join(deps.AgentsDir, "researcher")
	if err := os.MkdirAll(agentDir, 0700); err != nil {
		t.Fatalf("mkdir agent: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("Research agent\n"), 0600); err != nil {
		t.Fatalf("write agent: %v", err)
	}
	s := newTestServer(t, deps)
	s.suggestions.Set("sess-2", "check local evidence", time.Date(2026, 6, 8, 14, 0, 0, 0, time.UTC))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var got suggestionResponse
	getJSON(t, ts.URL+"/agents/researcher/sessions/sess-2/suggestion", http.StatusOK, &got)
	if got.Text != "check local evidence" {
		t.Fatalf("suggestion = %#v", got)
	}
	postJSON(t, ts.URL+"/agents/researcher/sessions/sess-2/suggestion/accept", `{}`, http.StatusOK, &suggestionResponse{})
}

func TestSuggestionAPIValidation(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	getJSON(t, ts.URL+"/sessions/bad%5Cid/suggestion", http.StatusBadRequest, &map[string]string{})
	getJSON(t, ts.URL+"/agents/missing/sessions/sess/suggestion", http.StatusNotFound, &map[string]string{})
}
