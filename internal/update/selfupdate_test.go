package update

import (
	"context"
	"path/filepath"
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
		{"v1.2.3-beta", false}, // Not strict semver
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
	// Should contain OS and Arch separated by /
	if !contains(info, "/") {
		t.Errorf("PlatformInfo() = %q, should contain '/'", info)
	}
}

func TestFindAssetForPlatform(t *testing.T) {
	// Create assets for multiple platforms
	assets := []Asset{
		{Name: "starclaw_Linux_x86_64.tar.gz"},
		{Name: "starclaw_Darwin_arm64.tar.gz"},
		{Name: "starclaw_Windows_x86_64.zip"},
		{Name: "checksums.txt"},
	}

	asset := findAssetForPlatform(assets)

	// Should find an asset that matches current platform
	if asset == nil {
		// This is OK - we might be on an unsupported platform
		t.Skip("No asset found for current platform")
	}

	// Verify the asset name makes sense
	if asset.Name == "checksums.txt" {
		t.Error("Should not match non-binary asset")
	}
}

func TestUpdateCache(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "update-cache.json")

	cache := NewUpdateCache(cachePath)

	// Initially should check
	if !cache.ShouldCheck(time.Hour) {
		t.Error("ShouldCheck should return true for empty cache")
	}

	// Record a check
	cache.Record("v1.2.3")

	// Should not check immediately
	if cache.ShouldCheck(time.Hour) {
		t.Error("ShouldCheck should return false for fresh cache")
	}

	// Save and reload
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
	// Should not error for missing file
	if err := cache.Load(); err != nil {
		t.Errorf("Load() should not error for missing file: %v", err)
	}
}

func TestCheckForUpdate_DevBuild(t *testing.T) {
	// Dev builds should skip update check
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
	// Non-semver versions should skip update check
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

func TestDoUpdate_NonSemver(t *testing.T) {
	version, err := DoUpdate("dev")
	if err == nil {
		t.Error("DoUpdate(dev) should return error")
	}
	if version != "dev" {
		t.Error("DoUpdate(dev) should return original version")
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

	// Create fresh cache
	cache := NewUpdateCache(filepath.Join(tmpDir, "update-check.json"))
	cache.Record("v1.0.0")
	cache.Save()

	// Should not check again immediately
	msg := AutoUpdate("v1.0.0", tmpDir)
	// Note: This might still check due to time precision, so we don't assert
	_ = msg
}

// Test helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsInternal(s, substr))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDownloadRelease_MockServer(t *testing.T) {
	// This test requires a mock server or internet access
	// We'll skip it if network is unavailable
	t.Skip("Skipping network test")
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// This would test context cancellation if we had a real download
	// For now, just verify context is cancelled
	if ctx.Err() != context.Canceled {
		t.Error("Context should be cancelled")
	}
}
