# UI redesign prototype

## Goal

Create a reviewable product UI redesign prototype for Astria before changing production code.

The prototype should respond to the current critique:

- The existing UI does not yet feel like a mature agent product.
- The product should carry a deep-space / starfield theme with restraint and utility.
- Chinese and English copy currently mix at the same hierarchy, producing awkward typography and poor adaptation.
- The interface needs stronger information architecture, not only a visual skin.

## Requirements

- Produce design artifacts only. Do not modify production UI files during this planning/prototype step.
- Keep Astria as the product-facing UI name and StarClaw as the project / CLI identity.
- Use Chinese as the primary UI language, while preserving technical terms where they are stronger as English: Agent, MCP, Prompt, API.
- Reframe the product around a mature local agent workspace: task intake, run observation, artifact review, context management, and system controls.
- Reduce top-level navigation complexity into a smaller set of first-class areas.
- Establish a deep-space visual direction that feels calm, precise, and product-grade rather than decorative.
- Include typography guidance for Chinese and English mixed content.
- Provide a static HTML prototype that can be opened locally for visual review.
- Do not require a build system or network access for the prototype.

## Acceptance Criteria

- [ ] `design.md` captures the information architecture, visual language, typography rules, interaction model, and implementation constraints.
- [ ] `prototype.html` provides a complete first-screen concept for the redesigned Astria workspace.
- [ ] Prototype uses Chinese-first UI copy with intentional English technical terms.
- [ ] Prototype demonstrates the deep-space theme without hiding the utility of the workspace.
- [ ] Prototype includes the five proposed first-class areas: `任务台`, `运行`, `产物`, `上下文`, `系统`.
- [ ] Prototype is self-contained and can be opened from the filesystem.
- [ ] No production UI files are changed as part of this prototype-only step.

## Notes

- Current production Web UI files already have uncommitted changes in the worktree. Treat those as out of scope for this prototype step.
- This task remains in planning until the prototype direction is reviewed.
