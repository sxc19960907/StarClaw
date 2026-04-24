// Package update provides self-update functionality.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	RepoOwner = "starclaw"
	RepoName  = "starclaw"
	GitHubAPI = "https://api.github.com"
)

// Release represents a GitHub release.
type Release struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Asset represents a release asset.
type Asset struct {
	Name                 string `json:"name"`
	Size                 int    `json:"size"`
	BrowserDownloadURL   string `json:"browser_download_url"`
}

// VersionInfo represents parsed version information.
type VersionInfo struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion parses a semver string into VersionInfo.
func ParseVersion(v string) (*VersionInfo, error) {
	// Remove 'v' prefix if present
	v = strings.TrimPrefix(v, "v")

	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	major, _ := parseInt(parts[0])
	minor, _ := parseInt(parts[1])
	patch, _ := parseInt(parts[2])

	return &VersionInfo{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// GreaterThan returns true if v is greater than other.
func (v *VersionInfo) GreaterThan(other *VersionInfo) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

// String returns the version string.
func (v *VersionInfo) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsSemver checks if a string looks like a semver version.
func IsSemver(v string) bool {
	// Must start with optional 'v', then have numbers
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

// CheckForUpdate checks if a newer version is available.
// Returns: (release, hasUpdate, error)
func CheckForUpdate(currentVersion string) (*Release, bool, error) {
	// Skip update check for non-semver versions (e.g. "dev")
	if !IsSemver(currentVersion) {
		return nil, false, nil
	}

	current, err := ParseVersion(currentVersion)
	if err != nil {
		return nil, false, fmt.Errorf("parse current version: %w", err)
	}

	release, err := fetchLatestRelease()
	if err != nil {
		return nil, false, fmt.Errorf("fetch latest release: %w", err)
	}

	latest, err := ParseVersion(release.TagName)
	if err != nil {
		return nil, false, fmt.Errorf("parse latest version: %w", err)
	}

	if latest.GreaterThan(current) {
		return release, true, nil
	}

	return nil, false, nil
}

// fetchLatestRelease fetches the latest release from GitHub.
func fetchLatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", GitHubAPI, RepoOwner, RepoName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// DoUpdate performs the update to the latest version.
// Returns: (newVersion, error)
func DoUpdate(currentVersion string) (string, error) {
	// Reject non-semver versions
	if !IsSemver(currentVersion) {
		return currentVersion, fmt.Errorf("cannot update from non-semver version: %s", currentVersion)
	}

	release, hasUpdate, err := CheckForUpdate(currentVersion)
	if err != nil {
		return "", fmt.Errorf("check for update: %w", err)
	}
	if !hasUpdate {
		return currentVersion, fmt.Errorf("already up to date (%s)", currentVersion)
	}

	// Find appropriate asset for current platform
	asset := findAssetForPlatform(release.Assets)
	if asset == nil {
		return "", fmt.Errorf("no update available for %s", PlatformInfo())
	}

	// TODO: Download and replace binary
	// This is a placeholder - full implementation would:
	// 1. Download the asset
	// 2. Verify checksum
	// 3. Extract if archive
	// 4. Atomic replace current binary
	// 5. Verify new binary works

	return release.TagName, fmt.Errorf("update download not implemented (would download %s)", asset.Name)
}

// findAssetForPlatform finds the appropriate asset for the current platform.
func findAssetForPlatform(assets []Asset) *Asset {
	platform := PlatformInfo()
	osArch := strings.Replace(platform, "/", "_", -1)

	for _, asset := range assets {
		// Look for asset matching platform (e.g., starclaw_Darwin_arm64.tar.gz)
		if strings.Contains(asset.Name, osArch) || strings.Contains(asset.Name, runtime.GOOS) {
			return &asset
		}
	}

	return nil
}

// PlatformInfo returns the current platform string (e.g., "darwin/arm64").
func PlatformInfo() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

// UpdateCache tracks when updates were last checked.
type UpdateCache struct {
	Path          string
	LastChecked   time.Time
	LastVersion   string
}

// NewUpdateCache creates a new update cache.
func NewUpdateCache(path string) *UpdateCache {
	return &UpdateCache{Path: path}
}

// Load loads the cache from disk.
func (c *UpdateCache) Load() error {
	data, err := os.ReadFile(c.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	type cacheData struct {
		LastChecked string `json:"last_checked"`
		LastVersion string `json:"last_version"`
	}

	var cd cacheData
	if err := json.Unmarshal(data, &cd); err != nil {
		return err
	}

	c.LastChecked, _ = time.Parse(time.RFC3339, cd.LastChecked)
	c.LastVersion = cd.LastVersion

	return nil
}

// Save saves the cache to disk.
func (c *UpdateCache) Save() error {
	type cacheData struct {
		LastChecked string `json:"last_checked"`
		LastVersion string `json:"last_version"`
	}

	cd := cacheData{
		LastChecked: c.LastChecked.Format(time.RFC3339),
		LastVersion: c.LastVersion,
	}

	data, err := json.MarshalIndent(cd, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.Path, data, 0644)
}

// ShouldCheck returns true if enough time has passed since last check.
func (c *UpdateCache) ShouldCheck(interval time.Duration) bool {
	if c.LastChecked.IsZero() {
		return true
	}
	return time.Since(c.LastChecked) > interval
}

// Record records a check with the given version.
func (c *UpdateCache) Record(version string) {
	c.LastChecked = time.Now()
	c.LastVersion = version
}

// AutoUpdate performs a background-safe update check.
// Returns a user-facing message (empty if nothing to report).
func AutoUpdate(currentVersion, cacheDir string) string {
	// Skip for dev builds
	if !IsSemver(currentVersion) {
		return ""
	}

	// Load cache
	cachePath := filepath.Join(cacheDir, "update-check.json")
	cache := NewUpdateCache(cachePath)
	_ = cache.Load()

	// Check if we should check (24 hour interval)
	if !cache.ShouldCheck(24 * time.Hour) {
		return ""
	}

	release, found, err := CheckForUpdate(currentVersion)
	if err != nil {
		// Record check to avoid hammering API on errors
		cache.Record(currentVersion)
		_ = cache.Save()
		return ""
	}
	if !found {
		cache.Record(currentVersion)
		_ = cache.Save()
		return ""
	}

	cache.Record(release.TagName)
	_ = cache.Save()

	return fmt.Sprintf("Update available: %s — run 'starclaw update' to install", release.TagName)
}

// DownloadRelease downloads a release asset to the specified path.
func DownloadRelease(ctx context.Context, asset *Asset, targetPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", asset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	// Create temp file first for atomic write
	tmpPath := targetPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	f.Close()

	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Atomic rename
	return os.Rename(tmpPath, targetPath)
}
