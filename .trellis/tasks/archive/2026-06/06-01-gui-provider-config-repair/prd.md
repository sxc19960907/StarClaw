# Add GUI provider config repair

## Goal

Let users fix provider setup problems directly from the daemon Web UI, especially when Diagnostics reports `needs_setup` for provider credentials, endpoint, or model settings.

## User Value

Users should not need to leave the GUI, open `~/.starclaw/config.yaml`, and hand-edit provider fields just to make the local daemon ready. Diagnostics should identify setup gaps and the GUI should provide a focused repair path for the fields that commonly block first run.

## Confirmed Facts

- The Web UI already has a Diagnostics panel backed by `GET /diagnostics`.
- `GET /diagnostics` reports provider issues as structured checks with actionable text.
- The daemon already exposes `GET /config` and `PATCH /config`.
- StarClaw's config file is YAML (`config.yaml`), but the current daemon config handler tries to parse and write JSON maps.
- `ServerDeps.Config` is the in-memory config used by diagnostics after daemon startup.
- Existing embedded GUI has no build step and must keep same-origin daemon API calls.
- Unrelated/untracked paths remain excluded: `.agents/skills/obsidian-cli/SKILL.md` and `output/playwright/daemon-webui-smoke.png`.

## Requirements

- Add a provider setup form reachable from the Web UI, preferably from Diagnostics and as a compact Config/Setup panel.
- The form must support the provider fields needed by diagnostics:
  - provider: `anthropic`, `openai`, or `ollama`
  - Anthropic: `endpoint`, `model_tier`, `api_key`
  - OpenAI: `openai_endpoint`, `openai_model`, `openai_api_key`
  - Ollama: `ollama_endpoint`, `ollama_model`
- Do not expose stored API key values in the GUI response.
- Blank API key inputs must preserve existing stored keys instead of clearing them.
- `PATCH /config` must update YAML config safely and keep `ServerDeps.Config` in sync so Diagnostics reflects changes without daemon restart.
- The GUI must refresh config and diagnostics after saving.
- The UI must render errors instead of throwing uncaught exceptions.
- Keep the embedded GUI static; do not add a Node/Vite frontend build pipeline.
- Update smoke coverage to exercise provider config repair rendering and save behavior.

## Acceptance Criteria

- [ ] `GET /config` returns structured non-secret provider config data for YAML config files.
- [ ] `PATCH /config` writes valid YAML and updates in-memory daemon config.
- [ ] API keys are never returned in plaintext by `GET /config`.
- [ ] Blank API key form values do not overwrite existing keys.
- [ ] Web UI includes a provider setup form with provider-specific fields.
- [ ] Diagnostics panel offers a direct path to the setup form.
- [ ] Saving provider config refreshes Diagnostics and shows success/error feedback.
- [ ] Backend tests cover YAML config get/patch and secret redaction/preservation.
- [ ] Smoke script verifies the provider setup UI renders and can save a local Ollama config.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, `go vet ./...`, and `scripts/smoke_webui.sh` pass.

## Out Of Scope

- Full arbitrary config editing.
- Editing permissions, hooks, MCP servers, audit settings, or schedule config.
- Validating remote paid provider credentials with live API calls.
- Clearing an existing API key from the GUI; users can still edit config.yaml manually for that edge case.
