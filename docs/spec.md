# Specification

## Purpose

`x11_sleep_manager` is a user-session daemon for X11 environments, with i3 as the primary target. It observes `systemd-logind` inhibitor state and applies or restores X11 session behavior so that tools like `systemd-inhibit` can effectively suppress screen blanking and screen locking in setups that otherwise rely on `xset`, `xss-lock`, and lock commands such as `slock`.

The daemon must be usable independently of any single application or plugin. It should work for any matching inhibitor source, with OpenCode as an initial motivating use case.

## Goals

- Bridge `systemd-logind` inhibitor state to X11 idle and lock behavior.
- Work reliably in i3 and similar lightweight X11 sessions.
- Preserve user control over existing idle and lock commands rather than replacing them wholesale.
- Provide strong observability for debugging session behavior.
- Expose a local CLI for querying state and changing runtime configuration.
- Run as a `systemd --user` service.

## Non-Goals

- Replacing the entire X11 idle ecosystem with a full `swayidle` clone.
- Managing Wayland sessions.
- Becoming a generic system-wide power manager.
- Owning suspend, hibernate, or shutdown policy beyond responding to inhibitor state.

## Runtime Model

The project consists of two binaries:

- `x11-sleep-manager`: long-running daemon
- `x11smctl`: local control and inspection CLI

The daemon runs in the user session and requires access to:

- system bus D-Bus for `org.freedesktop.login1`
- X11 session environment, including `DISPLAY`
- the current user's authority to invoke X11 tools such as `xset`

The CLI communicates with the daemon over a local control interface.

## Primary Behavior

### Inhibitor Watch

The daemon must determine whether any active inhibitor matches the configured policy.

Matching inputs include:

- inhibitor `what` values such as `idle` and `sleep`
- inhibitor `who`
- inhibitor `why`
- inhibitor owner UID
- optional regex or exact-match filters

The daemon should support a default matcher suitable for the motivating use case:

- current user only
- `what` contains `idle` or `sleep`
- `who` equals `OpenCode`

### Active State Transition

When the first matching inhibitor becomes active, the daemon enters `inhibited` state and applies configured actions.

Default i3/X11-oriented actions should be possible:

- disable X11 screensaver via `xset s off`
- disable DPMS via `xset -dpms`
- optionally pause a locker helper such as `xss-lock`

### Inactive State Transition

When the last matching inhibitor disappears, the daemon exits `inhibited` state and restores prior X11 behavior.

Restoration options must include:

- restoring saved `xset`-derived screensaver settings
- restoring DPMS state
- resuming previously paused helper processes

### Reconciliation

The daemon must not rely solely on edge-triggered events. It must support periodic reconciliation so missed D-Bus signals or transient failures do not leave the session in the wrong state.

## i3 Compatibility Requirements

The daemon must support common i3/X11 setups where:

- `xset s <timeout> <cycle>` manages X11 idle timers
- `xss-lock` invokes `slock`, `i3lock`, or similar
- monitor blanking uses DPMS

The daemon must not assume i3 itself provides idle notifications.

The daemon should work by orchestrating existing user tools rather than depending on i3-specific IPC.

## Control Interface

The daemon must expose a local control API over a Unix domain socket.

The socket API should support at least:

- `GET /v1/status`
- `GET /v1/config`
- `POST /v1/config`
- `POST /v1/reconcile`
- `POST /v1/enable`
- `POST /v1/disable`
- `GET /v1/events`

The transport may be HTTP-over-unix-socket or a simpler RPC protocol. HTTP-over-unix-socket is preferred for easier debugging.

## CLI Requirements

`x11smctl` must support:

- `status`: show current daemon state, matched inhibitors, last transition time, active actions, last errors
- `config get`: print effective config
- `config set`: change runtime config parameters
- `enable`: force-enable daemon behavior for testing
- `disable`: force-disable daemon behavior for testing
- `reconcile`: force an immediate rescan of inhibitors and state
- `events`: stream or print recent state transitions and warnings
- `doctor`: run environment checks for X11, D-Bus, `xset`, and locker helpers

The CLI should support JSON output for scripting.

## Configuration Model

Configuration must support both file-based startup configuration and runtime overrides.

Suggested config sections:

- `match`
- `x11`
- `processes`
- `reconcile`
- `logging`
- `socket`

Example shape:

```json
{
  "match": {
    "uid": "self",
    "who": ["OpenCode"],
    "what_any": ["idle", "sleep"],
    "why_regex": ""
  },
  "x11": {
    "disable_screensaver": true,
    "disable_dpms": true,
    "restore_previous_state": true
  },
  "processes": {
    "pause": ["xss-lock"],
    "resume": ["xss-lock"]
  },
  "reconcile": {
    "interval": "15s"
  },
  "logging": {
    "level": "info",
    "format": "json"
  },
  "socket": {
    "path": "$XDG_RUNTIME_DIR/x11_sleep_manager.sock"
  }
}
```

## State Model

The daemon must maintain explicit state:

- `disabled`: daemon is running but not enforcing
- `idle`: no matching inhibitors are active
- `inhibited`: one or more matching inhibitors are active and actions are applied
- `degraded`: daemon is running but some actions failed or state could not be fully restored

State transitions must be logged and exposed via the control interface.

## Observability Requirements

The daemon must prioritize observability.

Required observability features:

- structured logs with stable event names
- recent in-memory event ring buffer exposed via CLI/API
- explicit state transition logs
- action execution result logs
- reconciliation summary logs
- last-known matching inhibitor set in status output
- startup environment diagnostics

Recommended metrics to expose via status and optionally Prometheus text output:

- current state
- active matching inhibitor count
- total transitions to `inhibited`
- total transitions to `idle`
- reconciliation count
- reconciliation failures
- action failures by action type
- restore failures
- last successful reconcile time

## Testing Requirements

Automated tests are required, not optional.

Minimum expectations:

- unit tests for config parsing, inhibitor matching, `xset q` parsing, state transitions, and control handlers
- integration-style tests built on interfaces or fakes for D-Bus observation, X11 command execution, and helper process control
- regression tests for restore behavior after partial failures
- CLI tests for machine-readable output on `status`, `config get`, and `doctor`

The project should prefer deterministic tests over environment-dependent tests. Real X11 and D-Bus manual validation is still required before release, but must not be the only verification path.

## Failure Handling

The daemon must fail safe and remain inspectable.

Requirements:

- action failures must not crash the daemon
- repeated reconcile failures must move state to `degraded`
- shutdown should attempt to restore any altered X11 state
- partial restore failures must be visible in status and logs
- the daemon must tolerate missing optional helpers such as `xss-lock`

## Security Considerations

- Control socket must be user-local and not world-accessible.
- The daemon should not require root privileges.
- Runtime configuration changes should be limited to the current user.
- The daemon should avoid shell execution where direct argument invocation is sufficient.
- The daemon must record which commands it ran and why.

## Deployment

The primary deployment target is a `systemd --user` unit.

The unit should:

- start after the graphical session is available
- inherit or be provided `DISPLAY`
- use `XDG_RUNTIME_DIR` for the control socket
- restart on failure

## Acceptance Criteria

- On i3/X11, when a matching `systemd-inhibit` session appears, the daemon suppresses configured X11 idle and lock behavior.
- When the inhibitor disappears, the daemon restores previous behavior.
- The user can inspect the daemon using `x11smctl status`.
- The user can run `x11smctl doctor` to diagnose environment problems.
- The user can adjust runtime parameters without restarting the daemon.
- The daemon remains understandable from logs and status output when something fails.
- `go test ./...` passes with meaningful unit and integration-style coverage for the implemented feature set.
