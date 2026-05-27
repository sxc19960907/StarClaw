# Stop scheduler align timer

## Goal

Ensure the daemon scheduler releases its initial minute-alignment timer promptly when the scheduler context is cancelled.

## Requirements

- Replace the production `time.After` used for the initial scheduler alignment wait with a stoppable timer.
- Preserve existing scheduler behavior: immediate catch-up tick, first aligned tick, then minute ticker loop.
- Keep the change local to scheduler timer lifecycle handling.
- Maintain or improve cancellation test coverage for scheduler startup.

## Acceptance Criteria

- [ ] Scheduler cancellation during the initial alignment wait stops and drains the timer when needed.
- [ ] Existing scheduler tests continue to pass.
- [ ] Full repository tests continue to pass.

## Notes

- Relevant guideline: `.trellis/spec/backend/quality-guidelines.md` requires `time.NewTimer` plus `Stop`/non-blocking drain for long waits that race context cancellation.
