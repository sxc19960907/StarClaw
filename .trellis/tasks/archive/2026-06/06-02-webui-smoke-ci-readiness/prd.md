# Improve Web UI smoke CI readiness

## Goal

Make Web UI smoke tests easier to run in CI and easier to debug when they fail, without expanding CI runtime by running every browser smoke layer on every PR.

## Requirements

- CI should continue to run the lightweight Web UI core smoke, not the full browser suite.
- Web UI smoke runs should write reusable diagnostics artifacts outside the temporary working directory:
  - screenshot;
  - daemon log;
  - small metadata file describing mode and URLs.
- GitHub CI should upload those smoke artifacts when the Web UI core smoke fails.
- Documentation should explain the available Web UI smoke layers and which one CI runs.
- Existing local smoke commands must keep working.

## Acceptance Criteria

- [x] `scripts/smoke_webui_core.sh` still passes locally.
- [x] Failed Web UI smoke runs leave daemon log and metadata under the artifact directory.
- [x] GitHub CI uploads Web UI smoke artifacts on core smoke failure.
- [x] Docs list `core`, `permissions`, `agents`, `runs`, and `full` smoke entrypoints.
- [x] Diff check passes.
