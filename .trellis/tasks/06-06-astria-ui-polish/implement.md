# Astria UI Polish Implementation Plan

## Checklist

1. Read current Web UI structure and style patterns.
2. Add Inbox to Home docked tools and keep count rendering consistent.
3. Polish empty/action states for Inbox, Memory, MCP, and Council without changing APIs.
4. Add narrow viewport CSS refinements for Home docked tools, cards, and split panels where needed.
5. Extend smoke coverage for the new Home Inbox entry.
6. Run validation.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
git diff --check
go test ./internal/daemon
./scripts/smoke_webui_core.sh
```

## Rollback Points

- Revert Web UI HTML/CSS/JS changes if smoke reveals broad layout regression.
- Revert smoke selector additions separately if only test targeting is wrong.
