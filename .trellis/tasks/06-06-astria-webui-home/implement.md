# Implementation Plan

1. Read applicable backend and shared Trellis specs.
2. Update Web UI HTML:
   - rebrand visible UI copy to Astria where appropriate
   - add home panel
   - reorganize sidebar navigation
   - preserve existing element ids used by JavaScript/tests
3. Update JavaScript:
   - default to `home`
   - add view metadata
   - keep home composer wired to existing submit flow
   - render home activity counts from runs/approvals
   - add shortcut/card interactions
4. Update CSS:
   - introduce Astria celestial tokens
   - style sidebar, home launcher, chips, cards, and responsive behavior
   - preserve existing panel/form styling for current workflows
5. Validate:
   - `go test ./...`
   - run at least the core Web UI smoke script if feasible
   - inspect resulting page where feasible
