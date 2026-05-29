// Package update provides update check functionality.
package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// RepoOwner is the GitHub owner for releases.
const RepoOwner = "sxc19960907"

// RepoName is the GitHub repo name for releases.
const RepoName = "StarClaw"

// GitHubAPI is the base URL for the GitHub API.
var GitHubAPI = "https://api.github.com"

var renameFile = os.Rename

// Release represents a GitHub release.
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	HTMLURL     string  `json:"html_url"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	Size               int    `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
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

	asset := findAssetForPlatform(release.Assets)
	if asset == nil {
		return "", fmt.Errorf("no update available for %s", PlatformInfo())
	}

	checksums := findAssetByName(release.Assets, "checksums.txt")
	if checksums == nil {
		return "", fmt.Errorf("release %s does not include checksums.txt", release.TagName)
	}

	exePath, err := currentExecutablePath()
	if err != nil {
		return "", fmt.Errorf("resolve current executable: %w", err)
	}

	if err := installReleaseAsset(context.Background(), asset, checksums, exePath); err != nil {
		return release.TagName, err
	}

	return release.TagName, nil
}

// findAssetForPlatform finds the appropriate asset for the current platform.
func findAssetForPlatform(assets []Asset) *Asset {
	return findAssetByName(assets, platformAssetName(runtime.GOOS, runtime.GOARCH))
}

func findAssetByName(assets []Asset, name string) *Asset {
	for _, asset := range assets {
		if asset.Name == name {
			return &asset
		}
	}
	return nil
}

func platformAssetName(goos, goarch string) string {
	arch := goarch
	if goarch == "amd64" {
		arch = "x86_64"
	}

	switch goos {
	case "darwin":
		return fmt.Sprintf("starclaw_Darwin_%s.tar.gz", arch)
	case "linux":
		return fmt.Sprintf("starclaw_Linux_%s.tar.gz", arch)
	case "windows":
		return fmt.Sprintf("starclaw_Windows_%s.zip", arch)
	default:
		return fmt.Sprintf("starclaw_%s_%s", goos, arch)
	}
}

// PlatformInfo returns the current platform string (e.g., "darwin/arm64").
func PlatformInfo() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

// UpdateCache tracks when updates were last checked.
type UpdateCache struct {
	Path        string
	LastChecked time.Time
	LastVersion string
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

	return fmt.Sprintf("Update available: %s — run 'starclaw update --check' for details", release.TagName)
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
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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
	return renameFile(tmpPath, targetPath)
}

func installReleaseAsset(ctx context.Context, asset, checksums *Asset, exePath string) error {
	tempDir, err := os.MkdirTemp("", "starclaw-update-*")
	if err != nil {
		return fmt.Errorf("create update temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, asset.Name)
	if err := DownloadRelease(ctx, asset, archivePath); err != nil {
		return fmt.Errorf("download %s: %w", asset.Name, err)
	}

	checksumPath := filepath.Join(tempDir, checksums.Name)
	if err := DownloadRelease(ctx, checksums, checksumPath); err != nil {
		return fmt.Errorf("download checksums: %w", err)
	}

	if err := verifyChecksumFile(archivePath, asset.Name, checksumPath); err != nil {
		return fmt.Errorf("verify checksum: %w", err)
	}

	extractedPath, err := extractArchive(archivePath, tempDir)
	if err != nil {
		return fmt.Errorf("extract %s: %w", asset.Name, err)
	}

	if err := replaceExecutable(extractedPath, exePath); err != nil {
		return fmt.Errorf("replace executable: %w", err)
	}

	return nil
}

func currentExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		return resolved, nil
	}
	return exePath, nil
}

func verifyChecksumFile(filePath, assetName, checksumPath string) error {
	expected, err := expectedChecksum(assetName, checksumPath)
	if err != nil {
		return err
	}

	actual, err := sha256File(filePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func expectedChecksum(assetName, checksumPath string) (string, error) {
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return "", fmt.Errorf("read checksums: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimPrefix(fields[len(fields)-1], "*")
		if filepath.Base(name) == assetName {
			if len(fields[0]) != sha256.Size*2 {
				return "", fmt.Errorf("invalid sha256 for %s", assetName)
			}
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("checksum for %s not found", assetName)
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func extractArchive(archivePath, destDir string) (string, error) {
	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"):
		return extractTarGz(archivePath, destDir)
	case strings.HasSuffix(archivePath, ".zip"):
		return extractZip(archivePath, destDir)
	default:
		return "", fmt.Errorf("unsupported archive format: %s", filepath.Base(archivePath))
	}
}

func extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("open gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar: %w", err)
		}
		if header.Typeflag != tar.TypeReg || filepath.Base(header.Name) != "starclaw" {
			continue
		}
		target := filepath.Join(destDir, "starclaw-new")
		if err := writeExecutable(target, tr, 0755); err != nil {
			return "", err
		}
		return target, nil
	}

	return "", fmt.Errorf("starclaw binary not found in archive")
}

func extractZip(archivePath, destDir string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	defer zr.Close()

	for _, file := range zr.File {
		if file.FileInfo().IsDir() || filepath.Base(file.Name) != "starclaw.exe" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("open zip entry: %w", err)
		}
		target := filepath.Join(destDir, "starclaw-new.exe")
		err = writeExecutable(target, rc, 0755)
		rc.Close()
		if err != nil {
			return "", err
		}
		return target, nil
	}

	return "", fmt.Errorf("starclaw.exe not found in archive")
}

func writeExecutable(path string, r io.Reader, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create executable: %w", err)
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return fmt.Errorf("write executable: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close executable: %w", err)
	}
	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("chmod executable: %w", err)
	}
	return nil
}

func replaceExecutable(newPath, targetPath string) error {
	targetDir := filepath.Dir(targetPath)
	tempPath := filepath.Join(targetDir, "."+filepath.Base(targetPath)+".new")
	backupPath := targetPath + ".old"

	_ = os.Remove(tempPath)
	_ = os.Remove(backupPath)

	if err := copyFile(newPath, tempPath, 0755); err != nil {
		return fmt.Errorf("stage new executable: %w", err)
	}
	if err := renameFile(targetPath, backupPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("backup current executable: %w", err)
	}
	if err := renameFile(tempPath, targetPath); err != nil {
		restoreErr := renameFile(backupPath, targetPath)
		if restoreErr != nil {
			return fmt.Errorf("install new executable: %w; restore failed: %v", err, restoreErr)
		}
		return fmt.Errorf("install new executable: %w", err)
	}

	_ = os.Remove(backupPath)
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("copy file: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close destination: %w", err)
	}
	return os.Chmod(dst, mode)
}
