# GUI usability polish batch

## Goal

Improve the daemon Web UI's day-to-day usability with a focused batch of small features: clearer run activity, basic session operations, a permissions overview entry point, and more useful diagnostics repair actions.

## Confirmed Facts

- The embedded Web UI is static HTML/CSS/JS served by the daemon; no frontend build step exists.
- Chat already renders messages, tool calls/results, approval cards, diagnostics, config, schedules, sessions, agents, and skills.
- The session package already supports tags/favorite and persisted session titles, but daemon HTTP routes currently only list/get/delete/search sessions.
- The Web UI currently deletes sessions immediately without confirmation.
- The daemon `/permissions` endpoint currently returns an empty list even when config permissions exist.
- Diagnostics already has structured check IDs and the GUI has a Config panel for provider repair.
- Unrelated/untracked paths remain excluded: `.agents/skills/obsidian-cli/SKILL.md` and `output/playwright/daemon-webui-smoke.png`.

## Requirements

- Improve chat activity readability for tool calls, tool results, approval cards, system messages, and errors without changing agent execution semantics.
- Add basic session operations in GUI:
  - rename session title
  - toggle favorite
  - confirm delete before removal
  - show clearer active-session state
- Add daemon support needed for those session operations.
- Make `/permissions` return a useful read-only overview of configured permission policy.
- Add a GUI Permissions panel/entry point that shows allowed dirs, allowed commands, denied commands, network allowlist, and sensitive patterns when present.
- Extend Diagnostics actions so provider checks route to Config, permissions checks route to Permissions, and storage/tool checks show specific action text.
- Keep all changes same-origin and embedded; do not add a frontend build pipeline.
- Update smoke coverage for the new visible GUI features.

## Acceptance Criteria

- [ ] Session rename and favorite toggle are available through daemon API and tested.
- [ ] GUI session list supports rename, favorite toggle, and delete confirmation.
- [ ] Active session state is visible in the chat/composer area.
- [ ] `/permissions` returns real read-only policy data from loaded config.
- [ ] GUI has a Permissions panel rendering permission categories and empty states.
- [ ] Diagnostics action controls route users to Config or Permissions where applicable.
- [ ] Chat activity/tool/approval/error presentation is clearer and remains stable in smoke.
- [ ] Smoke script exercises at least session rename/favorite, permissions panel render, diagnostics action routing, and approval card rendering.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, `go vet ./...`, and `scripts/smoke_webui.sh` pass.

## Out Of Scope

- Full Agent editor.
- Full permissions rule editor or persistence through the GUI.
- Reworking the chat protocol or agent execution loop.
- New frontend framework/build tooling.
