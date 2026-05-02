# Next Session Checklist

## Read First

1. `docs/handover.md`
2. `docs/spec.md`
3. `docs/implementation-plan.md`
4. `docs/testing.md`

## First Coding Tasks

1. create `internal/` packages from the planned layout
2. implement config defaults and loading
3. implement daemon state model
4. implement control socket skeleton with `GET /v1/status`
5. add tests for the above

## First Non-Goals During Coding

- do not add real `xss-lock` pause/resume yet
- do not add broad shell-command hooks as the primary design
- do not overbuild packaging before the first end-to-end reconcile loop exists

## Done Criteria For First Slice

- `go test ./...` passes
- `go build ./...` passes
- daemon can report status over Unix socket
- daemon can scan and match inhibitors through an abstraction
- daemon can disable and restore `xset` state through an abstraction
- tests exist for success and failure paths of implemented behavior
