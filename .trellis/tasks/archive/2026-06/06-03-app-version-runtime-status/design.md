# Design

## API

Extend the existing `/version` response rather than creating a new endpoint. The version payload already owns launch/update command hints, so adding runtime-support fields keeps the Web UI simple and avoids another API call.

## Web UI

Keep `Release readiness` as the first card. Add a `Runtime context` card below it with support-oriented fields. Leave Diagnostics as the detailed runtime readiness page.

## Compatibility

Adding JSON fields is backward-compatible for current clients. Existing field names and update-check semantics remain unchanged.
