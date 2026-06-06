# Design

## Scope

The daemon Web UI remains a static embedded HTML/CSS/JS app. The change adds a new `home` panel as the default route and reuses the existing chat submission pipeline for the home composer.

## UI Structure

- Sidebar:
  - Brand: Astria, local celestial agent workspace.
  - Workflow group: Home, Chat, Runs, Agents, Schedules, Publish.
  - Capabilities group: Skills, Connectors, Browser, MCP Servers, Local Files, Memory Map.
  - System group: Diagnostics, Config, Permissions, Version, Settings hub.
- Home:
  - Hero/welcome centered in the workspace.
  - Activity counters derived from `state.runs` and `state.approvals`.
  - Mission composer uses the existing `chat-form` element ids to avoid duplicating stream logic.
  - Shortcut chips either navigate to existing panels or seed useful prompts in the composer.
  - Cards navigate to existing panels or seed future-facing prompts without requiring backend support.

## Compatibility

Existing tests and scripts rely on ids such as `chat-form`, `chat-input`, `chat-agent`, `send-button`, and `chat-output`. Keep those ids on the primary form controls. The transcript output can live in the Chat panel, but the home composer should submit through the same controls.

## Styling

Use a light macOS desktop-app shell as the base:

- warm off-white background
- soft sidebar rail
- dark text
- star blue primary accent
- cyan/gold/pink status accents
- subtle orbit/radial decoration only on the home panel

Avoid a full dark sci-fi theme or decorative clutter.
