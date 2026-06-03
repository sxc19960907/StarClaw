# Align app install and launch docs

## Goal

Make the public docs match the current one-command GUI launch, readiness check, and troubleshooting behavior.

## Requirements

- README quickstart should present `starclaw app` as the primary GUI startup command.
- Install docs should explain `starclaw app --check`, `starclaw app --no-open`, already-running daemon reuse, and browser-open failure fallback.
- Examples should include the current app launch/readiness workflow.
- Documentation should not claim Homebrew support is available.
- No code behavior changes in this task.

## Acceptance Criteria

- [x] README launch section reflects current `starclaw app` behavior.
- [x] docs/INSTALL.md documents launch diagnostics and troubleshooting clearly.
- [x] docs/EXAMPLES.md remains consistent with current app commands.
- [x] Documentation avoids unsupported Homebrew claims.
- [x] Diff check passes.
