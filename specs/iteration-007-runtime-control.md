# Iteration 007: Runtime Control Endpoints

## Scope Delivered

This iteration adds explicit runtime control for manual testing and operations.

Delivered pieces:

- app-level runtime enabled/disabled state
- `POST /v1/reconcile`
- `POST /v1/enable`
- `POST /v1/disable`
- matching CLI commands in `x11smctl`
- tests for control endpoint invocation and error propagation

## Runtime Behavior

The daemon now has a runtime enabled flag separate from process lifetime.

When runtime control is disabled:

- active helper-process and X11 side effects are cleared
- daemon state becomes `disabled`
- subsequent reconcile cycles still update match metadata but do not apply side effects

When runtime control is enabled again:

- daemon immediately performs a reconcile
- matching inhibitors can re-apply helper/X11 suppression as normal

## Control API Added

New endpoints:

- `POST /v1/reconcile`
- `POST /v1/enable`
- `POST /v1/disable`

All currently return `204 No Content` on success.

## CLI Added

`x11smctl` now supports:

- `reconcile`
- `enable`
- `disable`

These are intended primarily for local operator testing and diagnosis.

## Manual Validation Additions

Useful command sequence during end-to-end testing:

```sh
x11smctl status
x11smctl disable
x11smctl status
x11smctl enable
x11smctl reconcile
x11smctl status
```

Expected checks:

1. after `disable`, status reports `state: disabled`
2. after `disable`, any active `x11_overrides_active` or `paused_helpers` state is cleared
3. after `enable` plus `reconcile`, status reflects the current live inhibitor state again

## Remaining Gaps

Still not implemented:

- `doctor` command and environment diagnostics
- config get/set endpoints
- events endpoint or ring buffer

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
