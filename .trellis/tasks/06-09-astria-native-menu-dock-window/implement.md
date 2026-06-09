# Astria native menu Dock and window shell implementation plan

## Checklist

1. Add `AstriaNativeCommandSpec` and `--native-command-smoke`.
2. Add `AstriaAppActions` observable object for reload/diagnostics/retry hooks.
3. Wire root view to register reload/diagnostics/retry closures.
4. Replace empty `.newItem` command group with New Window command.
5. Add Astria command menu entries for Reload, Diagnostics, Retry Daemon.
6. Update shell smoke script and macOS shell spec.
7. Validate:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-native-menu-dock-window`
   - `scripts/smoke_macos_astria_shell.sh`
   - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
   - `go test ./...`
   - `git diff --check`

## Risk Points

- Avoid retaining stale root view closures after windows close.
- Do not block command creation on daemon health.
- Do not change WebView URL persistence semantics.
