# Iteration 014: Doctor Enhancements

## Scope Delivered

This iteration expands the existing doctor command with checks that better match the current daemon's real dependencies.

Delivered pieces:

- runtime directory writability check based on the daemon socket directory
- helper-process readiness check based on the daemon's configured helper names
- tests updated for the expanded doctor report

## New Checks Added

### `runtime-dir`

This check verifies that the directory containing the daemon socket is writable by creating and removing a temporary file.

### `helpers`

This check reads the daemon's effective config through `GET /v1/config` and validates configured helper names by checking whether each helper is:

- currently running, or
- available on `PATH`

This is a pragmatic operator check for the current name-based helper model.

## Why This Slice Matters

The doctor command now covers more of the actual end-to-end failure surface:

- daemon socket access
- D-Bus access
- X11 access
- helper-process readiness
- runtime directory usability

That makes it a better first command to run before session registration and inhibitor testing.

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
