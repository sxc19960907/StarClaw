# Design

## Smoke Runner

- Keep existing `WEBUI_SMOKE_MODE` wrappers as the public interface.
- Add `WEBUI_SMOKE_ARTIFACT_DIR`, defaulting to `output/playwright`.
- Copy daemon logs to the artifact directory during cleanup/failure and write metadata after paths are known.
- Preserve temporary dependency install behavior and daemon lifecycle.

## CI

- Keep CI on `scripts/smoke_webui_core.sh` to protect runtime.
- Upload the smoke artifact directory only on failure.
- Do not add full/agents/runs smoke to required CI.

## Docs

- Add a concise Web UI smoke section to the existing examples docs so contributors can choose the right layer locally.
