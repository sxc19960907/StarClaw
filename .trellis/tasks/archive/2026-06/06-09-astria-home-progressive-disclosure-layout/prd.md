# Astria home progressive disclosure layout

## Goal

Reduce the Astria Home page's first-screen length and density by moving
secondary workbench content behind progressive disclosure. The page should keep
the primary mission composer immediately visible while making strategy,
recipes, suggestions, workspace context, knowledge, tool, approval, and review
sections accessible after a second click.

## Requirements

- Preserve existing Home workflows and data bindings.
- Keep the mission composer, activity status, and key launch actions visible on
  the first screen.
- Move lower-priority Home modules into collapsible/secondary areas.
- Make collapsed content reachable with clear buttons and keyboard-friendly
  native controls.
- Improve narrow/mobile layout so the main Home content is visible rather than
  being pushed offscreen by the sidebar.
- Do not introduce a frontend build step or external assets.

## Acceptance Criteria

- [ ] Desktop first screen shows the primary mission area without the long
      secondary module stack immediately dominating the page.
- [ ] Strategy/workflow/suggestion/context/review modules are available through
      progressive disclosure.
- [ ] Mobile/narrow viewport does not horizontally overflow and shows Home
      content instead of only the sidebar.
- [ ] Existing Web UI smoke tests still pass.
- [ ] Browser visual check covers desktop and mobile widths.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
