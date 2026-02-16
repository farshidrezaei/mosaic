# Contributing to Mosaic

Thanks for contributing.

## Development Setup

1. Fork and clone.
2. Install Go (`1.25+`).
3. Install FFmpeg/FFprobe.

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt install ffmpeg
```

## Build and Test

```bash
# Build all packages
go build ./...

# Run tests
go test ./...

# Coverage (if your environment blocks default build cache, override it)
GOCACHE=/tmp/go-build go test ./... -cover

# Lint (if installed)
golangci-lint run
```

## Code Guidelines

- Follow standard Go style and run `gofmt`.
- Keep packages focused and avoid circular dependencies.
- Add/adjust tests for behavior changes.
- Preserve executor-based testability (avoid hard-coding shell calls in business logic).
- Keep exported types/functions documented.

## Documentation Freshness Policy

To keep docs up to date, every behavior/API change should update Markdown in the same PR.

Required checks before merge:

1. `README.md` reflects public API and actual behavior.
2. `STRUCTURE.md` reflects package/file layout and runtime flow.
3. `CHANGELOG.md` has an `[Unreleased]` entry.
4. `ROADMAP.md` moves completed roadmap items to `Done`.
5. Commands in docs are copy-paste runnable.

If docs are intentionally unchanged, state why in the PR description.

## Pull Request Checklist

1. Create a branch.
2. Implement change and tests.
3. Run build/tests/lint.
4. Update docs + changelog.
5. Submit PR with clear scope and reasoning.

## Reporting Issues

Please include:

- reproducible steps
- expected vs actual behavior
- OS, Go version, FFmpeg version
- sample command/logs (when relevant)
