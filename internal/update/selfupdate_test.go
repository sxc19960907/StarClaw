package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected VersionInfo
		wantErr  bool
	}{
		{"v1.2.3", VersionInfo{1, 2, 3}, false},
		{"1.2.3", VersionInfo{1, 2, 3}, false},
		{"0.0.1", VersionInfo{0, 0, 1}, false},
		{"10.20.30", VersionInfo{10, 20, 30}, false},
		{"v2.0.0", VersionInfo{2, 0, 0}, false},
		{"dev", VersionInfo{}, true},
		{"1.2", VersionInfo{}, true},
		{"", VersionInfo{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseVersion(%q) should return error", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseVersion(%q) returned error: %v", tt.input, err)
				return
			}
			if *v != tt.expected {
				t.Errorf("ParseVersion(%q) = %+v, want %+v", tt.input, *v, tt.expected)
			}
		})
	}
}

func TestVersionInfo_GreaterThan(t *testing.T) {
	tests := []struct {
		a        VersionInfo
		b        VersionInfo
		expected bool
	}{
		{VersionInfo{1, 2, 3}, VersionInfo{1, 2, 2}, true},
		{VersionInfo{1, 2, 3}, VersionInfo{1, 2, 3}, false},
		{VersionInfo{1, 2, 3}, VersionInfo{1, 2, 4}, false},
		{VersionInfo{1, 3, 0}, VersionInfo{1, 2, 5}, true},
		{VersionInfo{2, 0, 0}, VersionInfo{1, 5, 5}, true},
		{VersionInfo{1, 0, 0}, VersionInfo{2, 0, 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.a.String()+"_"+tt.b.String(), func(t *testing.T) {
			result := tt.a.GreaterThan(&tt.b)
			if result != tt.expected {
				t.Errorf("%v.GreaterThan(%v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestVersionInfo_String(t *testing.T) {
	v := VersionInfo{1, 2, 3}
	if v.String() != "v1.2.3" {
		t.Errorf("String() = %q, want v1.2.3", v.String())
	}
}

func TestIsSemver(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"v1.2.3", true},
		{"1.2.3", true},
		{"0.0.1", true},
		{"10.20.30", true},
		{"dev", false},
		{"1.2", false},
		{"", false},
		{"v1.2.3-beta", false},
		{"1.2.3-alpha", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsSemver(tt.input)
			if result != tt.expected {
				t.Errorf("IsSemver(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPlatformInfo(t *testing.T) {
	info := PlatformInfo()
	if info == "" {
		t.Error("PlatformInfo() returned empty string")
	}
	if !strings.Contains(info, "/") {
		t.Errorf("PlatformInfo() = %q, should contain '/'", info)
	}
}

func TestFindAssetForPlatform(t *testing.T) {
	assets := []Asset{
		{Name: "starclaw_Linux_x86_64.tar.gz"},
		{Name: "starclaw_Darwin_arm64.tar.gz"},
		{Name: "starclaw_Windows_x86_64.zip"},
		{Name: "checksums.txt"},
	}

	asset := findAssetForPlatform(assets)
	if asset == nil {
		t.Skip("No asset found for current platform")
	}
	if asset.Name == "checksums.txt" {
		t.Error("Should not match non-binary asset")
	}
}

func TestUpdateCache(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "update-cache.json")

	cache := NewUpdateCache(cachePath)

	if !cache.ShouldCheck(time.Hour) {
		t.Error("ShouldCheck should return true for empty cache")
	}

	cache.Record("v1.2.3")

	if cache.ShouldCheck(time.Hour) {
		t.Error("ShouldCheck should return false for fresh cache")
	}

	if err := cache.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	cache2 := NewUpdateCache(cachePath)
	if err := cache2.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cache2.LastVersion != "v1.2.3" {
		t.Errorf("LastVersion = %q, want v1.2.3", cache2.LastVersion)
	}
	if cache2.LastChecked.IsZero() {
		t.Error("LastChecked should not be zero after load")
	}
}

func TestUpdateCache_LoadMissing(t *testing.T) {
	cache := NewUpdateCache("/nonexistent/path/cache.json")
	if err := cache.Load(); err != nil {
		t.Errorf("Load() should not error for missing file: %v", err)
	}
}

func TestUpdateCache_LoadCorrupted(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "bad-cache.json")
	os.WriteFile(cachePath, []byte("not valid json"), 0644)

	cache := NewUpdateCache(cachePath)
	if err := cache.Load(); err == nil {
		t.Error("Load() should error for corrupted cache")
	}
}

func TestCheckForUpdate_DevBuild(t *testing.T) {
	release, hasUpdate, err := CheckForUpdate("dev")
	if err != nil {
		t.Errorf("CheckForUpdate(dev) returned error: %v", err)
	}
	if hasUpdate {
		t.Error("CheckForUpdate(dev) should not have update")
	}
	if release != nil {
		t.Error("CheckForUpdate(dev) should return nil release")
	}
}

func TestCheckForUpdate_NonSemver(t *testing.T) {
	release, hasUpdate, err := CheckForUpdate("v1.2.3-beta")
	if err != nil {
		t.Errorf("CheckForUpdate returned error: %v", err)
	}
	if hasUpdate {
		t.Error("CheckForUpdate should not have update for non-semver")
	}
	if release != nil {
		t.Error("CheckForUpdate should return nil release for non-semver")
	}
}

func TestCheckForUpdate_WithMockServer(t *testing.T) {
	// Create mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/starclaw/starclaw/releases/latest" {
			release := Release{
				TagName:     "v2.0.0",
				Name:        "Release v2.0.0",
				HTMLURL:     "https://github.com/starclaw/starclaw/releases/v2.0.0",
				PublishedAt: "2026-01-01T00:00:00Z",
				Assets: []Asset{
					{Name: "starclaw_Darwin_arm64.tar.gz", Size: 100, BrowserDownloadURL: "https://example.com/asset"},
				},
			}
			json.NewEncoder(w).Encode(release)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Override GitHubAPI temporarily
	oldAPI := GitHubAPI
	GitHubAPI = server.URL
	defer func() { GitHubAPI = oldAPI }()

	release, hasUpdate, err := CheckForUpdate("v1.0.0")
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}
	if !hasUpdate {
		t.Error("Expected update to be available (v1.0.0 → v2.0.0)")
	}
	if release == nil {
		t.Fatal("Expected non-nil release")
	}
	if release.TagName != "v2.0.0" {
		t.Errorf("TagName = %q, want v2.0.0", release.TagName)
	}
}

func TestCheckForUpdate_AlreadyLatest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v1.0.0",
			Name:    "Release v1.0.0",
			HTMLURL: "https://github.com/starclaw/starclaw/releases/v1.0.0",
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	oldAPI := GitHubAPI
	GitHubAPI = server.URL
	defer func() { GitHubAPI = oldAPI }()

	_, hasUpdate, err := CheckForUpdate("v1.0.0")
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}
	if hasUpdate {
		t.Error("Should not have update when already latest")
	}
}

func TestCheckForUpdate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	oldAPI := GitHubAPI
	GitHubAPI = server.URL
	defer func() { GitHubAPI = oldAPI }()

	_, _, err := CheckForUpdate("v1.0.0")
	if err == nil {
		t.Error("CheckForUpdate should fail on server error")
	}
}

func TestDoUpdate_NonSemver(t *testing.T) {
	version, err := DoUpdate("dev")
	if err == nil {
		t.Error("DoUpdate(dev) should return error")
	}
	if version != "dev" {
		t.Error("DoUpdate(dev) should return original version")
	}
}

func TestDoUpdate_AlreadyLatest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v1.0.0",
			Assets: []Asset{
				{Name: "starclaw_Darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/asset"},
			},
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	oldAPI := GitHubAPI
	GitHubAPI = server.URL
	defer func() { GitHubAPI = oldAPI }()

	_, err := DoUpdate("v1.0.0")
	if err == nil {
		t.Error("DoUpdate should error when already latest")
	}
}

func TestDoUpdate_HasUpdateInstallationNotImplemented(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v2.0.0",
			Assets: []Asset{
				{Name: "starclaw_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz", BrowserDownloadURL: "https://example.com/asset"},
			},
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	oldAPI := GitHubAPI
	GitHubAPI = server.URL
	defer func() { GitHubAPI = oldAPI }()

	version, err := DoUpdate("v1.0.0")
	if err == nil {
		t.Fatal("DoUpdate should error until automatic installation is implemented")
	}
	if version != "v2.0.0" {
		t.Fatalf("DoUpdate version = %q, want v2.0.0", version)
	}
	if !strings.Contains(err.Error(), "automatic update installation is not implemented yet") {
		t.Fatalf("DoUpdate error should explain installation limitation, got: %v", err)
	}
}

func TestAutoUpdate_DevBuild(t *testing.T) {
	tmpDir := t.TempDir()
	msg := AutoUpdate("dev", tmpDir)
	if msg != "" {
		t.Error("AutoUpdate(dev) should return empty message")
	}
}

func TestAutoUpdate_CacheFresh(t *testing.T) {
	tmpDir := t.TempDir()

	cache := NewUpdateCache(filepath.Join(tmpDir, "update-check.json"))
	cache.Record("v1.0.0")
	cache.Save()

	msg := AutoUpdate("v1.0.0", tmpDir)
	// Should not check because cache is fresh
	_ = msg
}

func TestAutoUpdate_HasUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v2.0.0",
			Assets: []Asset{
				{Name: "starclaw_Darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/asset"},
			},
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	oldAPI := GitHubAPI
	GitHubAPI = server.URL
	defer func() { GitHubAPI = oldAPI }()

	tmpDir := t.TempDir()
	msg := AutoUpdate("v1.0.0", tmpDir)
	if msg == "" {
		t.Error("AutoUpdate should return update message")
	}
	if !strings.Contains(msg, "v2.0.0") {
		t.Errorf("AutoUpdate message should mention new version: %q", msg)
	}
}

func TestDownloadRelease_ToTemp(t *testing.T) {
	// Create a test server serving a small file
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-binary-content"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	asset := &Asset{
		Name:               "starclaw_test",
		BrowserDownloadURL: server.URL,
	}

	targetPath := filepath.Join(tmpDir, "downloaded-binary")
	ctx := context.Background()
	err := DownloadRelease(ctx, asset, targetPath)
	if err != nil {
		t.Fatalf("DownloadRelease failed: %v", err)
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if string(data) != "fake-binary-content" {
		t.Errorf("Download content = %q, want 'fake-binary-content'", string(data))
	}
}

func TestDownloadRelease_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	asset := &Asset{
		Name:               "nonexistent",
		BrowserDownloadURL: server.URL,
	}

	err := DownloadRelease(context.Background(), asset, filepath.Join(tmpDir, "output"))
	if err == nil {
		t.Error("DownloadRelease should fail on server error")
	}
}

func TestDownloadRelease_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	asset := &Asset{
		Name:               "test",
		BrowserDownloadURL: "http://localhost:1/nonexistent",
	}

	tmpDir := t.TempDir()
	err := DownloadRelease(ctx, asset, filepath.Join(tmpDir, "output"))
	if err == nil {
		t.Error("DownloadRelease should fail with cancelled context")
	}
}

// contains is a helper for string matching in test messages.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
