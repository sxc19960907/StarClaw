package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
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

func TestPlatformAssetName(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{"darwin", "amd64", "starclaw_Darwin_x86_64.tar.gz"},
		{"darwin", "arm64", "starclaw_Darwin_arm64.tar.gz"},
		{"linux", "amd64", "starclaw_Linux_x86_64.tar.gz"},
		{"linux", "arm64", "starclaw_Linux_arm64.tar.gz"},
		{"windows", "amd64", "starclaw_Windows_x86_64.zip"},
		{"windows", "arm64", "starclaw_Windows_arm64.zip"},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"_"+tt.goarch, func(t *testing.T) {
			if got := platformAssetName(tt.goos, tt.goarch); got != tt.want {
				t.Fatalf("platformAssetName() = %q, want %q", got, tt.want)
			}
		})
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
	if err := os.WriteFile(cachePath, []byte("not valid json"), 0644); err != nil {
		t.Fatal(err)
	}

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
		if r.URL.Path == "/repos/sxc19960907/StarClaw/releases/latest" {
			release := Release{
				TagName:     "v2.0.0",
				Name:        "Release v2.0.0",
				HTMLURL:     "https://github.com/starclaw/starclaw/releases/v2.0.0",
				PublishedAt: "2026-01-01T00:00:00Z",
				Assets: []Asset{
					{Name: "starclaw_Darwin_arm64.tar.gz", Size: 100, BrowserDownloadURL: "https://example.com/asset"},
				},
			}
			if err := json.NewEncoder(w).Encode(release); err != nil {
				t.Errorf("encode release: %v", err)
			}
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
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Errorf("encode release: %v", err)
		}
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
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Errorf("encode release: %v", err)
		}
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
	if err := cache.Save(); err != nil {
		t.Fatal(err)
	}

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
		_, _ = w.Write([]byte("fake-binary-content"))
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

func TestExpectedChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	checksumPath := filepath.Join(tmpDir, "checksums.txt")
	if err := os.WriteFile(checksumPath, []byte("abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789  starclaw_Linux_x86_64.tar.gz\n"), 0644); err != nil {
		t.Fatalf("write checksum: %v", err)
	}

	got, err := expectedChecksum("starclaw_Linux_x86_64.tar.gz", checksumPath)
	if err != nil {
		t.Fatalf("expectedChecksum failed: %v", err)
	}
	if got != "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789" {
		t.Fatalf("checksum = %q", got)
	}
}

func TestVerifyChecksumFile_Mismatch(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "starclaw_Linux_x86_64.tar.gz")
	checksumPath := filepath.Join(tmpDir, "checksums.txt")
	if err := os.WriteFile(archivePath, []byte("archive"), 0644); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	if err := os.WriteFile(checksumPath, []byte("abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789  starclaw_Linux_x86_64.tar.gz\n"), 0644); err != nil {
		t.Fatalf("write checksum: %v", err)
	}

	if err := verifyChecksumFile(archivePath, "starclaw_Linux_x86_64.tar.gz", checksumPath); err == nil {
		t.Fatal("verifyChecksumFile should fail on mismatch")
	}
}

func TestExtractArchive_TarGz(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "starclaw_Linux_x86_64.tar.gz")
	if err := writeTarGz(archivePath, "starclaw", []byte("new-binary")); err != nil {
		t.Fatalf("write tar.gz: %v", err)
	}

	extracted, err := extractArchive(archivePath, tmpDir)
	if err != nil {
		t.Fatalf("extractArchive failed: %v", err)
	}
	data, err := os.ReadFile(extracted)
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(data) != "new-binary" {
		t.Fatalf("extracted content = %q", data)
	}
}

func TestExtractArchive_Zip(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "starclaw_Windows_x86_64.zip")
	if err := writeZip(archivePath, "starclaw.exe", []byte("new-binary")); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	extracted, err := extractArchive(archivePath, tmpDir)
	if err != nil {
		t.Fatalf("extractArchive failed: %v", err)
	}
	data, err := os.ReadFile(extracted)
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(data) != "new-binary" {
		t.Fatalf("extracted content = %q", data)
	}
}

func TestReplaceExecutable_RestoresOnInstallFailure(t *testing.T) {
	tmpDir := t.TempDir()
	currentPath := filepath.Join(tmpDir, "starclaw")
	newPath := filepath.Join(tmpDir, "new-starclaw")
	if err := os.WriteFile(currentPath, []byte("old"), 0755); err != nil {
		t.Fatalf("write current: %v", err)
	}
	if err := os.WriteFile(newPath, []byte("new"), 0755); err != nil {
		t.Fatalf("write new: %v", err)
	}

	oldRename := renameFile
	renameFile = func(oldpath, newpath string) error {
		if strings.HasSuffix(oldpath, ".starclaw.new") && newpath == currentPath {
			return fmt.Errorf("injected install failure")
		}
		return oldRename(oldpath, newpath)
	}
	defer func() { renameFile = oldRename }()

	err := replaceExecutable(newPath, currentPath)
	if err == nil {
		t.Fatal("replaceExecutable should fail when staged temp path cannot be replaced")
	}
	data, readErr := os.ReadFile(currentPath)
	if readErr != nil {
		t.Fatalf("read restored executable: %v", readErr)
	}
	if string(data) != "old" {
		t.Fatalf("current executable = %q, want old", data)
	}
}

func TestInstallReleaseAsset_Success(t *testing.T) {
	tmpDir := t.TempDir()
	exePath := filepath.Join(tmpDir, "starclaw")
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}
	if err := os.WriteFile(exePath, []byte("old-binary"), 0755); err != nil {
		t.Fatalf("write exe: %v", err)
	}

	archiveName := platformAssetName(runtime.GOOS, runtime.GOARCH)
	archiveBytes := archiveForCurrentPlatform(t, []byte("new-binary"))
	checksumBytes := []byte(fmt.Sprintf("%x  %s\n", sha256.Sum256(archiveBytes), archiveName))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/" + archiveName:
			_, _ = w.Write(archiveBytes)
		case "/checksums.txt":
			_, _ = w.Write(checksumBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	asset := &Asset{Name: archiveName, BrowserDownloadURL: server.URL + "/" + archiveName}
	checksums := &Asset{Name: "checksums.txt", BrowserDownloadURL: server.URL + "/checksums.txt"}

	if err := installReleaseAsset(context.Background(), asset, checksums, exePath); err != nil {
		t.Fatalf("installReleaseAsset failed: %v", err)
	}
	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("read installed executable: %v", err)
	}
	if string(data) != "new-binary" {
		t.Fatalf("installed executable = %q, want new-binary", data)
	}
}

func TestInstallReleaseAsset_ChecksumMismatchKeepsExistingBinary(t *testing.T) {
	tmpDir := t.TempDir()
	exePath := filepath.Join(tmpDir, "starclaw")
	if err := os.WriteFile(exePath, []byte("old-binary"), 0755); err != nil {
		t.Fatalf("write exe: %v", err)
	}

	archiveName := platformAssetName(runtime.GOOS, runtime.GOARCH)
	archiveBytes := archiveForCurrentPlatform(t, []byte("new-binary"))
	checksumBytes := []byte(fmt.Sprintf("%064x  %s\n", 1, archiveName))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/" + archiveName:
			_, _ = w.Write(archiveBytes)
		case "/checksums.txt":
			_, _ = w.Write(checksumBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	asset := &Asset{Name: archiveName, BrowserDownloadURL: server.URL + "/" + archiveName}
	checksums := &Asset{Name: "checksums.txt", BrowserDownloadURL: server.URL + "/checksums.txt"}

	err := installReleaseAsset(context.Background(), asset, checksums, exePath)
	if err == nil {
		t.Fatal("installReleaseAsset should fail on checksum mismatch")
	}
	data, readErr := os.ReadFile(exePath)
	if readErr != nil {
		t.Fatalf("read existing executable: %v", readErr)
	}
	if string(data) != "old-binary" {
		t.Fatalf("existing executable = %q, want old-binary", data)
	}
}

func archiveForCurrentPlatform(t *testing.T, content []byte) []byte {
	t.Helper()
	if runtime.GOOS == "windows" {
		return zipBytes(t, "starclaw.exe", content)
	}
	return tarGzBytes(t, "starclaw", content)
}

func writeTarGz(path, name string, content []byte) error {
	return os.WriteFile(path, tarGzBytesForName(name, content), 0644)
}

func tarGzBytes(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	return tarGzBytesForName(name, content)
}

func tarGzBytesForName(name string, content []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(content))})
	_, _ = tw.Write(content)
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func writeZip(path, name string, content []byte) error {
	return os.WriteFile(path, zipBytesForName(name, content), 0644)
}

func zipBytes(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	return zipBytesForName(name, content)
}

func zipBytesForName(name string, content []byte) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create(name)
	_, _ = io.Copy(w, bytes.NewReader(content))
	_ = zw.Close()
	return buf.Bytes()
}
