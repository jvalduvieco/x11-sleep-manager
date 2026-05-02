# Iteration 001: Bootstrap Control Plane

## Scope Delivered

This iteration delivers the first executable slice of the project:

- daemon config loading with defaults and validation
- explicit daemon state store and status snapshot
- Unix-socket HTTP control server
- `GET /v1/status`
- daemon startup and graceful shutdown
- CLI `status` command
- automated tests for config, state, and control server behavior

## Implemented Packages

- `internal/config`
- `internal/state`
- `internal/control`
- `internal/app`

## Behavior

### Daemon

`cmd/x11-sleep-manager` now:

1. parses `--config`
2. loads config from JSON or defaults
3. starts the Unix-socket control server
4. waits for `SIGINT` or `SIGTERM`
5. shuts down the server cleanly and removes the socket

### CLI

`cmd/x11smctl` now supports:

- `status`

It connects to the daemon over the Unix socket and prints formatted JSON.

### Socket Handling

- parent socket directory is created automatically
- stale socket path is removed before listening
- socket mode is set to `0600`
- socket file is removed on shutdown

## Current Status Payload

The status payload currently contains:

- daemon state
- daemon start time
- last transition time
- effective config
- version string
- session readiness flag
- last error string

## Intentional Omissions

This iteration does not yet include:

- inhibitor scanning
- matcher logic in use by the daemon
- session registration API
- X11 command execution
- event ring buffer
- runtime config mutation
- doctor checks

## Why This Slice Matters

This establishes the control-plane foundation needed for all later work:

- stable process lifecycle
- inspectable daemon status
- a testable server boundary
- a central in-memory state model

## Follow-up Dependencies

The next practical slice should add:

1. session registration from `x11smctl`
2. session data surfaced in `status`
3. inhibitor source abstraction and matcher package
4. reconcile loop skeleton

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
