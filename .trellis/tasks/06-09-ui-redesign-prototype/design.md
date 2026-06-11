# Astria UI Redesign Prototype

## Design Read

Astria should feel like a mature local agent workspace for technical and semi-technical users. The brand direction is deep-space, but the product must remain quiet, useful, and repeatable for daily work. The interface should not become a decorative sci-fi poster.

## Product Thesis

Astria is not a generic chat app. It is a local command center for delegating work to agents, watching execution, reviewing outputs, and controlling context and permissions.

The redesigned UI should make one loop obvious:

1. Describe the task.
2. Choose the operating context.
3. Watch the run.
4. Review the result.
5. Save or route the artifact.

## Information Architecture

The current UI exposes too many peer-level destinations. The prototype reduces the product to five first-class areas:

| Area | Role | Existing concepts absorbed |
| --- | --- | --- |
| 任务台 | Start and steer work | Home, Chat, workflow recipes, prompt suggestions |
| 运行 | Observe execution | Runs, approvals, quality, compare, budget |
| 产物 | Review reusable output | Results, reuse gallery, playbooks, starter kits, share packs, snapshots |
| 上下文 | Manage working memory and inputs | Files, memory, sources, browser, data, MCP, skills, agents |
| 系统 | Runtime and safety controls | Settings, diagnostics, permissions, connectors, schedules |

Secondary tools can still exist, but they should appear as contextual panels inside these five areas instead of permanent top-level navigation.

## Layout Model

The prototype uses a three-zone workspace:

- Left rail: compact product navigation and operational status.
- Center command field: the primary task intake and active mission state.
- Right observatory: run timeline, context readiness, and review queue.

This preserves the maturity of a productivity app while giving the starfield concept a structural role: tasks become orbits, context becomes docked material, and runs become observable telemetry.

## Visual Language

### Palette

The base is deep navy-black, not pure black.

- Void base: `#070b16`
- Panel base: `#0d1424`
- Elevated panel: `#111b2e`
- Border: `rgba(143, 164, 214, 0.18)`
- Primary text: `#eef3ff`
- Secondary text: `#93a4c7`
- Muted text: `#607091`
- Main accent: astral cyan `#76e4ff`
- Warm signal accent: star gold `#f3c969`
- Risk accent: red ember `#ff6b6b`
- Success accent: orbital green `#6ee7b7`

The star theme is carried by low-contrast radial light, faint grain, and small fixed stars. Avoid large purple-blue AI gradients and ornamental blobs.

### Surfaces

Use layered dark panels with hairline borders and inner highlights. Panels should feel like glass over a dark observatory surface, but content readability wins over transparency.

Cards should only frame distinct operational units:

- current mission
- active run
- context bundle
- review item
- artifact

Do not wrap every section in a card.

### Motion

Motion should be operational:

- subtle command field focus glow
- timeline pulses for active work
- staged panel entrance on first load
- reduced motion fallback

Avoid constant background animation that competes with task reading.

## Typography

### Language Rule

Use Chinese as the primary interface language. Preserve English only where it is a product or technical noun:

- Astria
- StarClaw
- Agent
- MCP
- Prompt
- API
- Run ID

Do not mix English headings and Chinese explanations at the same hierarchy. If a term needs English, place it as a secondary label or small technical tag.

### Font Strategy

Production should use a Chinese-capable stack first and a mono stack for telemetry:

```css
font-family: "PingFang SC", "Microsoft YaHei", "Noto Sans CJK SC", "Source Han Sans SC", system-ui, sans-serif;
font-family: "JetBrains Mono", "SFMono-Regular", "Cascadia Code", monospace;
```

Avoid relying on system UI alone for brand expression. If a self-hosted font is introduced later, prefer one Chinese-compatible UI font for body text and one mono font for IDs, tokens, usage, and run telemetry.

### Hierarchy

- Product title: compact and confident, not oversized marketing type.
- Panel headings: Chinese, short, sentence-like.
- Technical tags: small mono text.
- Numbers: tabular figures.
- Buttons: two to four Chinese characters where possible.

## Copy Direction

Replace vague labels with work-oriented labels:

| Current style | Preferred style |
| --- | --- |
| 首页 | 任务台 |
| 消息 | 对话 |
| Runs | 运行 |
| Focus Brief | 任务摘要 |
| Mission mode | 任务模式 |
| Docked tools | 已连接能力 |
| Result Library | 产物库 |
| Agent Council | 多 Agent 评议 |

## Prototype Scope

The static prototype represents the first screen and core shell. The second iteration is intentionally desktop-app shaped, not webpage shaped. It uses a fixed-height window, a narrow app rail, a central message/task stream, and a right observatory panel.

## Kocoro Reference Principles

Kocoro Desktop is installed locally at `/Applications/Kocoro Desktop.app`. Its WebView message surface is compact and desktop-native:

- The window is fixed-height and `overflow: hidden` at the app frame level.
- The message stream is the primary surface.
- Human messages use compact right-aligned bubbles.
- Agent messages read as a vertical work log, with tool events embedded inline.
- Status and actions are lightweight, not dashboard-heavy.
- Borders are hairline, radii are modest, and typography is small enough for daily work.

Astria should borrow these structural principles without copying Kocoro's brand. The Astria layer adds:

- deep-space color and faint star material
- a local-agent observatory panel
- explicit run telemetry, context readiness, and approval state
- the five-part information architecture from this redesign

This means the prototype should feel like a desktop agent console, not a long landing page or analytics dashboard.

## macOS Native Reference Principles

The third prototype iteration adds macOS-native structure research:

- AppKit's new design frames key window regions in glass. Toolbar elements should appear grouped on glass surfaces, while non-interactive labels should avoid looking like buttons.
- Sidebars and inspectors have different jobs. The sidebar is navigation; the inspector belongs to the current selection or mission. They should not feel like two identical side panels.
- Toolbar controls should be grouped by behavior. Search, primary actions, status, and navigation controls need separate visual groups.
- Concentricity matters: window radius, toolbar radius, sidebar radius, and inner controls should feel nested rather than randomly rounded.
- Liquid Glass should clarify hierarchy, not become a blur effect pasted over everything.
- Raycast contributes a different macOS lesson: command-first, keyboard-first, and fast switching. Astria should make `Command-K` style search feel central.

## Star Semantics

The word `星辰` should be treated as a product model, not a decoration.

- `星` = a visible work object: task, agent, file, tool, result, approval.
- `辰` = time and orientation: sequence, phase, schedule, active moment.
- `星图` = the relationship map among goals, context, tools, and outputs.
- `轨道` = the run lifecycle.
- `亮度` = urgency, activity, or confidence.
- `星座` = reusable workflows and connected capabilities.

This gives the theme operational meaning. The UI can show a constellation-like task map without becoming a fantasy screen.

## Further Design Language Directions

The design field around space, science, and telemetry suggests several additional directions for Astria. These are not visual effects to paste on top; each should map to a real product behavior.

### 1. Mission Control Calm

NASA mission interfaces emphasize consistency, live telemetry clarity, stale-data warnings, command feedback, and avoiding mental transposition. Astria can borrow that rigor:

- Every run state should be visually explicit: running, paused, waiting, stale, failed, completed.
- Approval and risky commands need consequence previews before execution.
- Data should be shown at the level needed for action, not as decorative metrics.
- The baseline UI should be calm enough that a problem becomes immediately visible.

This argues for restrained color, stable panels, and very clear state language.

### 2. Astronomical Cartography

Star maps are not just pretty skies. They are coordinate systems. Astria can treat the workspace as a map:

- tasks are stars
- agents are observers or instruments
- tools are constellations of capability
- files and memories are reference catalogs
- runs are tracks or plotted paths
- artifacts are named discoveries

This could produce a unique interaction model: zoom from high-level constellation map into the detailed run log.

### 3. Scientific Instrument Panels

Science software often separates the observed object, the instrument settings, and the analysis output. Astria can use the same separation:

- center: observed mission or task graph
- left: navigation and instrument choice
- right: inspector with measurements, logs, approvals, and metadata
- bottom: command input

This keeps the interface focused without hiding power.

### 4. Quiet Cinematic Space

Tech illustration can teach atmosphere, but the UI should not become a poster. Use cinematic cues only where they support product meaning:

- thin orbit lines for lifecycle
- parallax depth for hierarchy
- tiny spectral color accents for status
- glints for active nodes
- soft horizon light to separate toolbar from content

Avoid loud nebula backgrounds, huge gradients, or constant motion.

### 5. Command Observatory

Raycast-like command-first behavior and mission-control observability can coexist. The best Astria shell may be:

- `Command-K` as the universal launcher
- a star map as mission overview
- a message/run log as drill-down
- an inspector for consequence, context, and artifact metadata

This makes Astria feel like a cockpit for local agents rather than a chat clone.

## Prototype Iteration 4: Operable Star Instrument

The fourth iteration moves the prototype from a striking concept toward a product instrument:

- Star nodes now have labels and state roles, so the map can be read without guessing.
- The central star map includes a run-orbit timeline, connecting the poetic theme to execution state.
- The bottom command field exposes working context chips such as workspace, approval state, and selected agent.
- The inspector now includes instrument dials, not only descriptive cards.
- Context references are shown as a catalog, matching the astronomical metaphor while remaining utilitarian.
- A `星图 / 日志` segmented control makes the future interaction model explicit: high-level map first, drill-down log second.

This version should be judged less like a static marketing mockup and more like a possible desktop product shell.

It is intended to answer these review questions:

- Does the deep-space direction feel mature rather than decorative?
- Does the Chinese-first copy resolve the current mixed-language awkwardness?
- Does the reduced navigation model make the product easier to understand?
- Does the task intake feel like the primary workflow?
- Does the right observatory panel make agent work observable without overwhelming the user?
- Does the shell fit a desktop window without creating a long-page feel?

## Implementation Constraints For Later

When implementation begins, preserve existing daemon APIs and DOM IDs required by smoke tests unless the tests are updated intentionally. The current Web UI is served from embedded static files:

- `internal/daemon/webui/index.html`
- `internal/daemon/webui/assets/styles.css`
- `internal/daemon/webui/assets/app.js`

The production migration should be incremental:

1. Introduce design tokens and typography in CSS.
2. Restructure visible navigation while preserving panel wiring.
3. Update copy and labels systematically.
4. Redesign the home/task screen.
5. Redesign run and output surfaces.
6. Validate with Web UI smoke scripts and Playwright screenshots.

## Prototype Iteration 5: Star Map + Observatory Log

The fifth iteration makes the shell feel less like a static concept image and more like a working desktop application:

- The center now has two product layers: a relationship star map above, and a compact observatory log below.
- The selected star is `Kocoro / macOS`, making the inspector feel tied to a real object rather than a generic right panel.
- The log records concrete design events, so the user can see why a node matters and what changed recently.
- Inspector metrics were renamed to product-usable signals: context completeness, approval risk, tool readiness, and artifact trust.
- English is retained only where it acts as a compact technical label; Chinese carries the user-facing meaning.
- The fixed-height desktop window is preserved. The design should not become a long dashboard or landing page.

This version tests whether Astria's first visual identity can be `星图 + 观测日志 + 上下文检查器`: a local Agent observatory rather than another chat clone.

## Prototype Iteration 6: Product Texture Pass

The sixth iteration reduces decorative density and increases desktop-product credibility:

- The observatory log is no longer three decorative cards. It now behaves more like a compact activity stream with time, event type, object, description, and result.
- The inspector has been toned down from HUD-style dials into object signals, properties, and next actions.
- The selected context object now has a clearer operational role: it can be adopted, scheduled, or confirmed.
- The prototype remains a fixed desktop window, preserving the user requirement that this should not feel like a long web page.
- The design language is still distinctive, but the UI now carries more real Agent workflow semantics: read, select, render, gate.

This pass should be judged on whether the interface now feels less like a poster and more like a professional local Agent application.

## Prototype Iteration 7: High-Fidelity Visual Pass

The seventh iteration upgrades visual fidelity without changing the information architecture:

- The background now has a deeper astronomical material: layered radial light, subtle star grain, and a vignette-like depth field.
- The desktop window uses a double-bezel treatment with inner highlights, a stronger glass shell, and a thin spectral separator under the toolbar.
- The star map now reads more like an optical instrument: lens-like center lighting, brighter selected-node halo, refined beams, and more tactile labels.
- Sidebar, toolbar, command search, activity log, and inspector now share the same glass/metal material system.
- Motion is restrained: window entrance, section rise-in, beam reveal, and active-node pulse. No decorative animation loops beyond the selected signal.

This pass is intended to make the prototype feel premium and memorable while preserving the practical desktop Agent workflow established in Iterations 5 and 6.

## Prototype Iteration 8: Peripheral Element Upgrade

The eighth iteration extends the high-fidelity language beyond the main star map:

- The sidebar now has active-route light rails, hover movement, and a bottom local-status capsule.
- Toolbar details were refined: traffic lights, command-search depth, keyboard hint capsule, segmented switch, and primary button highlight.
- Run signals now behave like compact status instruments with live dots.
- The bottom composer now reads as a command bay, with a subtle top light line, richer surface material, and stronger primary action.
- Inspector sections gained object-status chips, section beacons, micro progress bars, and more tactile evidence/action rows.

This pass should make the whole desktop shell feel coherent, so the main canvas is not the only visually mature area.

## Prototype Iteration 9: Deeper Star Semantics

The ninth iteration deepens the meaning of `星辰` inside the UI itself:

- The star map now contains a compact semantic legend: `星` as visible work object, `辰` as time/phase/orientation, `亮度` as priority/activity, and `星座` as workflow composition.
- A `辰刻` instrument was added to show the active phase and make time part of the interface, not just a progress bar.
- The activity log now uses star-language result states such as `星 +1`, `发光`, `产物`, and `辰定`.
- The inspector includes a `Star Semantics` section that explains why the selected Kocoro/macOS reference is a context star and how its brightness affects design decisions.
- On narrow screens, the semantic instruments are hidden to preserve the fixed desktop shell and avoid overlap.

This pass should make the theme feel conceptually earned: `星辰` becomes a product grammar for local Agent work, not a surface aesthetic.

## Prototype Iteration 10: Observatory Precision

The tenth iteration pushes the interface further toward an astronomical observation system:

- Star nodes now expose a magnitude-like readout (`M0.2`, `M1.1`, etc.), giving brightness and priority a measurable UI language.
- The orbit now has phase markers (`01` through `04`) with the active phase highlighted, connecting the central map to the run lifecycle.
- Activity logs gained observation IDs (`OBS-1` through `OBS-4`) so the log reads like an observatory record rather than a generic feed.
- The inspector now includes an `Observation` card with spectrum, magnitude, observation ID, orbit phase, and spectral role.
- Narrow screens hide the precision overlays to preserve usability.

This pass treats Astria as a local agent observatory: every visible element has a measurable role in the mission, not just an aesthetic role.

## Prototype Iteration 11: Dynamic Constellation Motion

The eleventh iteration prototypes the motion language for active LLM/tool work:

- Star-to-star connections now animate as flowing dashed paths, representing tool-call propagation between target, context, and tool nodes.
- Energy packets travel across the map to suggest information moving through the agent system.
- A distant approval star and its label periodically fade and blur, representing remote context losing relevance or being consumed by the current conversation.
- An annihilation ring appears at the fading node to make the "distant stars gradually vanish" concept legible without needing a full interactive runtime.
- The activity log now mirrors the animation with `tool` and `fade` events.

This pass defines a future implementation rule: when the LLM calls a tool, the UI should show a transient connection between the involved task/context/tool/artifact stars; unused distant stars may dim or collapse as conversation focus narrows.

## Prototype Iteration 12: Agent Link Grammar

The twelfth iteration expands the motion concept from one animated tool-call into a reusable grammar for many Agent actions:

- `tool`: short bright bridge with energy packets moving between target, context, and tool stars.
- `context`: gold breathing track that wakes a related constellation without overpowering the main task.
- `edit`: green converging path and burst, mapping file changes into an artifact star.
- `retry`: rose fractured path, mapping failed or risky execution into a broken route that can be re-plotted.
- `fade`: distant stars can dim, blur, and collapse when the conversation no longer uses them.

The star map now includes a secondary grammar-link layer and a faint workflow halo. The inspector includes a `Motion Grammar` panel so reviewers can understand the system rather than guessing from animation alone.

Future production implementation should bind these motion states to real run events: tool calls, context reads, file edits, approvals, retries, sub-agent spawning, memory saves, and final artifact formation.

## Prototype Iteration 13: Companion, Memory, Artifact Motions

The thirteenth iteration adds three subtle micro-motions without increasing page length or replacing the existing visual language:

- `spawn`: a companion star orbits the target star, representing a sub-agent or helper process attached to the main mission.
- `memory`: a small comet leaves the center orbit and settles into a distant low-brightness star, representing memory save or reusable knowledge being archived.
- `artifact`: a green formation ring blooms around the artifact star, representing file edits converging into a reviewable output.
- The activity log now describes artifact formation and memory sedimentation.
- The `Motion Grammar` inspector panel now includes `spawn` and `memory` rules in addition to tool/context/edit/retry.

This pass preserves the high-fidelity star map while making the motion system cover more real Agent actions.

## Prototype Iteration 14: Event Sequencer Contract

The fourteenth iteration adds control and legibility to the motion system:

- A compact event sequencer appears inside the star map, showing `tool → edit → spawn → memory → fade` as an ordered run sequence.
- The sequencer animates softly so reviewers can see that motion is tied to state progression, not random visual activity.
- The observatory log header now describes the feed as an event sequence with five states.
- The inspector includes an `Event Mapping` panel that maps UI motion to future run events such as `tool_call.started`, `file_edit.applied`, `sub_agent.spawned`, and `memory.saved`.
- Small-screen rules hide the sequencer to preserve layout quality.

This pass turns the visual language into an implementation contract: real Agent events should drive specific star motions, and every motion should be explainable from the run stream.

## Prototype Iteration 15: Causal Path Readability

The fifteenth iteration improves readability without adding more animation density:

- Small causal chips now label the key animated routes: `context wakes tool`, `edit forms artifact`, and `retry replots orbit`.
- The mission card now includes a compact breadcrumb path: `target → context → tool → artifact → memory`.
- The inspector's `Event Mapping` section includes the current causal path as an explicit run-path contract.
- These additions make the motion layer easier to understand at a glance: the user can read the cause-and-effect chain instead of only watching light move.
- On small screens, causal chips are hidden with other precision overlays.

This pass keeps visual quality stable while making the star map more operationally legible.

## Prototype Iteration 16: Motion Control And Focus

The sixteenth iteration adds product-level control over the increasingly rich motion system:

- A toolbar `Motion: balanced` control makes animation intensity visible as a user/product setting.
- The star map now shows a compact focus trail: `target / context / tool`, clarifying which stars are currently allowed to animate strongly.
- The inspector includes a `Control State` panel with motion intensity, focus trail, distant-star fade policy, and reduced-motion awareness.
- These changes prevent the constellation UI from feeling like uncontrolled ambient animation.
- Small-screen rules hide the focus trail with other precision overlays.

This pass improves maturity: Astria can have expressive star motion while still feeling controllable, inspectable, and suitable for a desktop productivity app.

## Prototype Iteration 17: Link Health And Safety Signals

The seventeenth iteration adds state confidence to the visual system:

- The `Gate` signal now reads as a review state instead of a generic counter, with a gold waiting indicator.
- The star map includes a compact `Link Health` readout for active, review, and remote/decaying chains.
- The inspector includes a `Link Health` panel that mirrors those states: stable active chain, pending review gate, and decaying remote context.
- This strengthens the desktop-product feel: users can judge what is active, what is waiting, and what is safely fading out.
- Small-screen rules hide the health overlay with other precision controls.

This pass makes the star-motion system safer and more operationally legible without reducing the existing visual fidelity.

## Prototype Iteration 18: Calm Default, Active Reveal

The eighteenth iteration intentionally moves from exploration toward product maturity:

- The main star map keeps its visual identity, but secondary precision overlays now sit quieter by default.
- Hover/focus reveals richer overlays, modeling the future behavior where Agent activity or user inspection brings the motion layer forward.
- The toolbar motion control now expresses a two-state model: `calm` default shell and `active` run overlays.
- The inspector's `Control State` panel now documents this hierarchy explicitly.
- The review gate was rewritten around the implementation decision: use a calm shell by default, reveal star links during Agent execution.

This pass is a deliberate reduction step. It preserves the star-language progress while preventing the prototype from becoming visually noisy or difficult to implement.

## Prototype Iteration 19: Design Hardening

The nineteenth iteration closes the exploratory loop and turns the prototype into a more implementation-ready design reference:

- Inspector section titles were localized to Chinese-first labels. Technical English remains only where it names actual event types or implementation contracts.
- A `状态命名` section was added so `M`, `OBS`, and spectral roles are not decorative abbreviations.
- The `控制状态` panel now includes a performance policy: pause off-focus.
- The review path remains implementation-oriented rather than concept-oriented.
- This pass is intentionally not a new visual direction; it makes the current direction more durable.

### State Naming Table

These names should be treated as implementation-facing UI tokens:

- `M` / magnitude: priority, brightness, and influence of a visible work object.
- `OBS`: observation event ID in the run stream.
- `spectrum role`: object type plus status color, such as `context-gold` or `artifact-green`.
- `calm shell`: default quiet desktop state.
- `active overlays`: transient motion/precision layer shown during Agent activity or inspection.
- `fade unused`: policy for distant context that is not used by the current conversation.

### Motion And Performance Policy

Production should not run every animation at full intensity all the time:

- Respect `prefers-reduced-motion`; replace path travel with static state highlights.
- Pause nonessential animations when the app window is unfocused.
- Keep `calm shell` as default; reveal `active overlays` on run events, hover, focus, or selected objects.
- Avoid animating layout properties. Use transform, opacity, stroke-dashoffset, and filter sparingly.
- Make all motion explainable from the run stream: no animation should exist without an event or state reason.

### Production Component Split

The production migration should be broken into these components:

- `AstriaShell`: macOS-like window, toolbar, sidebar, inspector layout.
- `CommandSearch`: command/search control and keyboard hint.
- `MissionHeader`: mission title, run/agent/gate status.
- `StarMap`: visible work objects, orbits, labels, semantic overlays.
- `MotionLayer`: tool/context/edit/retry/spawn/memory/fade animations driven by run events.
- `EventSequencer`: ordered event state strip.
- `ObservationLog`: compact event feed with OBS IDs.
- `InspectorPanel`: selected object details, observation, grammar, mapping, health, controls.
- `CommandComposer`: task prompt, context chips, primary action.

The next implementation step should start with `AstriaShell`, `MissionHeader`, `ObservationLog`, and a static `StarMap`, then add `MotionLayer` after event data is available.
