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
- `xset q` parsing and X11 override control in `internal/x11state`
- helper process pause/resume control in `internal/processctl`
- daemon startup and signal handling in `cmd/x11-sleep-manager`
- CLI `status`, `events`, `register-session`, `reconcile`, `enable`, `disable`, `doctor`, and `config get` commands in `cmd/x11smctl`
- in-memory X11 session registration captured from the CLI environment

Not implemented yet:

- broader config mutation coverage beyond the currently supported runtime-safe sections

The repository now also includes packaging and operator examples under `examples/` and a live validation guide in `docs/manual-smoke-test.md`.

Current release workflow note:

- `.github/workflows/release.yml` requires `permissions: contents: write` so `softprops/action-gh-release` can create releases with `GITHUB_TOKEN`

## Current Control API

- `GET /v1/status`
- `GET /v1/config`
- `GET /v1/events`
- `POST /v1/config`
- `POST /v1/session`
- `POST /v1/reconcile`
- `POST /v1/enable`
- `POST /v1/disable`

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
- whether X11 overrides are currently active
- paused helper names

## Current Event Model

The daemon now keeps a small in-memory event buffer.

Current event sources include:

- state transitions
- session registration
- reconcile failures
- recorded errors
- runtime enable/disable commands

Current transport:

- `GET /v1/events`

## Current Doctor Checks

`x11smctl doctor` currently verifies:

- daemon socket reachable
- system bus reachable
- `org.freedesktop.login1` reachable
- `DISPLAY` set
- `XAUTHORITY` set
- `xset` available and callable
- `xset q` parseable
- configured helper processes installed or running
- runtime directory writable

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
- `x11smctl events`
- `x11smctl reconcile`
- `x11smctl enable`
- `x11smctl disable`
- `x11smctl config set`
- `x11smctl config get`

Current runtime config mutation support is intentionally constrained to whole-section replacement for:

- `match`
- `x11`
- `processes`

Not currently live-mutable:

- `socket`
- `reconcile.interval`
- `logging`

Expected use:

- invoke `x11smctl register-session` from the live i3/X11 session so the daemon learns the correct X11 environment before later X11 actions are implemented

## Current X11 Behavior

`internal/x11state` currently manages `xset`-based state only.

Implemented behavior:

- parse `xset q` output for screensaver timeout/cycle and DPMS enabled state
- on first transition to `inhibited`, save current `xset` state and apply:
  - `xset s off`
  - `xset -dpms`
- on transition back to `idle`, restore the saved screensaver timeout/cycle and DPMS enabled state
- best-effort restore on daemon shutdown if overrides are still active and a session is registered

Current config defaults for `x11`:

- `disable_screensaver: true`
- `disable_dpms: true`
- `restore_previous_state: true`

Current limitation:

- X11 actions require a registered session; matching inhibitors without one move the daemon to `degraded`

## Current Helper Process Behavior

`internal/processctl` currently manages configured helper processes by process name.

Implemented behavior:

- discover matching process IDs from `/proc/*/comm`
- pause configured helper names with `SIGSTOP`
- resume only the processes this daemon previously paused using `SIGCONT`
- expose currently paused helper names in daemon status

Current config defaults for `processes`:

- `pause: ["xss-lock"]`
- `resume: ["xss-lock"]`

Current limitation:

- process discovery is name-based and Linux `/proc`-based

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

## Live Validation Results

End-to-end validation completed successfully on 2026-05-02:

- **Test**: 220s OpenCode inhibitor detection with xss-lock paused
- **Result**: Daemon remained in `inhibited` state throughout, screen stayed unlocked
- **Baseline normalization**: `x11smctl disable` → `enable` → `reconcile`
- **Final state**: inhibited (1 matching inhibitor), xss-lock paused (Tl), X11 overrides active

Verified that daemon correctly:
- Detects real OpenCode inhibitor from logind D-Bus
- Pauses xss-lock when OpenCode inhibitor present
- Disables screensaver/DPMS when inhibited
- Maintains state throughout inhibitor lifecycle

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
