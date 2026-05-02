# AGENTS

## Purpose

This file captures durable working knowledge for future coding sessions in this repository.

## Project Summary

`x11_sleep_manager` is a Go project for an X11/i3 user-session daemon plus local CLI.

- Daemon binary: `x11-sleep-manager`
- CLI binary: `x11smctl`
- Primary goal: bridge `systemd-logind` inhibitors to X11 idle and lock behavior
- Primary desktop target: i3 on X11

## Current Implementation Boundary

The repository now contains the first bootstrap slice:

- JSON config defaults and loading in `internal/config`
- explicit daemon state in `internal/state`
- Unix-socket HTTP control server in `internal/control`
- bootstrap app wiring in `internal/app`
- normalized inhibitor model in `internal/observe`
- config-based inhibitor matching in `internal/matcher`
- periodic reconcile loop over an inhibitor source abstraction
- real `logind` D-Bus integration in `internal/observe`
- daemon startup and signal handling in `cmd/x11-sleep-manager`
- CLI `status` and `register-session` commands in `cmd/x11smctl`
- in-memory X11 session registration captured from the CLI environment

Not implemented yet:

- X11 command execution
- `doctor`, `enable`, `disable`, `reconcile`, or config mutation endpoints
- helper process control such as `xss-lock` pause/resume

## Current Control API

- `GET /v1/status`
- `POST /v1/session`

Transport details:

- HTTP over Unix domain socket
- socket path comes from config
- socket file mode is `0600`

`POST /v1/session` currently requires at least:

- `display`
- `xauthority`

Optional fields currently accepted:

- `xdg_session_type`
- `dbus_session_bus_address`

## Current Status Model

`internal/state` exposes these daemon states:

- `disabled`
- `idle`
- `inhibited`
- `degraded`

The initial daemon starts in `idle`.

Current status payload also includes:

- matching inhibitor count
- matching inhibitor details
- last reconcile timestamp
- total reconcile count
- reconcile failure count

## Current Inhibitor Model

`internal/observe` currently defines a normalized inhibitor with:

- `what`
- `who`
- `why`
- `uid`

The `NoopSource` still exists for tests and wiring experiments, but the default app wiring now uses the real system-bus `logind` source.

`internal/observe` now also provides a real system-bus `logind` source that calls `org.freedesktop.login1.Manager.ListInhibitors` and normalizes the result.

If the system bus cannot be reached at startup, app construction falls back to a static error source so the daemon can still start and report degraded reconcile status instead of failing closed.

## Current Matching Rules

`internal/matcher` currently applies these filters from config:

- `uid == self`
- exact `who` match against configured values
- token match for colon-delimited `what` values, such as `sleep:idle`

Default policy still targets:

- current user only
- `who == OpenCode`
- `what_any` contains `idle` or `sleep`

## Current Session Registration Model

The daemon stores one in-memory active session context.

It currently records:

- `DISPLAY`
- `XAUTHORITY`
- optional `XDG_SESSION_TYPE`
- optional `DBUS_SESSION_BUS_ADDRESS`
- session registration timestamp

CLI command:

- `x11smctl register-session`

Expected use:

- invoke `x11smctl register-session` from the live i3/X11 session so the daemon learns the correct X11 environment before later X11 actions are implemented

## Config Rules

Config is JSON for now.

Defaults:

- socket path: `$XDG_RUNTIME_DIR/x11_sleep_manager.sock`
- fallback socket path if `XDG_RUNTIME_DIR` is missing: `os.TempDir()/x11_sleep_manager.sock`
- match UID: `self`
- match `who`: `OpenCode`
- match `what_any`: `idle`, `sleep`
- reconcile interval: `15s`
- logging level: `info`
- logging format: `json`

## Testing Baseline

After each iteration:

1. run `gofmt -w` on touched Go files
2. run `go test ./...`
3. run `go build ./...`
4. update `AGENTS.md` if durable repo knowledge changed
5. add or update a `specs/*.md` file describing the delivered slice
6. create a meaningful git commit

## Commit Convention For This Repo

Each implementation iteration should end with:

- a focused code change
- updated knowledge capture in `AGENTS.md`
- updated implementation notes under `specs/`
- a meaningful commit message with a co-author trailer

Current co-author trailer:

`Co-authored-by: OpenCode <opencode@local>`
