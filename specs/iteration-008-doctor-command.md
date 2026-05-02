# Iteration 008: Doctor Command

## Scope Delivered

This iteration adds the first operator-focused diagnostics command.

Delivered pieces:

- `x11smctl doctor`
- socket reachability check
- system bus reachability check
- `logind` reachability check
- `DISPLAY` and `XAUTHORITY` presence checks
- `xset` path lookup check
- `xset q` call-and-parse check using the invoking environment
- test coverage for doctor report generation and rendering

## Current Doctor Checks

`x11smctl doctor` currently verifies:

- daemon socket reachable through `GET /v1/status`
- system bus reachable
- `org.freedesktop.login1` reachable
- `DISPLAY` is set
- `XAUTHORITY` is set
- `xset` is installed
- `xset q` runs and its output is parseable by the daemon's parser

## Output Behavior

The command prints one line per check using:

- `[PASS] <name>`
- `[FAIL] <name>: <detail>`

The command exits nonzero if any check fails.

## Why This Slice Matters

At this point the daemon can:

- talk to `logind`
- register a live X11 session
- apply and restore X11 state
- pause and resume helper processes
- accept runtime control commands

`doctor` is the first compact command that checks the most important prerequisites for that end-to-end path.

## Manual End-To-End Flow Update

The recommended real-session validation flow now starts with:

```sh
x11smctl doctor
```

Then continue with:

```sh
x11smctl register-session
x11smctl status
systemd-inhibit --what=sleep:idle --who=OpenCode sleep infinity
x11smctl status
x11smctl disable
x11smctl enable
x11smctl reconcile
x11smctl status
```

## Remaining Gaps

Still not implemented:

- config get/set support
- event history endpoint or CLI
- richer doctor checks for helper installation and runtime directory permissions

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
