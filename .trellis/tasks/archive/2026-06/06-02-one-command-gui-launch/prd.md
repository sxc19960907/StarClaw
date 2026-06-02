# Add one-command GUI launch

## Goal

Make the daemon-hosted GUI feel like a local app by giving users a single command that starts the daemon when needed and opens the Web UI.

## Confirmed Facts

- The daemon Web UI is served at `http://127.0.0.1:7533/app/`.
- Existing commands require two steps: `starclaw daemon start` in one terminal and `starclaw daemon open` in another.
- `daemon open` currently only opens the browser and does not check daemon health.
- `daemon status` already checks `/status`; smoke scripts already validate daemon health with `/health`.
- The CLI already has a browser opener abstraction used by `daemon open`.

## Requirements

- Add a top-level `starclaw app` command that:
  - checks whether the daemon is already healthy;
  - starts the daemon in the background when it is not running;
  - waits for `/health`;
  - opens the Web UI in the default browser;
  - prints the Web UI URL and whether it started or reused the daemon.
- Add `starclaw daemon open --start` with the same ensure-and-open behavior.
- Preserve existing `starclaw daemon open` behavior when `--start` is not set.
- Keep daemon startup failures actionable.
- Add unit tests for existing-running and start-then-open flows.
- Update README/docs to promote the one-command GUI entry.

## Acceptance Criteria

- [x] `starclaw app` exists in CLI help.
- [x] `starclaw app` reuses an already healthy daemon and opens the Web UI.
- [x] `starclaw app` starts the daemon when `/health` is unavailable, waits until ready, then opens the Web UI.
- [x] `starclaw daemon open --start` uses the same behavior.
- [x] Existing `starclaw daemon open` still only opens the Web UI.
- [x] README/docs mention the one-command GUI launch.
- [x] Targeted tests, full tests, vet, and CLI smoke pass.

## Out of Scope

- System service installation.
- Desktop app packaging.
- Port selection beyond the current fixed daemon port.

## Goal

TBD.

## Requirements

- TBD

## Acceptance Criteria

- [ ] TBD

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
