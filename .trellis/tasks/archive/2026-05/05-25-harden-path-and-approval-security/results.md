# Results

## Fixed / Covered

- `internal/tools/safe_path.go`
  - Replaced raw prefix containment with `filepath.Rel` based subpath checks.
  - Resolve symlinks before checking whether a path is under CWD or home.
  - Resolve the nearest existing ancestor for not-yet-created paths so symlinked parent directories cannot bypass containment.
  - Reject paths outside CWD or home instead of allowing arbitrary non-sensitive absolute paths.

- `internal/tools/publish_to_web.go`
  - Added `IsSafePath` validation before publishing a file.
  - Existing approval requirement was already present and remains enforced.

- `internal/tools/grep.go`
  - Added `IsSafePath` validation before searching.
  - Existing ripgrep command already used `--` before the user pattern.

- `internal/tools/screenshot.go`
  - Validate explicit output paths before the platform check, making unsafe path behavior testable on non-macOS too.

- Tests
  - Added regression coverage for sibling-prefix path rejection.
  - Added regression coverage for symlink escape rejection.
  - Added publish and screenshot unsafe-path tests.
  - Updated tests that used temp files outside CWD to run inside a temp project directory under the new security contract.

## Verification

- `go test ./internal/tools` passed.
- `go test ./...` passed.
- `go test -race ./internal/tools` was attempted. It failed on an existing `ProcessTool` stdout/stderr buffer race in `internal/tools/process.go`, unrelated to this path-hardening task and already aligned with the next concurrency/lifecycle workstream.
