# Layer Web UI smoke and CI coverage

## Goal

Split the Web UI browser smoke into focused layers so failures are easier to diagnose, and add a stable CI smoke layer without making CI slow or flaky.

## Requirements

- Smoke scripts:
  - Keep `scripts/smoke_webui.sh` as the full local Web UI smoke entrypoint.
  - Add focused layer entrypoints for:
    - core shell / diagnostics / config
    - permissions editor
    - agent editor and direct test runner
    - runs / sessions / approval
  - Share daemon setup, Playwright install, and helper code to avoid divergent scripts.
  - Each layer should emit a clear label in logs and write a layer-specific screenshot.
- CI:
  - Run only the stable core Web UI smoke layer in GitHub Actions.
  - Keep full Web UI smoke local by default.
  - Avoid adding long CI runtime.
- Compatibility:
  - Existing `scripts/smoke_webui.sh` must still pass and cover all current workflows.
  - Existing Go CI steps must remain intact.

## Acceptance Criteria

- [ ] `scripts/smoke_webui.sh` still runs full Web UI smoke.
- [ ] Focused Web UI smoke layers can be run individually.
- [ ] Shared harness avoids duplicated daemon setup logic.
- [ ] CI runs the core Web UI smoke layer.
- [ ] Full test/vet/smoke validation passes locally.

## Notes

- The main goal is maintainability and CI signal quality, not additional GUI functionality.
