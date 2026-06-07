# Mission Control Run Board

## Goal

Upgrade the Runs panel into an Astria Mission Control board that makes active, failed, and completed work easier to scan and resume.

## Requirements

- Use existing `/runs` data.
- Add status overview cards for active, attention, completed, and total runs.
- Add quick filters for all, active, attention, completed, and council handoffs.
- Keep existing run detail, copy, rerun, open session, and timeline behavior.
- Keep the UI dense and operational, not a marketing dashboard.

## Acceptance Criteria

- [x] Runs panel shows Mission Control summary cards.
- [x] User can filter runs by status groups.
- [x] Selecting a filtered run still opens the existing run detail.
- [x] Core smoke verifies a Mission Control filter.
- [x] JS syntax, daemon tests, and Web UI smoke pass.

## Non-Goals

- No new backend run endpoint.
- No live queue scheduler changes.
- No persistence for selected filter.
