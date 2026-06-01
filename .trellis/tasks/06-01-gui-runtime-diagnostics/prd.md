# Add GUI runtime diagnostics

## Goal

Give the daemon Web UI a clear runtime-readiness view so users can understand whether StarClaw is ready to run an agent, needs setup, or has a provider/config/filesystem problem before they submit chat prompts.

## Confirmed Facts

- The daemon already exposes `/health`, `/status`, and `/config`.
- The Web UI currently shows daemon uptime/version/active request count but not setup readiness.
- `config.Load()` creates a default `~/.starclaw/config.yaml` when missing, so diagnostics must distinguish "file exists but not configured" from "loaded successfully".
- `config.NeedsSetup(cfg)` is already used by CLI entrypoints to detect missing provider credentials.
- The daemon has access to `ServerDeps` including `StarclawDir`, `ConfigPath`, loaded `Config`, `Registry`, and `ScheduleManager`.
- Existing smoke scripts use isolated HOME and no real LLM provider.
- The unrelated untracked `.agents/skills/obsidian-cli/SKILL.md` and `output/playwright/daemon-webui-smoke.png` must remain untouched.

## Requirements

- Add `GET /diagnostics` to the daemon.
- Return an overall status: `ready`, `needs_setup`, `warning`, or `error`.
- Return structured checks with id, label, status, detail, and optional action.
- Cover at minimum:
  - config file exists
  - provider configured enough to run (`anthropic`, `openai`, `ollama`)
  - provider endpoint/model fields are present
  - StarClaw directory is writable
  - sessions directory is writable or creatable
  - schedules persistence is available
  - tool registry has tools
  - permissions config presence
- For `ollama`, perform a short best-effort reachability check against the configured endpoint without requiring a chat completion.
- Avoid network calls to paid remote LLM providers in diagnostics.
- Add Web UI display for diagnostics in the existing Codex-like workbench, without adding a landing page.
- Make diagnostics visible from the first screen via compact status and a detail panel.
- Update smoke coverage to assert diagnostics are visible.

## Acceptance Criteria

- [ ] `GET /diagnostics` is registered and tested.
- [ ] Diagnostics returns `needs_setup` when provider credentials/model setup are incomplete.
- [ ] Diagnostics returns check details for config, provider, storage, tools, and permissions.
- [ ] Web UI loads diagnostics and shows a compact runtime state in the topbar.
- [ ] Web UI provides a diagnostics detail view/panel with actionable text.
- [ ] `scripts/smoke_webui.sh` verifies diagnostics render in the browser.
- [ ] `scripts/smoke_webui.sh`, `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, and `go vet ./...` pass.

## Out Of Scope

- Editing config from the diagnostics panel.
- Running paid/provider chat probes.
- Solving remote authentication errors beyond static config checks.
