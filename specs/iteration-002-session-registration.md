# Iteration 002: Session Registration

## Scope Delivered

This iteration adds the first X11-session-aware control flow:

- daemon-side in-memory session registration
- `POST /v1/session`
- CLI `register-session` command
- session details included in `GET /v1/status`
- tests for registration payload capture and control API behavior

## Why This Slice Exists

The daemon will eventually need to run X11 actions like `xset` against the correct live user session.

Hardcoding `DISPLAY` and `XAUTHORITY` in the user service is possible, but this project now prefers a session-registration model:

1. `x11smctl register-session` runs inside the real i3/X11 session
2. the CLI captures the current environment values
3. the daemon stores them as the active X11 session context
4. later X11 action code can use that stored context

This keeps the long-lived daemon independent from assumptions like `DISPLAY=:0`.

## Delivered API

### `POST /v1/session`

Request body shape:

```json
{
  "display": ":0",
  "xauthority": "/home/user/.Xauthority",
  "xdg_session_type": "x11",
  "dbus_session_bus_address": "unix:path=/run/user/1000/bus"
}
```

Current validation rules:

- `display` is required
- `xauthority` is required

Response:

- `204 No Content` on success

### `GET /v1/status`

Status now includes a `session` object when one has been registered, plus the existing `session_ready` flag.

## CLI Behavior

`x11smctl` now supports:

- `status`
- `register-session`

`register-session` currently reads these environment variables from the calling process:

- `DISPLAY`
- `XAUTHORITY`
- `XDG_SESSION_TYPE`
- `DBUS_SESSION_BUS_ADDRESS`

Current client-side validation:

- fail if `DISPLAY` is missing
- fail if `XAUTHORITY` is missing

## Current Limitation

Session registration is stored only in memory.

Implications:

- daemon restart loses the registered session context
- i3 or login startup should invoke `x11smctl register-session` again after daemon start

## Recommended Integration Shape

The intended user-session flow is now:

```i3
exec --no-startup-id x11smctl register-session
exec --no-startup-id xset s 180 240
exec --no-startup-id xss-lock -- slock &
```

This ordering ensures the daemon learns the live X11 environment early in session startup.

## Follow-up Dependencies

The next slice should build on this registration model by adding:

1. inhibitor source abstraction
2. matching rules package
3. reconcile loop skeleton
4. X11 runner that executes with the registered session environment

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
