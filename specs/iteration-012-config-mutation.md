# Iteration 012: Runtime Config Mutation

## Scope Delivered

This iteration adds the first constrained runtime config update path.

Delivered pieces:

- `POST /v1/config`
- app-level runtime config update logic
- live updates for matcher, X11, and helper-process config sections
- `x11smctl config set`
- tests for config patch application and control endpoint invocation

## Supported Runtime Patch Model

Runtime config updates currently support whole-section replacement for:

- `match`
- `x11`
- `processes`

Example patch:

```json
{
  "match": {
    "uid": "self",
    "who": ["Editor"],
    "what_any": ["idle"]
  }
}
```

The daemon applies the patch to the current in-memory config, validates the resulting config, updates live controller settings, and records a config-updated event.

## CLI Added

`x11smctl config set` now supports passing a raw JSON patch payload.

Current usage forms:

```sh
x11smctl config set '{"match":{"uid":"self","who":["Editor"],"what_any":["idle"]}}'
```

or via stdin:

```sh
printf '%s\n' '{"match":{"uid":"self","who":["Editor"],"what_any":["idle"]}}' | x11smctl config set
```

## Intentionally Excluded For Now

These fields are not currently live-mutable:

- `socket`
- `reconcile.interval`
- `logging`

Those require additional lifecycle behavior that this slice intentionally does not fake.

## Why This Slice Matters

The project now has enough moving parts that being able to adjust live match or suppression settings without restarting the daemon is useful for real-session testing.

This slice adds that ability while keeping the mutation surface limited to settings that can be applied safely and honestly with the current architecture.

## Manual Validation Additions

Useful sequence:

```sh
x11smctl config get
x11smctl config set '{"match":{"uid":"self","who":["Editor"],"what_any":["idle"]}}'
x11smctl config get
x11smctl reconcile
x11smctl status
```

Expected checks:

1. config output reflects the updated `match` section
2. reconcile uses the new matcher behavior immediately
3. events include `config.updated`

## Remaining Gaps

Still not implemented:

- fine-grained per-field patch semantics within a section
- live mutation for reconcile interval, logging, or socket path
- schema-aware CLI helpers beyond raw JSON patch payloads

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
