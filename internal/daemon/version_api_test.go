package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleVersionDevelopmentBuild(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/version")
	if err != nil {
		t.Fatalf("GET /version: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Version != "test-version" {
		t.Fatalf("version = %q, want test-version", body.Version)
	}
	if body.UpdateSupported {
		t.Fatal("development test version should not support update checks")
	}
	if body.Status != "development" {
		t.Fatalf("status = %q, want development", body.Status)
	}
	if body.Platform == "" {
		t.Fatal("platform should be present")
	}
	if body.WebURL != "http://127.0.0.1:0/app/" {
		t.Fatalf("web_url = %q, want port-specific URL", body.WebURL)
	}
	if body.LaunchCommand != "starclaw app" {
		t.Fatalf("launch_command = %q, want starclaw app", body.LaunchCommand)
	}
}

func TestHandleUpdateCheckDevelopmentBuildSkipsNetwork(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/update/check")
	if err != nil {
		t.Fatalf("GET /update/check: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body updateCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != "development" {
		t.Fatalf("status = %q, want development", body.Status)
	}
	if body.UpdateSupported {
		t.Fatal("development test version should not support update checks")
	}
	if body.LatestVersion != "" {
		t.Fatalf("latest_version = %q, want empty", body.LatestVersion)
	}
}
