# Improve version and runtime status surface

## Goal

Make the Web UI Version and Settings surfaces more useful for local app startup and support diagnostics by showing runtime context alongside build/update information.

## Requirements

- Version API should expose stable local runtime fields that are already known by the daemon: daemon health/status URL, diagnostics URL, data directory, config path, and launch command.
- Web UI Version page should show those runtime fields in a readable section without duplicating noisy diagnostics details.
- Settings hub should keep a compact version/build signal while Version remains the detailed view.
- Existing update-check behavior must remain unchanged, including disabled checks for development builds.
- Existing Web UI smoke should cover the new fields.

## Acceptance Criteria

- [x] `/version` includes runtime URLs/paths useful for support and app launch troubleshooting.
- [x] Version page displays runtime context including Web UI, diagnostics, data, and config paths.
- [x] Development build update behavior remains disabled and clear.
- [x] Tests validate new Version API fields.
- [x] Web UI smoke validates the new Version page fields.
