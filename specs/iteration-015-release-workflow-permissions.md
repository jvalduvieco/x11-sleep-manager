# Iteration 015: Release Workflow Permissions

## Scope Delivered

This iteration fixes the GitHub release workflow so tag-triggered runs can create a GitHub Release with the default Actions token.

Delivered pieces:

- explicit `contents: write` permission for the release workflow
- no change to the release trigger or artifact upload behavior

## Why This Slice Matters

`softprops/action-gh-release` needs release creation permission on repository contents.
Without an explicit write grant, some repositories and organizations provide a read-only `GITHUB_TOKEN`, which causes release creation to fail with `403 Resource not accessible by integration`.

## Verification For This Iteration

- workflow file reviewed for tag trigger plus release permission grant
- `go test ./...`
- `go build ./...`
