# Iteration 004: Real Logind Source

## Scope Delivered

This iteration replaces the production app's no-op inhibitor source with a real `systemd-logind` system-bus source.

Delivered pieces:

- real `org.freedesktop.login1.Manager.ListInhibitors` integration
- normalization of D-Bus inhibitor rows into the existing `internal/observe.Inhibitor` model
- test coverage for normalization and failure propagation
- app wiring updated to use the real system bus source by default

## D-Bus Call Shape

The source now calls:

- service: `org.freedesktop.login1`
- path: `/org/freedesktop/login1`
- method: `org.freedesktop.login1.Manager.ListInhibitors`

The raw row data currently decoded from D-Bus is:

- `what`
- `who`
- `why`
- `mode`
- `uid`
- `pid`

The daemon currently normalizes and retains:

- `what`
- `who`
- `why`
- `uid`

`mode` and `pid` are decoded but not yet surfaced because the current matcher and status model do not need them.

## App Wiring Behavior

`internal/app.New` now tries to construct a real system-bus source.

If successful:

- the normal reconcile loop queries live `logind` inhibitor state

If startup source construction fails:

- app construction falls back to a static error source
- the daemon still starts
- reconcile moves the daemon into `degraded`
- status remains inspectable for diagnosis

This failure mode is intentional and matches the project goal of remaining observable when dependencies are unavailable.

## Why This Slice Matters

This is the first version that can report real inhibitor state from the host system instead of only proving the architecture with fakes.

With this slice in place, `x11smctl status` can now show live matching inhibitor information once a session is running and `logind` is reachable.

## Remaining Gaps

Still not implemented:

- explicit operator-facing `doctor` diagnostics for system bus and `login1`
- X11 command execution using the registered session environment
- process control for `xss-lock`

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
