# Handover

## Start Here

This project is intentionally at the scaffold-and-design stage. No real daemon logic exists yet.

Read these files in order:

1. `docs/handover.md`
2. `docs/spec.md`
3. `docs/implementation-plan.md`
4. `docs/testing.md`
5. `docs/current-environment.md`

## Project Goal

Build a user-session Go daemon for X11, with i3 as the primary target, that watches `systemd-logind` inhibitors and applies X11-compatible behavior so matching `systemd-inhibit` sessions can suppress screen blanking and screen locking.

## Why This Exists

The motivating issue is that `systemd-inhibit` works at the `logind` layer, but the current i3/X11 desktop setup uses a separate idle path:

- `xset` manages X11 screensaver and DPMS timers
- `xss-lock` reacts to X11 idle behavior and runs `slock`
- `systemd-inhibit` does not automatically suppress that X11 path

This daemon is meant to bridge those two layers.

## Decisions Already Made

- Language: Go
- Process model: separate daemon, not part of the OpenCode plugin
- Scope: X11 user-session daemon plus local CLI
- Primary desktop target: i3 on X11
- Source of truth for inhibitors: `systemd-logind`
- Local control plane: Unix domain socket
- Preferred control transport: HTTP over Unix socket
- First-class observability: structured logs, event buffer, status surface, doctor command
- Runtime control: CLI can inspect and update selected parameters
- Testing: mandatory from the first functional slice

## Explicit Non-Goals

- Full `swayidle` replacement
- Wayland support
- Generic desktop power manager
- Embedding the behavior inside the OpenCode plugin

## Current State Of The Repo

- `go.mod` exists
- daemon and CLI entrypoints exist as stubs only
- the key design docs are written
- no production implementation exists in `internal/` yet

## Recommended First Slice

Implement the smallest end-to-end version that proves the architecture:

1. config loading with sensible defaults
2. Unix socket control server
3. `logind` inhibitor scan with matcher
4. periodic reconcile loop
5. `xset` disable and restore only
6. `status` and `doctor` CLI commands
7. automated tests for that slice

Do not start with `xss-lock` pause/resume. Add that only after `xset` state management works.

## Immediate Implementation Priorities

1. Create the proposed `internal/` package layout.
2. Define the core domain types:
   - inhibitor model
   - config model
   - daemon state model
   - reconcile result model
3. Add interfaces for external dependencies:
   - inhibitor source
   - X11 command runner
   - process inspector/controller
   - clock if useful for state tests
4. Build the reconcile loop around those interfaces.
5. Add tests alongside each package as it appears.

## Suggested Initial Package Order

1. `internal/config`
2. `internal/state`
3. `internal/matcher`
4. `internal/observe`
5. `internal/control`
6. `internal/x11state`
7. `internal/dbuswatch`
8. `internal/app`

## Verification Baseline

Use these commands as the baseline during implementation:

```sh
go test ./...
go build ./...
```

Once real code exists, add:

```sh
go vet ./...
```

## Important Environment Assumptions

- daemon will run as `systemd --user`
- session is X11, not Wayland
- `DISPLAY` must be available
- `xset` must be callable from the daemon environment
- `org.freedesktop.login1` must be reachable on the system bus

## Current Operator Context

The current desktop setup that motivated this project is summarized in `docs/current-environment.md`.

That file captures the practical i3/X11 behavior this daemon should be validated against.
