# Specification: Self-Update Mechanism

## Overview

| Field | Value |
|-------|-------|
| Feature | Self-Update |
| Type | Utility Feature |
| Optional | Yes (can be disabled) |
| Default | Auto-check enabled, auto-install disabled |

## Purpose

Enable StarClaw to check for and install updates from GitHub releases automatically, improving user experience by ensuring users have the latest features and bug fixes.

## Configuration

```yaml
# ~/.starclaw/config.yaml
update:
  auto_check: true        # Check on startup
  auto_install: false     # Auto-install if update available
  channel: stable         # stable or beta
  cache_ttl: 24h          # Hours between checks
```

## Version Requirements

- Current version must be valid semver (e.g., "v1.2.3")
- Dev builds (e.g., "dev", "0.0.0") skip update checks
- Comparison uses semantic versioning

## Update Check Flow

```
1. Check if version is semver
   └─ No → Skip (dev build)
   
2. Check cache TTL
   └─ Cache fresh → Skip
   
3. Query GitHub API for latest release
   └─ GET https://api.github.com/repos/starclaw/starclaw/releases/latest
   
4. Compare versions
   └─ release > current → Update available
   
5. Cache result
```

## Update Installation Flow

```
1. Download release asset for current platform
   └─ e.g., starclaw_Darwin_arm64.tar.gz
   
2. Verify checksum from checksums.txt
   └─ Download fails → Abort
   
3. Extract binary to temp location
   
4. Atomic replace
   a. Rename current binary to .backup
   b. Move new binary to current location
   c. Remove .backup on success
   └─ Any step fails → Restore from .backup
   
5. Verify new binary works
   └─ Run `starclaw version`
   
6. Report success
```

## CLI Commands

### Manual Update

```bash
starclaw update              # Check and install if available
starclaw update --check      # Check only, don't install
```

### Output Examples

```
$ starclaw update --check
Current version: v1.2.3
Latest version: v1.3.0
Update available! Run 'starclaw update' to install.

$ starclaw update
Current version: v1.2.3
Latest version: v1.3.0
Downloading v1.3.0...
Verifying checksum... ✓
Installing... ✓
Updated to v1.3.0

$ starclaw update
Already up to date (v1.3.0)
```

## API

```go
package update

// Release info
type Release struct {
    Version string
    URL     string
    Assets  []Asset
}

type Asset struct {
    Name string
    URL  string
}

// Check for available update
// Returns: (release, hasUpdate, error)
func CheckForUpdate(currentVersion string) (*Release, bool, error)

// Perform update
// Returns: (newVersion, error)
func DoUpdate(currentVersion string) (string, error)

// Background auto-update check
// Returns user-facing message (empty if nothing to report)
func AutoUpdate(currentVersion, cacheDir string) string

// Platform info for asset selection
func PlatformInfo() string  // e.g., "darwin/arm64"
```

## Asset Naming

GitHub release assets follow pattern:

```
starclaw_<OS>_<Arch>.tar.gz
starclaw_<OS>_<Arch>.zip          # Windows
```

Examples:
- `starclaw_Darwin_arm64.tar.gz`
- `starclaw_Linux_x86_64.tar.gz`
- `starclaw_Windows_x86_64.zip`

## Checksum Verification

- Checksums stored in `checksums.txt` with release
- Format: `<sha256>  <filename>`
- Binary rejected if checksum doesn't match

## Error Handling

| Error | Behavior |
|-------|----------|
| GitHub API unavailable | Log warning, continue with current version |
| Download failure | Report error, keep current binary |
| Checksum mismatch | Abort, don't install |
| Permission denied | Report error with fix instructions |
| Atomic replace fails | Restore from backup, report error |
| New binary doesn't run | Restore from backup, report error |

## Security Considerations

- Downloads over HTTPS only
- Checksum verification required
- Atomic replacement prevents corruption
- Backup enables rollback
- No automatic installation by default (opt-in)

## Testing Requirements

### Unit Tests

- Version comparison (semver edge cases)
- Platform detection
- Cache logic

### Integration Tests

- Mock GitHub API server
- Test with fake release
- Test failure scenarios (network, permission)

### Mock GitHub Server

```go
func setupMockGitHub(t *testing.T) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Return fake release data
    }))
}
```

## Rollback

On update failure:

1. If `.backup` exists, restore it
2. Report error to user
3. Suggest manual update

## References

- go-selfupdate: https://github.com/creativeprojects/go-selfupdate
- Semantic Versioning: https://semver.org/
- ShanClaw Reference: `/Users/timmy/PycharmProjects/ShanClaw/internal/update/selfupdate.go`
