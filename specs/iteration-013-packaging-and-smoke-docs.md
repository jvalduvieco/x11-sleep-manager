# Iteration 013: Packaging And Smoke Docs

## Scope Delivered

This iteration adds example deployment files and a concrete smoke-test guide for the current implementation.

Delivered pieces:

- example JSON config at `examples/config.json`
- example `systemd --user` unit at `examples/systemd-user/x11-sleep-manager.service`
- live session validation guide at `docs/manual-smoke-test.md`
- README refresh to reflect the implemented daemon rather than the original scaffold-only state

## Example Files Added

### `examples/config.json`

This captures the current default operator-facing config shape, including:

- match rules
- X11 suppression settings
- helper-process control
- reconcile interval
- logging

### `examples/systemd-user/x11-sleep-manager.service`

This provides a concrete `systemd --user` unit example using:

- `graphical-session.target`
- config file under `%h/.config/x11-sleep-manager/config.json`
- restart-on-failure behavior

## Smoke-Test Guide Added

`docs/manual-smoke-test.md` now documents the current end-to-end validation flow covering:

- `doctor`
- session registration
- status and events inspection
- live inhibitor testing with `systemd-inhibit`
- X11 state validation with `xset q`
- runtime enable/disable/reconcile checks
- runtime config mutation checks

## Why This Slice Matters

The codebase now has enough working functionality that example operator files and a current smoke-test guide are more valuable than further speculative design notes.

This slice makes it easier to:

- run the daemon in a real user session
- verify what has been implemented
- avoid stale setup assumptions while manual testing

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
