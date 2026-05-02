# Iteration 009: Config Inspection

## Scope Delivered

This iteration adds a read-only config inspection path through both the socket API and CLI.

Delivered pieces:

- `GET /v1/config`
- `x11smctl config get`
- control-server test coverage for config response payloads

## Behavior

`GET /v1/config` returns the daemon's effective in-memory config object.

`x11smctl config get` fetches the same payload and prints formatted JSON.

This is read-only for now. Runtime mutation is still intentionally not implemented.

## Why This Slice Matters

At this point the daemon has accumulated several runtime defaults and operator-relevant settings:

- match policy
- X11 behavior
- helper process names
- reconcile interval
- socket location

Being able to inspect the effective config without reading local files makes end-to-end validation and service debugging much easier.

## Manual Validation Additions

Useful verification command:

```sh
x11smctl config get
```

Expected checks:

1. `match.who` contains `OpenCode`
2. `x11.disable_screensaver` and `x11.disable_dpms` are `true`
3. `processes.pause` contains `xss-lock`
4. `reconcile.interval` matches the expected runtime value

## Remaining Gaps

Still not implemented:

- partial runtime config update endpoint
- config file path reporting in status
- event stream or event history API

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
