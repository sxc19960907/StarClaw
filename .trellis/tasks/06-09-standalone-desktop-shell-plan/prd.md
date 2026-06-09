# Standalone desktop shell plan

## Goal

Define and scaffold the first standalone Astria desktop shell boundary so
Phase13 can move beyond `starclaw app` opening a browser while preserving the
existing daemon and Web UI architecture.

## Requirements

- Compare viable shell choices against StarClaw's current codebase and
  Kocoro's Desktop lifecycle model.
- Recommend the first implementation route, including repository layout,
  binary bundling, local data paths, and launch command shape.
- Keep the initial app shell thin: host the existing daemon-served Astria Web UI
  and delegate runtime behavior to the daemon.
- Do not replace the Web UI, add cloud transport, or require signing
  credentials in this first child task.
- Capture enough implementation detail for the next child
  `daemon-supervision-app-launcher` to start without re-deciding architecture.

## Acceptance Criteria

- [ ] `design.md` documents SwiftUI/native, Tauri, Electron, and current
      CLI-only trade-offs.
- [ ] `implement.md` defines a concrete first MVP skeleton and validation path.
- [ ] The selected shell route preserves `starclaw app`, `--no-open`, and
      `--check`.
- [ ] App-to-daemon launch/attach contract is documented, including whether the
      shell uses HTTP health only, Desktop RPC, pidfile, or a staged mix.
- [ ] Follow-up child requirements are updated if planning discovers a better
      task split.

## Notes

Complex planning task. Add `design.md` and `implement.md` before starting.
