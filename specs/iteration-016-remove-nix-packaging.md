# Iteration 016: Remove Nix Packaging

## Scope Delivered

This iteration removes the Nix packaging artifact from the GitHub release workflow and drops the remaining repository references to it.

Delivered pieces:

- removed the generated `default.nix` step from `.github/workflows/release.yml`
- stopped uploading `default.nix` as a release artifact
- updated `AGENTS.md` so the recorded release surface matches the workflow

## Why This Slice Matters

The repository no longer wants to ship or imply support for Nix packaging.

Removing the generated artifact keeps release outputs aligned with the maintained package formats and avoids publishing packaging metadata that the project does not intend to support.

## Verification For This Iteration

- `go test ./...`
- `go build ./...`
