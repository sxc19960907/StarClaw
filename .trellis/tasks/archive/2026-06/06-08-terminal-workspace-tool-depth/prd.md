# Terminal workspace tool depth

## Goal

Close the Phase 9 terminal workspace gap against local Kocoro commit `74cdb3c` by adding a local, approval-gated terminal workspace helper to StarClaw. The tool should be Ghostty-compatible on macOS when Ghostty is installed, while remaining explicit and safe when unavailable or on other platforms.

## Requirements

- Use `/Users/timmy/PycharmProjects/Kocoro/internal/tools/ghostty*.go` as the parity reference.
- Add a StarClaw local tool that supports visible terminal workspace actions:
  - `new_tab`
  - `new_split`
  - `send_input`
  - `list_tabs`
  - `status`
- Keep the tool local-only. Do not add remote terminal, SSH, cloud shell, external sync, or credentialed behavior.
- Require approval for terminal-affecting calls.
- Treat `list_tabs` and `status` as read-only calls; all other actions are not read-only.
- On macOS, use Ghostty when installed and at least the minimum supported version.
- When Ghostty is unavailable, return clear fallback guidance that points to existing local tools such as `applescript` or `bash`.
- On non-macOS platforms, compile cleanly and report explicit unsupported status.
- Register the tool in the default local tool registry.
- Keep StarClaw naming in code and user-facing text.

## Acceptance Criteria

- [ ] `terminal_workspace` appears in the local tool registry.
- [ ] The tool validates JSON arguments and action-specific required fields.
- [ ] `status` reports availability, platform, minimum version, and fallback guidance without approval-sensitive side effects.
- [ ] `list_tabs` reports tracked tabs without requiring Ghostty availability.
- [ ] `new_tab` and `new_split` track tab titles after successful Ghostty operations.
- [ ] `send_input` requires a known target and non-empty text.
- [ ] macOS implementation uses Ghostty via AppleScript only when Ghostty is available.
- [ ] non-darwin implementation compiles and returns unsupported availability.
- [ ] Unit tests cover registration, argument validation, read-only classification, availability fallback, tab tracking, and version comparison behavior.
- [ ] `go test ./internal/tools` passes.

## Notes

- Parent task: `.trellis/tasks/06-08-astria-kocoro-parity-phase-9`
- Reference research: `.trellis/research/kocoro-native-tool-parity-phase9-plan.md`
