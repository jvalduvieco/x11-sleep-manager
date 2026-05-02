# x11_sleep_manager

`x11_sleep_manager` is a user-session daemon and companion CLI for X11/i3 environments.

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
- Manual smoke test: `docs/manual-smoke-test.md`

## Proposed binaries

- `x11-sleep-manager`: daemon
- `x11smctl`: CLI

## Status

The daemon now has a working first implementation slice with:

- real `logind` inhibitor enumeration
- X11 session registration
- `xset`-based screensaver and DPMS suppression or restore
- helper-process pause and resume for `xss-lock`
- runtime control commands
- config inspection and constrained runtime config mutation
- event inspection and doctor diagnostics

## Examples

- example config: `examples/config.json`
- example `systemd --user` unit: `examples/systemd-user/x11-sleep-manager.service`

## Testing Expectation

The implementation is expected to ship with automated tests from the first functional slice onward.

- Unit tests for parsing, matching, and state transitions
- Integration-style tests for D-Bus, X11 command execution, and process control via interfaces/fakes
- `go test ./...` as the baseline verification command

See `docs/testing.md` for the detailed test strategy.
