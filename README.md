# x11_sleep_manager

`x11_sleep_manager` is a planned user-session daemon and companion CLI for X11/i3 environments.

It watches `systemd-logind` inhibitor state and applies X11-compatible session policy when matching inhibitors are active. The primary goal is to bridge `systemd-inhibit` with classic X11 idle and lock tooling such as `xset`, `xss-lock`, and `slock`.

## Scope

- User-level daemon intended for i3 and other X11 window manager sessions
- `logind` inhibitor-aware behavior
- Safe enable/disable integration with X11 idle and lock commands
- Strong observability for troubleshooting session state
- Management CLI to inspect state and change runtime parameters

## Docs

- Start here: `docs/handover.md`
- Spec: `docs/spec.md`
- Implementation plan: `docs/implementation-plan.md`
- Testing strategy: `docs/testing.md`
- Current environment notes: `docs/current-environment.md`

## Proposed binaries

- `x11-sleep-manager`: daemon
- `x11smctl`: CLI

## Status

Scaffold and design docs only. Implementation is intentionally deferred.

## Next Session

The next implementation session should start with `docs/handover.md`.

## Testing Expectation

The implementation is expected to ship with automated tests from the first functional slice onward.

- Unit tests for parsing, matching, and state transitions
- Integration-style tests for D-Bus, X11 command execution, and process control via interfaces/fakes
- `go test ./...` as the baseline verification command

See `docs/testing.md` for the detailed test strategy.
