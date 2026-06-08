# Native tool parity followup implementation plan

## Checklist

- [x] Confirm local Kocoro checkout exists and is at commit `74cdb3c`.
- [x] Inspect Phase 8 parent requirements and prior Kocoro comparison.
- [x] Compare Kocoro and StarClaw native/tool/Desktop files and registration surfaces.
- [x] Write a Phase 9 research plan with evidence, gap estimates, and task ordering.
- [x] Update the Phase 8 parent acceptance criteria after this child is archived.
- [x] Validate the task artifacts.
- [x] Archive this child task with `--no-commit`.
- [x] Commit planning and archive changes.

## Validation Commands

```bash
python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-native-tool-parity-followup
git diff --check
```

After archive:

```bash
python3 ./.trellis/scripts/task.py validate .trellis/tasks/archive/2026-06/06-08-native-tool-parity-followup
```

No Go tests are required unless runtime code changes are made.

## Review Gate

Before archiving, confirm:

- The research document names the local Kocoro checkout and commit.
- Phase 9 child tasks are independently verifiable.
- Keychain/sync/cloud/image-provider work is not accidentally treated as enabled-by-default behavior.
- The Phase 8 parent can truthfully mark its final comparison criterion complete.
