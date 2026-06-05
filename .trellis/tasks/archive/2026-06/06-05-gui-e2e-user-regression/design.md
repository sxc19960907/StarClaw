# Design

## Approach

This is primarily a regression/audit task. Use the deterministic local daemon and fake provider where possible, then inspect the actual browser-rendered GUI.

## Evidence Sources

- `scripts/smoke_webui_core.sh`
- `scripts/smoke_webui_tool_call.sh`
- Playwright CLI snapshots/screenshots
- Daemon logs under `output/playwright/`

## Pass/Fail Criteria

- Existing smoke scripts must still pass.
- Visual inspection should not reveal overlapping text, broken panels, hidden primary actions, or stale placeholder behavior.
- If a regression is found, implement the smallest fix and rerun the relevant smoke.

## Artifacts

Save screenshots under `output/playwright/` and record any notable findings in the final response and Trellis task.
