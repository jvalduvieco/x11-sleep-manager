# Iteration 005: X11 State Management

## Scope Delivered

This iteration is the first end-to-end X11 action slice.

Delivered pieces:

- `x11` config section with defaults
- `xset q` parser
- X11 runner abstraction
- save/apply/restore controller for screensaver and DPMS state
- reconcile integration that applies X11 overrides when matching inhibitors exist
- restore on inhibitor exit and best-effort restore on shutdown
- tests for parsing, apply/restore behavior, and app-level integration

## Config Added

The config model now includes:

```json
"x11": {
  "disable_screensaver": true,
  "disable_dpms": true,
  "restore_previous_state": true
}
```

Current defaults:

- `disable_screensaver: true`
- `disable_dpms: true`
- `restore_previous_state: true`

## X11 State Model

`internal/x11state` currently snapshots:

- screensaver timeout
- screensaver cycle
- DPMS enabled/disabled state

Current implementation intentionally restores only what it changes:

- screensaver is restored with `xset s <timeout> <cycle>`
- DPMS is restored with `xset +dpms` or `xset -dpms`

## Reconcile Integration

When the first matching inhibitor appears:

1. daemon requires a registered session
2. daemon runs `xset q`
3. daemon saves current state
4. daemon applies:
   - `xset s off`
   - `xset -dpms`
5. daemon enters `inhibited`

When the last matching inhibitor disappears:

1. daemon restores the saved `xset` state
2. daemon clears active override state
3. daemon returns to `idle`

The controller is idempotent, so repeated reconcile cycles while already inhibited do not re-run the same override sequence.

## Failure Behavior

If matching inhibitors exist but there is no registered session:

- daemon transitions to `degraded`
- last error explains that no session is registered

If `xset` query or apply/restore fails:

- daemon transitions to `degraded`
- last error records the action failure

## Status Additions

`GET /v1/status` now also includes:

- `x11_overrides_active`

This indicates whether the daemon currently believes it has active X11 suppression in place.

## Manual End-To-End Validation

When you are back, the practical validation flow for a real i3/X11 session is:

1. build the binaries
2. start the daemon in the user session
3. run `x11smctl register-session` from the live i3 session
4. run `x11smctl status` and confirm `session_ready: true`
5. start a matching inhibitor:

```sh
systemd-inhibit --what=sleep:idle --who=OpenCode sleep infinity
```

6. run `x11smctl status` and confirm:
   - `state: inhibited`
   - `matching_inhibitor_count` is non-zero
   - `x11_overrides_active: true`
7. run `xset q` and confirm screensaver/DPMS are suppressed
8. stop the inhibitor
9. run `x11smctl status` and confirm:
   - `state: idle`
   - `x11_overrides_active: false`
10. run `xset q` again and confirm the original screensaver timeout/cycle and DPMS state are restored

## Remaining Gaps

Still not implemented:

- `doctor` command to verify `xset` availability and parseability
- helper process control for `xss-lock`
- runtime enable/disable/reconcile endpoints

## Questions To Batch Later

- whether restoring only DPMS enabled state is sufficient, or whether this daemon should also snapshot and restore DPMS standby/suspend/off timers explicitly
- whether `x11smctl register-session` should accept fallback logic for missing `XAUTHORITY` in setups that rely on the default Xauthority path

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
