# Desktop window recovery implementation plan

## Checklist

1. Update Swift shell route handling.
   - Add `AstriaRouteStore`.
   - Persist same-origin `/app` routes as relative routes.
   - Restore persisted routes into the configured Web UI origin.
   - Reject external or non-`/app` routes.
2. Update WebView integration.
   - Notify route store after successful navigation.
   - Load restored route on startup.
   - Expose a reload token or coordinator hook for daemon recovery reload.
3. Update daemon health monitoring.
   - Poll health after `attached`.
   - Show unavailable/recovered shell banner.
   - Reload WebView after recovery.
   - Keep retry behavior for failed/crashed states.
4. Add smoke coverage.
   - Add a command-line route recovery smoke mode or extend existing smoke.
   - Verify persisted route normalization and unsafe-route fallback without
     requiring a visible signed app.
5. Update documentation/spec as needed.
   - macOS shell README.
   - backend directory structure shell scenario.
6. Validate.
   - Trellis validation.
   - macOS shell smoke.
   - targeted app/doctor Go tests.
   - full Go tests.
   - diff check.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-desktop-window-recovery`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not persist full development origins; store only relative routes.
- Do not let external WebView navigations become future restore targets.
- Do not unmount the WebView unnecessarily during short daemon health dips.
- Do not duplicate Web UI run recovery logic in Swift.
- Keep fallback route deterministic: `/app/`.
