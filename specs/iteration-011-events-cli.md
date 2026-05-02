# Iteration 011: Events CLI

## Scope Delivered

This iteration adds a CLI surface for the existing events API.

Delivered pieces:

- `x11smctl events`
- shared JSON GET rendering path reused by `status`, `events`, and `config get`

## Behavior

`x11smctl events` fetches `GET /v1/events` and prints the current event buffer as formatted JSON.

This is a simple read-only command intended to make recent daemon behavior visible without manual socket probing.

## Why This Slice Matters

The previous iteration added the event buffer but left it API-only.

This command completes the operator path so recent activity can be inspected locally with the same CLI that already handles status, doctor, and runtime control.

## Manual Validation Addition

Useful sequence:

```sh
x11smctl events
x11smctl disable
x11smctl enable
x11smctl events
```

Expected checks:

1. recent entries include runtime control events
2. state transition events appear around inhibit/idle/disabled changes

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
