# Implementation Plan

## Principles

- Prefer a small, explicit state machine.
- Keep `logind` observation, X11 mutation, and control-plane code separate.
- Make every side effect observable.
- Reconcile state periodically instead of trusting notifications alone.
- Start with i3/X11 defaults but keep policy configurable.
- Add tests in the same phase as the code they verify.

## Proposed Package Layout

```text
cmd/
  x11-sleep-manager/
    main.go
  x11smctl/
    main.go
internal/
  app/
  config/
  control/
  dbuswatch/
  x11state/
  processctl/
  matcher/
  observe/
  state/
```

## Phase 1: Bootstrap

1. Initialize Go module and binary entrypoints.
2. Add config loading from file plus defaults.
3. Add structured logging.
4. Add a basic daemon lifecycle with signal handling.
5. Add a Unix socket HTTP control server skeleton.
6. Add the initial test harness, shared fakes, and `go test ./...` wiring.

Deliverable:

- daemon starts, listens on socket, and serves `status`
- bootstrap tests cover config loading defaults and control server startup behavior

## Phase 2: `logind` Observation

1. Connect to the system bus.
2. Implement inhibitor enumeration against `org.freedesktop.login1`.
3. Define normalized inhibitor model.
4. Add configurable matching rules.
5. Add periodic reconcile loop.
6. Add signal-triggered re-query path.
7. Add matcher and reconcile-loop tests using a fake inhibitor source.

Deliverable:

- daemon reports matching inhibitors accurately without applying X11 changes yet
- tests cover matching semantics, reconcile transitions, and missed-signal recovery behavior

## Phase 3: X11 State Management

1. Implement X11 environment validation.
2. Implement `xset q` parser for current screensaver and DPMS state.
3. Implement save-and-restore behavior for screensaver and DPMS.
4. Implement idempotent action application.
5. Implement best-effort shutdown restore.
6. Add parser and restore tests using a fake X11 command runner.

Deliverable:

- daemon can disable and restore `xset` state safely
- tests cover `xset q` parsing, action idempotency, and restoration after failures

## Phase 4: Helper Process Control

1. Discover configured helper processes such as `xss-lock`.
2. Add pause/resume support using `SIGSTOP` and `SIGCONT`.
3. Track which processes this daemon actually paused.
4. Expose helper-process status via control API.
5. Add process-control tests with a fake process table.

Deliverable:

- daemon can pause and resume `xss-lock` predictably
- tests cover only-paused-by-us semantics and repeated pause/resume calls

## Phase 5: State Machine Integration

1. Introduce explicit states: `disabled`, `idle`, `inhibited`, `degraded`.
2. Apply side effects only on transitions.
3. Preserve a last-known-good snapshot for restore attempts.
4. Surface partial-failure conditions in state and status.
5. Add end-to-end state-machine tests across idle, inhibited, disabled, and degraded modes.

Deliverable:

- end-to-end inhibit and restore behavior works for i3/X11
- tests verify no duplicate side effects and correct degraded-state behavior

## Phase 6: CLI

1. Implement `status`.
2. Implement `config get`.
3. Implement `config set` for a constrained set of runtime fields.
4. Implement `enable`, `disable`, and `reconcile`.
5. Implement `events` and `doctor`.
6. Add JSON output mode.
7. Add CLI tests for human-readable and JSON output paths.

Deliverable:

- operator can inspect and control the daemon locally
- tests cover status, config, reconcile, and doctor command behavior

## Phase 7: Observability

1. Add event ring buffer.
2. Add counters and timestamps to status.
3. Add clear log event taxonomy.
4. Add optional metrics endpoint or Prometheus text rendering from CLI.
5. Add startup diagnostics and doctor checks.
6. Add tests for event buffering and metrics rendering.

Deliverable:

- failures and state transitions are easy to understand
- tests verify event ordering, truncation, and metrics snapshot consistency

## Phase 8: Packaging and Session Integration

1. Add example config.
2. Add `systemd --user` service example.
3. Add install instructions for i3 users.
4. Document environment requirements: `DISPLAY`, `XAUTHORITY`, `DBUS_SESSION_BUS_ADDRESS` if needed.
5. Add smoke-test instructions for service-level validation.

Deliverable:

- project is ready for use in a real i3 session
- release checklist includes automated tests plus manual i3 validation

## Control API Sketch

### `GET /v1/status`

Returns:

- current state
- effective config summary
- matching inhibitor count
- matching inhibitor details
- last reconcile result
- active X11 overrides
- helper process state
- last errors
- recent metrics snapshot

### `GET /v1/config`

Returns effective runtime config.

### `POST /v1/config`

Accepts partial config updates for runtime-adjustable fields.

Fields initially safe to update at runtime:

- match rules
- reconcile interval
- logging level
- helper process names

Fields likely requiring caution or restart:

- socket path
- low-level X11 backend mode

### `POST /v1/enable`

Turns enforcement on without changing stored config defaults unless explicitly requested.

### `POST /v1/disable`

Turns enforcement off and restores session state.

### `POST /v1/reconcile`

Forces an immediate re-query of inhibitors and re-evaluation of side effects.

### `GET /v1/events`

Returns the most recent event records.

## Doctor Command Checks

`x11smctl doctor` should verify:

- daemon socket reachable
- system bus reachable
- `org.freedesktop.login1` reachable
- `DISPLAY` set
- `xset` available and callable
- `xset q` parseable
- configured helper processes installed or running
- runtime directory writable

## Testing Strategy

1. Unit tests
   - config parsing
   - inhibitor matcher
   - `xset q` parsing
   - state machine transitions
   - control API request handling

2. Integration tests
   - fake D-Bus provider or interface abstraction
   - fake X11 command runner
   - fake process table for helper pause/resume

3. Manual validation on i3
   - start daemon in user session
   - run `systemd-inhibit --what=sleep:idle --who=OpenCode sleep infinity`
   - confirm `xset` and `xss-lock` behavior changes
   - confirm restoration on inhibitor exit

## Test Gating

Every implementation slice should satisfy all of the following before being considered complete:

- `go test ./...` passes
- new behavior is covered by either unit tests or integration-style tests
- failure paths are covered for new side effects
- manual verification steps are documented when the feature depends on a live i3/X11 session

## Risks

- D-Bus inhibitor change observation may require both signal subscription and periodic polling.
- Restoring exact prior X11 state can be tricky if other tools mutate it concurrently.
- Process-name-based pause/resume may be too coarse for some users.
- User service environment propagation for `DISPLAY` and `XAUTHORITY` may vary across setups.

## First Implementation Slice

The best first end-to-end slice is:

1. daemon with config and socket
2. periodic inhibitor scan with matching
3. `xset` disable/restore only
4. `status` and `doctor` CLI commands

That slice is enough to validate the core design before adding `xss-lock` process control.
