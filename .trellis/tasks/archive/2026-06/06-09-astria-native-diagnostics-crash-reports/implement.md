# Astria native diagnostics export and crash reports implementation plan

## Checklist

1. Add `AstriaDiagnosticsReport` and redaction helpers.
2. Add export path management under the Astria runtime/app-support boundary.
3. Add `Export Diagnostics` native command and action bridge.
4. Add `--diagnostics-export-smoke`.
5. Update shell smoke script and macOS shell spec.
6. Validate:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-native-diagnostics-crash-reports`
   - `scripts/smoke_macos_astria_shell.sh`
   - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
   - `go test ./...`
   - `git diff --check`

## Risk Points

- Do not export raw prompts, Desktop RPC event payloads, API keys, bearer
  tokens, socket paths, or pidfile paths.
- Do not send the report anywhere automatically.
- Keep diagnostics URL opening behavior unchanged.
