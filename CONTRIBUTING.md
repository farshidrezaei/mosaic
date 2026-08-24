# Contributing to Mosaic

Thanks for contributing.

This project values small, well-tested changes that keep FFmpeg behavior explicit and documented.

## Development Setup

1. Fork and clone the repository.
2. Install Go `1.25+`.
3. Install FFmpeg and FFprobe.

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt install ffmpeg
```

Verify tools:

```bash
go version
ffmpeg -version
ffprobe -version
```

## Build and Test

```bash
# Build all packages
go build ./...

# Run tests
go test ./...

# Use a writable Go cache if your environment restricts the default cache
GOCACHE=/tmp/go-build go test ./...

# Vet
GOCACHE=/tmp/go-build go vet ./...

# Coverage
GOCACHE=/tmp/go-build go test ./... -cover

# Lint (Mandatory)
golangci-lint run
```

## Code Guidelines

- Run `gofmt` on changed Go files.
- Keep packages focused.
- Avoid circular dependencies.
- Preserve executor-based testability.
- Do not call FFmpeg or FFprobe directly outside `internal/executor`.
- Add tests for behavior changes.
- Ensure `golangci-lint run` reports 0 issues.
- Keep exported types and functions documented.
- Prefer small, reviewable commits.

## Encoding Behavior Guidelines

When changing encode behavior, preserve these contracts unless the change explicitly intends to modify them:

- Output ladders preserve source display aspect ratio.
- Raw stored dimensions are not always display dimensions.
- Orientation metadata is considered during ladder generation.
- `WithNormalizeOrientation()` physically rotates metadata-rotated sources and clears output rotation metadata.
- HLS and DASH output directories are created automatically.
- HLS should not pad into a fixed 16:9 frame.
- Tests should not write to root paths such as `/output`; use `t.TempDir()`.

## Documentation Freshness Policy

Every behavior or API change should update documentation in the same PR.

Required checks:

1. `README.md` reflects public API and actual behavior.
2. `docs/API.md` reflects public API changes.
3. `docs/ENCODING.md` reflects encode behavior changes.
4. `docs/ARCHITECTURE.md` and `STRUCTURE.md` reflect package or flow changes.
5. `docs/TESTING.md` reflects test workflow changes.
6. `CHANGELOG.md` has an `[Unreleased]` entry.
7. Commands in docs are copy-paste runnable.

If docs are intentionally unchanged, state why in the PR description.

## Pull Request Checklist

1. Create a branch.
2. Implement the change.
3. Add or update tests.
4. Run:

```bash
GOCACHE=/tmp/go-build go test -v -race ./...
GOCACHE=/tmp/go-build go vet ./...
golangci-lint run
```

5. Ensure all tests, vet, and `golangci-lint run` pass with zero issues.
6. Update docs and changelog.
7. Submit a PR with clear scope, reasoning, and test results.

## Reporting Issues

Please include:

- reproducible steps
- expected behavior
- actual behavior
- input file or URL, if shareable
- OS and architecture
- Go version
- FFmpeg and FFprobe versions
- source FFprobe output
- output FFprobe output, when relevant
- logs or error output

Useful source probe command:

```bash
ffprobe -v error -select_streams v:0 \
  -show_entries stream=width,height,avg_frame_rate:stream_tags=rotate:stream_side_data=rotation \
  -of json input.mp4
```

Useful output probe command:

```bash
ffprobe -v error -select_streams v:0 \
  -show_entries stream=width,height:stream_tags=rotate:stream_side_data=rotation \
  -of json /tmp/output/init_0.mp4
```
