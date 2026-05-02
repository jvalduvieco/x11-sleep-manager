# Iteration 006: Helper Process Control

## Scope Delivered

This iteration adds helper-process suppression alongside the existing `xset` suppression path.

Delivered pieces:

- `processes` config section with default `xss-lock` pause/resume names
- Linux process discovery by `/proc/*/comm`
- helper pause with `SIGSTOP`
- helper resume with `SIGCONT`
- only-paused-by-us semantics in the controller
- app integration so helper pause/resume follows inhibit and restore transitions
- status reporting of paused helper names

## Config Added

The config model now includes:

```json
"processes": {
  "pause": ["xss-lock"],
  "resume": ["xss-lock"]
}
```

Current defaults target `xss-lock` directly.

## Process Control Model

`internal/processctl` uses two layers:

1. an inspector that finds process IDs and sends signals
2. a controller that tracks which PIDs this daemon actually paused

Pause behavior:

- find configured helper names
- send `SIGSTOP`
- record paused PID to helper-name mapping

Resume behavior:

- send `SIGCONT` only to recorded paused PIDs
- clear the tracked paused set after successful resume

This prevents the daemon from resuming unrelated processes it did not stop itself.

## Reconcile Integration

When entering `inhibited`:

1. apply X11 overrides
2. pause configured helper processes
3. expose paused helper names in status

When returning to `idle`:

1. resume paused helper processes
2. restore X11 state
3. clear paused helper status

On daemon shutdown:

- best-effort helper resume runs before final shutdown completes

## Status Additions

`GET /v1/status` now includes:

- `paused_helpers`

This is the set of helper names that the daemon currently believes it has paused.

## Manual End-To-End Validation Additions

After confirming `x11_overrides_active: true` during the X11 validation flow, add these checks:

1. verify `paused_helpers` contains `xss-lock`
2. inspect `xss-lock` state with a process tool such as `ps` and confirm it is stopped
3. after inhibitor exit, confirm `paused_helpers` is empty
4. confirm `xss-lock` is running again

## Remaining Gaps

Still not implemented:

- `doctor` checks for helper process availability
- operator-facing runtime enable/disable/reconcile endpoints
- non-name-based helper targeting for environments where process names are ambiguous

## Questions To Batch Later

- whether helper discovery should eventually use command-line matching instead of only `/proc/*/comm`
- whether helper pause/resume should be optional by default or remain on for `xss-lock`

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
