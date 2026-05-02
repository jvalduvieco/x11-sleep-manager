# Testing Strategy

## Goal

Testing must make the daemon safe to evolve without depending exclusively on a live i3/X11 session.

## Required Layers

### Unit Tests

Cover deterministic logic with no external dependencies:

- config parsing and defaults
- inhibitor matcher behavior
- D-Bus payload normalization
- `xset q` output parsing
- state machine transitions
- control API request validation
- CLI output rendering

### Integration-Style Tests

Use interfaces and fakes instead of the real environment:

- fake inhibitor source instead of real `logind`
- fake X11 runner instead of real `xset`
- fake process table instead of real `xss-lock`
- in-memory or temporary Unix socket control server

These tests should cover:

- reconcile cycles
- activation and restoration behavior
- degraded-state transitions
- repeated idempotent calls
- restart and restore semantics

### Manual Validation

Manual validation is still required for release candidates and environment changes.

Baseline manual flow on i3/X11:

1. Start the daemon in the user session.
2. Verify `x11smctl doctor` reports the environment as healthy.
3. Start a matching inhibitor, for example:
   `systemd-inhibit --what=sleep:idle --who=OpenCode sleep infinity`
4. Confirm the daemon transitions to `inhibited`.
5. Confirm `xset` and configured helper-process behavior changes as expected.
6. Stop the inhibitor.
7. Confirm restoration to the prior state.

## Initial Test Matrix

The first implementation slice should include automated tests for:

- config defaults
- matcher decisions
- reconcile loop behavior
- `xset` disable and restore logic
- `status` endpoint
- `doctor` command basic checks via fakes

## CI Baseline

At minimum, CI should run:

```sh
go test ./...
```

Recommended additions once the implementation starts:

- `gofmt -w` enforcement or `gofmt -l` check
- `go vet ./...`
- race-enabled tests for packages with concurrency-sensitive state

## Design Constraints For Testability

Implementation should preserve testability by:

- hiding external systems behind interfaces
- avoiding package-level mutable globals
- making state transitions explicit
- separating parsing from command execution
- returning structured results from reconciliation and action application

## Release Expectation

No feature should be considered complete without:

- automated coverage for success paths
- automated coverage for failure paths where side effects occur
- documented manual validation steps for real-session behavior
