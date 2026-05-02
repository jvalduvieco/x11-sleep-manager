# Development commands for x11-sleep-manager

default := "help"
set quiet

help:
    echo "Available commands:"
    echo "  just update-deps    - Update all Go dependencies to latest versions"
    echo "  just compile        - Build all binaries into bin/"
    echo "  just test           - Run all tests"
    echo "  just release [type] - Create a release (patch|minor|major), defaults to patch"
    echo "  just check          - Run format check, tests, and build"

update-deps:
    echo "Updating dependencies..."
    go get -u ./...
    go mod tidy
    echo "Dependencies updated. Review changes with: git diff go.mod go.sum"

compile:
    echo "Building binaries..."
    mkdir -p bin
    go build -o bin/x11-sleep-manager ./cmd/x11-sleep-manager
    go build -o bin/x11smctl ./cmd/x11smctl
    echo "Build complete"

test:
    go test ./...

release type="patch":
    #!/usr/bin/env bash
    set -euo pipefail

    VERSION_TYPE="{{ type }}"
    CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    MAJOR=$(echo $CURRENT_VERSION | cut -d. -f1 | tr -d v)
    MINOR=$(echo $CURRENT_VERSION | cut -d. -f2)
    PATCH=$(echo $CURRENT_VERSION | cut -d. -f3)

    case $VERSION_TYPE in
        major)
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            ;;
        minor)
            MINOR=$((MINOR + 1))
            PATCH=0
            ;;
        patch)
            PATCH=$((PATCH + 1))
            ;;
    esac

    NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
    echo "Creating release: $NEW_VERSION"

    git tag $NEW_VERSION
    git push origin $NEW_VERSION

    echo "Tagged and pushed $NEW_VERSION. GitHub Actions will build and publish packages."

check:
    echo "Running format check..."
    if [ -n "$(gofmt -l .)" ]; then \
        echo "Format errors:"; \
        gofmt -l .; \
        exit 1; \
    fi
    echo "Running tests..."
    go test ./...
    echo "Building..."
    go build ./...
    echo "All checks passed"
