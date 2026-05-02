# Manual Smoke Test

## Purpose

This document captures the current end-to-end validation flow for a live i3/X11 session.

## Preconditions

- daemon binary built and available in the user session
- CLI binary built and available in the user session
- session is X11
- `systemd-logind` reachable on the system bus
- `xset` installed
- `xss-lock` running if helper-process validation is desired

## Suggested Startup Shape

Example i3 startup lines:

```i3
exec --no-startup-id x11smctl register-session
exec --no-startup-id xset s 180 240
exec --no-startup-id xss-lock -- slock &
```

## Validation Flow

1. Start the daemon.
2. Run:

```sh
x11smctl doctor
```

3. Register the live session:

```sh
x11smctl register-session
```

4. Inspect config and baseline status:

```sh
x11smctl config get
x11smctl status
x11smctl events
```

5. Start a matching inhibitor:

```sh
systemd-inhibit --what=sleep:idle --who=OpenCode sleep infinity
```

6. Re-check status:

```sh
x11smctl status
x11smctl events
```

Expected checks while inhibitor is active:

- `state` is `inhibited`
- `matching_inhibitor_count` is non-zero
- `x11_overrides_active` is `true`
- `paused_helpers` contains `xss-lock` when helper suppression is active

7. Validate live X11 state:

```sh
xset q
```

Expected checks:

- screensaver is disabled
- DPMS is disabled

8. Stop the inhibitor.

9. Re-check status:

```sh
x11smctl status
x11smctl events
```

Expected checks after inhibitor exit:

- `state` is `idle`
- `x11_overrides_active` is `false`
- `paused_helpers` is empty

10. Validate restored X11 state:

```sh
xset q
```

Expected checks:

- original screensaver timeout/cycle restored
- original DPMS enabled state restored

11. Validate runtime controls:

```sh
x11smctl disable
x11smctl status
x11smctl enable
x11smctl reconcile
x11smctl status
```

Expected checks:

- `disable` moves daemon to `disabled`
- active X11/process side effects are cleared on disable
- `enable` plus `reconcile` restores normal evaluation behavior

12. Validate runtime config mutation:

```sh
x11smctl config set '{"match":{"uid":"self","who":["Editor"],"what_any":["idle"]}}'
x11smctl config get
x11smctl reconcile
x11smctl status
```

Expected checks:

- match policy changes are visible in config output
- reconcile immediately reflects the new matching rules

## Follow-Up Notes

- if `doctor` fails because `DISPLAY` or `XAUTHORITY` are missing, register-session and X11-side validation are expected to fail too
- if helper pause/resume does not behave as expected, inspect the live process name exposed by `/proc/*/comm`
