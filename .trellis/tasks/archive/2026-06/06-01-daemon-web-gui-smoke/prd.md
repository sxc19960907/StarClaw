# Run daemon Web GUI smoke

## Goal

Add and run a repeatable smoke workflow that validates the real daemon-hosted Web UI at `/app/` through a browser, covering static rendering, daemon API wiring, SSE event handling, and approval-card behavior.

## Confirmed Facts

- `starclaw daemon start` serves the embedded Web UI at `http://127.0.0.1:7533/app/`.
- The repository has `scripts/smoke_cli.sh`, but no repeatable Web GUI smoke script.
- Playwright CLI can be used through the bundled Codex skill wrapper, and `npx` is available on this machine.
- The Web UI can be smoke-tested without a real LLM by using daemon APIs/events directly for static render, status, schedules, sessions, and approval cards.
- The unrelated untracked file `.agents/skills/obsidian-cli/SKILL.md` must remain untouched.

## Requirements

- Add a repeatable script under `scripts/` for daemon Web GUI smoke testing.
- Build a local StarClaw binary into a temporary directory.
- Use an isolated `HOME` so smoke state does not mutate the developer's real StarClaw config.
- Start the real daemon process and stop it reliably on exit.
- Verify `/health`, `/status`, `/`, `/app`, `/app/`, and Web UI assets through the real daemon.
- Use a real browser automation path to open `/app/` and assert the UI renders expected controls.
- Exercise UI-backed API workflows where possible without real LLM credentials:
  - schedules list/create/toggle/delete
  - approval card rendering from daemon `/events`
  - approval allow/deny button path to `/approval`
- Produce a screenshot artifact under `output/playwright/`.
- Document the smoke command in user-facing docs.

## Acceptance Criteria

- [ ] `scripts/smoke_webui.sh` exists and is executable.
- [ ] The script starts and stops a real daemon process using an isolated HOME.
- [ ] The script verifies the Web UI renders in a real browser and captures a screenshot.
- [ ] The script validates schedule CRUD through the Web UI.
- [ ] The script validates approval-card display and decision UI through `/events` + `/approval`.
- [ ] The script can run without a real LLM provider/API key.
- [ ] Docs mention the Web UI smoke command.
- [ ] `scripts/smoke_webui.sh`, `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, and `go vet ./...` pass.

## Out Of Scope

- Full LLM-backed chat completion smoke against a real provider.
- Adding this smoke script to CI by default.
- Changing daemon port configuration.
