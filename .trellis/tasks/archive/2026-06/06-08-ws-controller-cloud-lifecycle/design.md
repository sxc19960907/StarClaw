# Design

## Controller

Add `internal/daemon/cloud_lifecycle.go`.

Types:

- `CloudLifecycleController`
- `CloudLifecycleRunner func(context.Context) error`
- `CloudLifecycleStatus`

Fields:

- parent context
- runner
- cancel
- done channel
- running flag
- started at
- stopped at
- restart count
- last error
- enabled/configured flag and note

Methods:

- `Start(context.Context)`
- `Stop()`
- `Restart(context.Context)`
- `Status() CloudLifecycleStatus`

Default runner:

- blocks until context cancellation and returns `nil`.
- does not perform network I/O.
- status note says cloud WebSocket transport is a disabled local lifecycle boundary.

## API

Add `internal/daemon/cloud_lifecycle_api.go`:

- `GET /cloud/lifecycle` returns status.
- `POST /cloud/lifecycle` accepts:
  ```json
  {"action":"start|stop|restart"}
  ```

Invalid actions return 400.

## Server Wiring

Add `cloudLifecycle *CloudLifecycleController` to `Server`.

Initialize with `context.Background()` in `NewServer`. This controller is process-local; daemon shutdown integration can be added when a real WS runner exists.

## Safety

- No secrets in status.
- No network calls by default.
- No automatic start on daemon boot.
