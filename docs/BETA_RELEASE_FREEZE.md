# Astria Beta Release Freeze

Date: 2026-06-11
Candidate: v0.3.0-beta.1
Git commit: d124430
Channel: internal beta

## Artifacts

- macOS DMG: `build/release/Astria-v0.3.0-beta.1-macos-arm64.dmg`
- Bundled daemon binary: `build/release/starclaw`
- DMG SHA-256: `368a31066884f30a3a0ca46741c12925e17820618f398ef90e3a793387139356`

The local macOS DMG is unsigned and not notarized. It is suitable for internal
trial distribution only. Public distribution still requires Developer ID
signing, Hardened Runtime, notarization, stapling, and final release metadata.

## Included Product Scope

- Astria desktop shell with bundled StarClaw daemon.
- Local Web UI at `http://127.0.0.1:7533/app/`.
- User-supplied LLM provider configuration; no bundled provider credentials.
- LLM connection testing with categorized failures.
- Runtime provider switching from saved config.
- Task launch, streaming response, tool-call rendering, run history, and run
  observation.
- Desktop layout guardrails for non-maximized windows.
- Daemon ownership checks before app reuse, status, and shutdown.

## Validation

The following checks passed for this freeze:

```bash
go test ./internal/config ./internal/client ./internal/agent ./internal/daemon ./cmd
scripts/validate_release_artifacts.sh --npm-only --astria-local
scripts/smoke_app_launch.sh
scripts/smoke_webui_config.sh
scripts/smoke_webui_tool_call.sh
scripts/smoke_webui_streaming.sh
scripts/smoke_webui_desktop_layout.sh
```

The release artifact was additionally checked with:

```bash
hdiutil imageinfo build/release/Astria-v0.3.0-beta.1-macos-arm64.dmg
shasum -a 256 build/release/Astria-v0.3.0-beta.1-macos-arm64.dmg
build/release/starclaw version
build/release/starclaw app --check
```

## Known Limitations

- The DMG is unsigned and not notarized.
- Auto-update remains disabled/unavailable-safe for Astria.
- LLM Base URL, model, and API key must be provided by the user.
- Public release packaging still needs tag creation, changelog publication, and
  signed release artifacts.

## Release Decision

This candidate is cleared for internal beta testing. It is not cleared as a
public stable release until signing/notarization and final release publication
steps are completed.
