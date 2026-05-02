# Iteration 003: Matcher And Reconcile Skeleton

## Scope Delivered

This iteration adds the first daemon-side inhibitor evaluation flow:

- normalized inhibitor model
- config-based matcher package
- inhibitor source abstraction
- reconcile method in the app layer
- periodic reconcile loop
- status fields for matching inhibitors and reconcile health
- tests for matcher behavior and reconcile state transitions

## Delivered Packages

- `internal/observe`
- `internal/matcher`

## Inhibitor Model

The normalized inhibitor shape currently contains:

- `what`
- `who`
- `why`
- `uid`

This is enough for the current matching policy and status output.

## Current Matcher Behavior

The matcher applies the repository's default policy model:

1. if config says `uid: self`, inhibitor UID must equal the daemon user's UID
2. `who` must match one of the configured values exactly
3. `what_any` matches colon-delimited `what` tokens such as `sleep:idle`

Example:

- `what: sleep:idle` matches default config
- `what: shutdown` does not

## Reconcile Behavior

`internal/app` now exposes a reconcile path that:

1. asks the configured inhibitor source for the current inhibitor list
2. filters inhibitors through `internal/matcher`
3. records reconcile success or failure in the state store
4. transitions daemon state to:
   - `inhibited` when matches exist
   - `idle` when no matches exist
   - `degraded` when source listing fails

The daemon starts a periodic reconcile loop after the control server starts.

The loop currently performs an immediate reconcile once at startup, then repeats at the configured interval.

## Current Source Wiring

The production app wiring currently uses a no-op inhibitor source.

That means:

- the reconcile loop is active
- `status` reports reconcile counts and timestamps
- matching inhibitor state remains empty until real `logind` integration is added

This is intentional for this slice. It proves the state-machine and status plumbing before connecting to D-Bus.

## Status Additions

`GET /v1/status` now includes:

- `matching_inhibitors`
- `matching_inhibitor_count`
- `last_reconcile_at`
- `reconcile_count`
- `reconcile_failures`

## Why This Slice Matters

This is the first version where the daemon actively evaluates inhibitor state instead of only serving static status.

The next X11-action slice can now consume:

- registered X11 session context
- current matching inhibitor set
- explicit reconcile outcomes

## Remaining Gaps

Still not implemented:

- actual `logind` D-Bus enumeration
- forced reconcile control endpoint or CLI command
- X11 action application and restore

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
