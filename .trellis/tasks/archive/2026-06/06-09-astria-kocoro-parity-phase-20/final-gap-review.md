# Phase20 final gap review

## Completed scope

Phase20 closed the production updater transaction-safety planning slice while
keeping Astria no-replacement and credential-free for local validation.

- `astria-staged-updater-transaction-plan` added a local `plan_only`
  transaction planner that combines updater metadata and compatibility manifest
  inputs, requires rollback and post-update health gate declarations, and
  rejects replacement-enabled metadata.
- `astria-updater-rollback-health-gates` added credential-free rollback and
  post-update health gate manifest validation for rollback source/target,
  daemon compatibility guard, manual approval, app launch, daemon health,
  Desktop RPC capabilities, and Web UI readiness.
- `astria-release-acceptance-gates` added production release acceptance metadata
  validation for Developer ID signing, Hardened Runtime, notarization,
  stapling, updater metadata, compatibility, rollback/health gates,
  transaction plan references, local credential-free validation, and private
  material absence.

## Validation evidence

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-release-acceptance-gates`
- `scripts/validate_release_artifacts.sh --astria-release-acceptance-gates-smoke`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`

Earlier Phase20 child validations also covered:

- `scripts/validate_release_artifacts.sh --updater-transaction-plan-smoke`
- `scripts/validate_release_artifacts.sh --updater-rollback-health-gates-smoke`

## Updated Kocoro parity estimate

Astria is now roughly 96-98% aligned with Kocoro for local-first desktop
platform behavior.

The remaining local-first gap is no longer missing updater safety metadata.
Astria now has local gates for metadata verification, compatibility, staged
transaction planning, rollback/health requirements, release acceptance
requirements, crash diagnostics, native shell supervision, and Desktop RPC
session readiness.

## Remaining Kocoro gaps

- Real replacement execution: Astria still does not download, install, replace,
  roll back, or relaunch through a transactional updater.
- Real Apple release operations: local validation declares signing,
  notarization, and stapling requirements, but does not perform Developer ID
  signing, notarytool upload, stapling, or public release publication.
- End-to-end updater rehearsal: no sandboxed installed-app replacement rehearsal
  exists yet.
- Broader cloud/channel parity: local-first desktop parity is nearly complete,
  but Kocoro's cloud IM, remote delivery, and team/distribution surfaces remain
  outside the Astria local shell scope.

## Next decision

The next phase should choose between two clear paths:

1. If the priority is local-first parity, Phase21 should add a sandbox-only
   updater rehearsal that exercises staged replacement and rollback against a
   disposable app fixture, still without touching the installed Astria app.
2. If the local-first target is considered effectively complete, the next
   parity track should pivot to cloud/channel parity: remote delivery lifecycle,
   channel auth boundaries, and team/distribution controls.
