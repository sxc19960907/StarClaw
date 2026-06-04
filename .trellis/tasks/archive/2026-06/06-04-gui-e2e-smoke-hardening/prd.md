# Harden GUI End-to-End Smoke Coverage

## Problem

The GUI is usable through manual browser testing, but several important user paths are only partially covered by reusable smoke tests. We need a repeatable local smoke path that exercises real GUI interactions for permissions, agent editing, agent test runs, chat summaries, and run history/detail behavior when a provider is unavailable.

## Scope

- Strengthen the existing Web UI smoke coverage rather than adding new product features.
- Keep CI bounded by leaving the core smoke as the default CI layer unless the full suite remains fast enough.
- Prefer deterministic browser route mocks for successful chat/agent-test paths.
- Include a provider-unavailable/error-run path that uses the daemon and verifies Runs/detail UI behavior.

## Acceptance Criteria

- [x] `scripts/smoke_webui.sh` full mode covers permissions multiline save/preview, agent create/edit/test-run, chat run summary, and Runs/detail actions.
- [x] Error run detail is covered for provider-unavailable behavior, including prompt/result visibility, copy prompt, and re-run prefill.
- [x] The smoke remains isolated from the user's real `~/.starclaw` data.
- [x] `scripts/smoke_webui.sh` passes locally.
- [x] `scripts/smoke_webui_core.sh` passes locally.
- [x] `go test ./...` passes locally.
- [x] `go vet ./...` passes locally.

## Notes

- A real external provider is not required for this task; provider-unavailable behavior is a supported daemon state and should remain deterministic.
- Browser warnings should be checked during validation; verbose browser hints are not blockers unless they indicate user-visible breakage or accessibility regressions.
- Validation completed with `scripts/smoke_webui_runs.sh`, `scripts/smoke_webui_core.sh`, `scripts/smoke_webui.sh`, `go test ./...`, `go vet ./...`, and `git diff --check`.
