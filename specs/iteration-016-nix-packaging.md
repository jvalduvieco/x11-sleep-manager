# Iteration 016: Nix Packaging

## Scope Delivered

This iteration replaces the ad hoc release-generated Nix expression with a tracked package definition that CI can actually validate.

Delivered pieces:

- tracked `default.nix` at the repository root
- reproducible Go dependency hash for `buildGoModule`
- packaging for both `x11-sleep-manager` and `x11smctl`
- runtime wrapping so `xset` is available on `PATH`
- CI and release workflow updates to install Nix and run `nix-build`
- Go toolchain pin updates to `1.24` in GitHub Actions to match `go.mod`

## Problems Addressed

The previous release workflow generated a `default.nix` file on the fly and uploaded it without validating that it built.

That expression also had a concrete reproducibility issue:

- `vendorHash = ""` is not a valid fixed-output dependency hash for `buildGoModule`

The result was a Nix packaging path that looked present in releases but was not maintained or exercised.

## Packaging Shape

`default.nix` now:

- uses Go 1.24-compatible builders
- builds both commands from `cmd/x11-sleep-manager` and `cmd/x11smctl`
- wraps both binaries with `xorg.xset` available on `PATH`
- exposes standard package metadata for Linux

## Workflow Changes

`validate.yml` now:

- installs Nix
- runs `nix-build default.nix --argstr version dev`

`release.yml` now:

- installs Go 1.24 explicitly
- installs Nix
- builds the tracked `default.nix` instead of generating one inline
- continues uploading `default.nix` as the release recipe artifact

## Verification For This Iteration

- `go test ./...`
- `go build ./...`

Local `nix-build` validation could not be completed in the current environment because Nix does not have permission to create entries under `/nix/store` here. The CI workflow now performs that validation on GitHub-hosted runners.
