# Current Environment

## Purpose

This file captures the current i3/X11 setup that motivated the daemon design. It is not a hard dependency for the project, but it is the first real environment the implementation should support.

## Observed Session Context

- session type: `x11`
- `systemd-inhibit` is available
- `systemd-logind` is active
- `xss-lock` is running

## Observed i3 Configuration

Relevant lines from `~/.config/i3/config` at design time:

- `xset s 180 240`
- `xss-lock -- slock &`
- manual lock bindings also invoke `slock`
- manual power-off screen action uses `xset dpms force off`

## Observed X11 State

Observed via `xset q` at design time:

- screensaver timeout: `180`
- screensaver cycle: `240`
- DPMS enabled
- DPMS standby/suspend/off: `600/600/600`

## Implication

The desktop lock path is currently:

1. X11 idle timeout via `xset`
2. X11 locker helper via `xss-lock`
3. lock command via `slock`

That path is separate from `systemd-logind` inhibitors.

## Validation Target

The daemon should eventually make the following practical behavior possible:

- when a matching `systemd-inhibit` session is active, X11 idle blanking and lock behavior are suppressed
- when the matching inhibitor disappears, the prior X11 behavior is restored

## Notes For Implementation

- do not hardcode these values into the daemon logic
- save and restore observed state instead of restoring fixed values when possible
- add tests around parsing and restoration behavior before adding process-pause features
