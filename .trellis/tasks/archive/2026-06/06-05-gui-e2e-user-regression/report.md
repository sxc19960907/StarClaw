# GUI End-to-End User Regression Report

## Automated Baseline

- `scripts/smoke_webui_core.sh`: passed.
- `scripts/smoke_webui_tool_call.sh`: passed.

## Visual Review Artifacts

- `output/playwright/daemon-webui-core-smoke.png`
- `output/playwright/daemon-webui-tool_call-smoke.png`
- `output/playwright/daemon-webui-core-smoke.log`
- `output/playwright/daemon-webui-tool_call-smoke.log`

## Reviewed Flows

- Diagnostics launch readiness and runtime context.
- Provider setup repair path.
- Agent editor create/edit/import/export/command editor path through existing core smoke.
- Agent Test result and run/session actions.
- Chat prompt submission, approval card state, run summary, and session persistence.
- Runs list and Run Detail including grouped tool event, copy result, copy prompt, copy summary, re-run, and open session.
- Settings Version page readiness/update context through existing core smoke.
- Targeted real tool-call path using the fake OpenAI provider and `version` tool.

## Findings

- No blocking GUI regression found.
- Layout is usable at the smoke viewport.
- Run Detail right column wraps long run/session IDs, but remains readable and does not block actions.
- Tool result copy and final response display are visible and functional in the targeted tool-call path.

## Follow-Up

- Optional future polish: provide a wider Run Detail layout or responsive drawer for very long IDs/results. This is not blocking.
