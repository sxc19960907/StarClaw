# Add release install smoke

## Goal

Add a release preflight smoke that simulates a user downloading a release archive, extracting it, and launching StarClaw from a clean home directory.

## Requirements

- Script accepts a release archive path via `RELEASE_ARCHIVE` or first argument.
- Script extracts `.tar.gz` and `.zip` archives to a temporary install directory.
- Script finds `starclaw` or `starclaw.exe` and verifies it is executable where applicable.
- Script runs `version`, `app --check`, `app --no-open`, `/version`, and `/diagnostics` with isolated `HOME`.
- Script stops the daemon and cleans temporary files.
- Release checklist documents how to run the clean install smoke after artifact validation.

## Acceptance Criteria

- [x] New smoke script exists and is executable.
- [x] Smoke works against a synthetic release archive.
- [x] Smoke validates clean install launch routes.
- [x] Release checklist includes the smoke command.
- [x] Diff check passes.
