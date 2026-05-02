# Iteration 010: Event Buffer

## Scope Delivered

This iteration adds a small in-memory event history and exposes it over the control API.

Delivered pieces:

- event model in `internal/state`
- fixed-size event buffer
- `GET /v1/events`
- event endpoint tests
- basic event emission from key daemon actions

## Current Event Buffer Behavior

The daemon now stores up to 100 recent events in memory.

Current event producers:

- state transitions
- session registration
- reconcile failures
- generic recorded errors
- runtime enable/disable commands

When the buffer exceeds its limit, older events are dropped and the newest 100 are kept.

## Control API Added

New endpoint:

- `GET /v1/events`

It returns the current event list as JSON.

## Why This Slice Matters

By this point the daemon has multiple moving parts:

- logind observation
- session registration
- X11 override application
- helper-process pause/resume
- runtime control commands

The event buffer provides a lightweight timeline for recent behavior without requiring external log collection.

## Remaining Gaps

Still not implemented:

- CLI `events` command
- event severity levels or categories beyond event names/messages
- persistent event storage across daemon restarts

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
