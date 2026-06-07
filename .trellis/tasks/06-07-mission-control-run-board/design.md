# Mission Control Run Board Design

## UI

Add a compact run board above `#runs-list`:

- Four summary cards: Active, Needs attention, Completed, Total.
- Filter buttons with counts: All, Active, Attention, Completed, Council.

The run list remains the existing row item list but reads from `filteredRuns()`.

## JavaScript

Add:

- `state.runFilter = "all"`
- `runStatusGroup(run)`
- `filteredRuns()`
- `renderMissionControl()`

`renderRunsList()` should call `renderMissionControl()` and render filtered runs. Existing run detail behavior stays unchanged.

## Testing

Extend core Web UI smoke to verify the Mission Control controls render and a filter can be selected without breaking run detail.
