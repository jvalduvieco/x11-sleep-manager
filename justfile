set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

default:
  @just --list

fmt:
  @echo "Formatting Go files"
  gofmt -w ./cmd ./internal

test:
  @echo "Running tests"
  go test ./...

build:
  @echo "Building binaries"
  go build -o x11-sleep-manager ./cmd/x11-sleep-manager
  go build -o x11smctl ./cmd/x11smctl

check: fmt test build
