# Design

## Script Shape

Create a shared shell harness:

- `scripts/lib/webui_smoke_common.sh`
  - build binary
  - create isolated HOME/config
  - install Playwright dependency
  - start/stop daemon
  - route curl checks
  - run a generated Node script

Layer scripts call the harness with a mode:

- `scripts/smoke_webui_core.sh`
- `scripts/smoke_webui_permissions.sh`
- `scripts/smoke_webui_agents.sh`
- `scripts/smoke_webui_runs.sh`

`scripts/smoke_webui.sh` runs the full mode for backward compatibility.

The Node script can remain generated from shell for this iteration, but it should dispatch by `WEBUI_SMOKE_MODE`:

- `core`: app loads, diagnostics, config save, static routes.
- `permissions`: core + permissions editor save/clear.
- `agents`: core + agent CRUD, command editor, import/export, direct test runner.
- `runs`: core + schedules, chat summary, run history/detail, sessions, approval.
- `full`: all layers in sequence.

## CI

Add a workflow step after build/test prerequisites:

```yaml
- name: Web UI smoke core
  run: scripts/smoke_webui_core.sh
```

Core only is selected because it verifies daemon-hosted Web UI boot and basic config without exercising long editor workflows.

## Validation

Run each layer individually and the full entrypoint:

```bash
scripts/smoke_webui_core.sh
scripts/smoke_webui_permissions.sh
scripts/smoke_webui_agents.sh
scripts/smoke_webui_runs.sh
scripts/smoke_webui.sh
```

Then run Go test/vet and diff check.
