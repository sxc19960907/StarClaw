# Design

## Architecture

The update package owns release discovery, asset selection, download, checksum verification, archive extraction, and binary replacement. The CLI stays thin: it prints status, calls `update.DoUpdate`, and reports the installed version.

## Data Flow

1. `cmd/root.go` calls `update.DoUpdate(Version)`.
2. `DoUpdate` calls `CheckForUpdate`.
3. The latest release is scanned for:
   - an exact platform archive asset;
   - `checksums.txt`.
4. The archive and checksums file are downloaded into a temporary directory.
5. The archive SHA-256 is compared with the expected checksum line.
6. The archive is extracted into the temporary directory.
7. The extracted executable replaces the current executable.
8. `DoUpdate` returns the release tag.

## Asset Matching

Use exact GoReleaser names for the current platform:

- `starclaw_Darwin_x86_64.tar.gz` for `darwin/amd64`
- `starclaw_Darwin_arm64.tar.gz` for `darwin/arm64`
- `starclaw_Linux_x86_64.tar.gz` for `linux/amd64`
- `starclaw_Linux_arm64.tar.gz` for `linux/arm64`
- `starclaw_Windows_x86_64.zip` for `windows/amd64`
- `starclaw_Windows_arm64.zip` for `windows/arm64`

This avoids the current fuzzy matching where a Linux build may match any Linux asset regardless of architecture.

## Replacement Strategy

Resolve the current executable path with `os.Executable()` and `filepath.EvalSymlinks` where possible.

On Unix:

- write the new executable to a sibling temporary file;
- set mode `0755`;
- rename the current executable to a `.old` backup;
- rename the new executable into place;
- remove the backup after success;
- if the final rename fails, attempt to restore the backup.

On Windows:

- replacing the currently running executable may fail because the file can be locked;
- use the same backup-and-rename strategy and surface a clear error if the OS refuses replacement;
- tests should exercise the replace helper through injected paths rather than relying on Windows-only locking behavior.

## Testability

Keep small helpers for:

- expected platform asset name;
- checksum parsing and verification;
- archive extraction;
- executable replacement;
- update installation from an explicit release and executable path.

`DoUpdate` can use these helpers with the real current executable path. Tests can use temporary files and mock HTTP servers.

## Rollback

If any phase before replacement fails, the current binary is untouched. If replacement fails after moving the old binary aside, the helper attempts to restore the backup and returns the original error with restore context if needed.
