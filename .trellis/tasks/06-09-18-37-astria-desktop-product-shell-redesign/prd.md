# Astria desktop product shell redesign

## Goal

Turn the Astria desktop surface from a long daemon workbench into a focused, Kocoro-aligned desktop product shell.

## Why

The current UI exposes too many platform capabilities at once. Users open the app and see a long, scroll-heavy control console instead of a clear place to start work. Kocoro's mature UI language is flatter, quieter, and task-first: white surfaces, fine borders, no decorative gradients/shadows, and secondary controls hidden until needed.

## Scope

- Make the default `/app/` screen usable in a 1040x720 desktop window without requiring vertical page scrolling.
- Reduce primary navigation to a small set of user-facing destinations.
- Move secondary capabilities into progressive disclosure / tool drawer surfaces.
- Preserve existing panels and data bindings so platform capabilities remain reachable.
- Remove decorative "stellar" gradients/shadows from the primary desktop shell.

## Acceptance

- First viewport presents one dominant mission/chat composer and a compact recent-run/status area.
- Left sidebar no longer reads as a full feature inventory.
- Advanced features are reachable via management/tool surfaces, not always visible.
- Browser verification shows no page-level overflow on the default desktop viewport.
- Existing core WebUI smoke tests still pass.
