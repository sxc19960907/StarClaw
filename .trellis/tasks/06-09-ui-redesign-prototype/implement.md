# Implementation Plan After Prototype Approval

This file is intentionally for the later development phase. Do not start these changes until the prototype direction is reviewed.

## Phase 1: Token And Language Foundation

- Add deep-space design tokens to `internal/daemon/webui/assets/styles.css`.
- Establish Chinese-first typography stack and monospace telemetry stack.
- Replace generic system surfaces with dark observatory surfaces.
- Add responsive constraints for the shell, task composer, right observatory panel, and mobile nav.

## Phase 2: Navigation Reframe

- Compress visible primary navigation into five areas: `任务台`, `运行`, `产物`, `上下文`, `系统`.
- Preserve existing panel IDs and JS routes where possible.
- Move secondary panels under contextual groups instead of removing functionality.
- Update smoke tests only where user-visible labels are intentionally changed.

## Phase 3: Task Workbench

- Redesign the home panel around the command field.
- Promote current run state, approval state, and context readiness.
- Replace decorative star naming with operational labels.
- Keep existing message submission behavior intact.

## Phase 4: Run And Artifact Surfaces

- Redesign run summaries, tool events, approvals, result cards, and reusable outputs.
- Make active state, completed state, failed state, and waiting-for-approval state visually distinct.
- Ensure long Chinese and English technical strings wrap cleanly.

## Phase 5: Verification

- Run `go test ./...`.
- Run `scripts/smoke_webui_core.sh`.
- Run `scripts/smoke_webui_streaming.sh`.
- Run `scripts/smoke_webui_tool_call.sh`.
- Capture desktop and mobile screenshots for visual review.

