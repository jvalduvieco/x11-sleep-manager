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
- daemon startup and signal handling in `cmd/x11-sleep-manager`
- CLI `status` command in `cmd/x11smctl`

Not implemented yet:

- inhibitor discovery or matching
- session registration from CLI
- X11 command execution
- `doctor`, `enable`, `disable`, `reconcile`, or config mutation endpoints
- helper process control such as `xss-lock` pause/resume

## Current Control API

- `GET /v1/status`

Transport details:

- HTTP over Unix domain socket
- socket path comes from config
- socket file mode is `0600`

## Current Status Model

`internal/state` exposes these daemon states:

- `disabled`
- `idle`
- `inhibited`
- `degraded`

The initial daemon starts in `idle`.

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
