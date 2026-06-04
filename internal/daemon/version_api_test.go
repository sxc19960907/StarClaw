package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestHandleVersionDevelopmentBuild(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/version")
	if err != nil {
		t.Fatalf("GET /version: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
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
	if body.HealthURL != "http://127.0.0.1:0/health" {
		t.Fatalf("health_url = %q, want port-specific URL", body.HealthURL)
	}
	if body.StatusURL != "http://127.0.0.1:0/status" {
		t.Fatalf("status_url = %q, want port-specific URL", body.StatusURL)
	}
	if body.DiagnosticsURL != "http://127.0.0.1:0/diagnostics" {
		t.Fatalf("diagnostics_url = %q, want port-specific URL", body.DiagnosticsURL)
	}
	if body.LaunchCommand != "starclaw app" {
		t.Fatalf("launch_command = %q, want starclaw app", body.LaunchCommand)
	}
	if body.StarclawDir != deps.StarclawDir {
		t.Fatalf("starclaw_dir = %q, want %q", body.StarclawDir, deps.StarclawDir)
	}
	if body.ConfigPath != deps.ConfigPath {
		t.Fatalf("config_path = %q, want %q", body.ConfigPath, deps.ConfigPath)
	}
	if filepath.Dir(body.ConfigPath) == "" {
		t.Fatalf("config_path = %q should include a directory", body.ConfigPath)
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
	defer func() {
		_ = resp.Body.Close()
	}()
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
