# Implementation Plan

## Scope

Address the high-priority security findings in the `internal/tools` path and publishing surface:

- `tools/safe_path.go`: symlink escape and sibling-prefix containment.
- `tools/imaging.go`: missing path expansion / safety validation.
- `tools/screenshot.go`: missing output path validation.
- `tools/publish_to_web.go`: missing approval requirement.
- `tools/grep.go`: missing `--` separator for ripgrep invocation.

## Checklist

1. Inspect current implementations and tests for the scoped files.
2. Update shared path helpers first; prefer one fix that benefits all tools.
3. Update scoped tools to call the shared helpers consistently.
4. Add regression tests:
   - sibling-prefix containment rejection
   - symlink escape rejection
   - publish tool approval
   - grep command separator behavior where testable
   - imaging/screenshot path validation behavior
5. Run focused tests:
   - `go test ./internal/tools`
6. Run full suite:
   - `go test ./...`

## Risk Notes

- Keep path validation behavior compatible with existing valid relative paths.
- Avoid introducing platform-specific behavior that breaks Linux/macOS/Windows tests.
- Some tools rely on external commands; test command construction/validation without requiring unavailable binaries where possible.
