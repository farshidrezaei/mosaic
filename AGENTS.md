# AI Agent Guide

This file is for AI coding agents and automated contributors working on Mosaic.

Human contributors should start with `README.md` and `CONTRIBUTING.md`.

## Project Summary

Mosaic is a Go library for ABR encoding to HLS and DASH CMAF.

Core flow:

```text
Job -> probe -> ladder -> optimize -> encoder -> FFmpeg
```

Do not bypass this flow unless the task explicitly requires it.

## Required Reading Before Changes

For code changes, read:

- `README.md`
- `docs/ARCHITECTURE.md`
- `docs/ENCODING.md`
- the package file you are editing
- the package tests next to that file

For public API changes, also read:

- `docs/API.md`
- `CONTRIBUTING.md`
- `CHANGELOG.md`

## Repository Rules

- Keep dependency direction simple. Lower-level packages should not import the root `mosaic` package.
- Preserve executor-based testability. Do not call `exec.Command` directly outside `internal/executor`.
- Use `executor.CommandExecutor` for any FFmpeg or FFprobe behavior.
- Keep output behavior documented when it changes.
- Do not introduce unrelated refactors in bug-fix changes.
- Keep exported Go identifiers documented.
- Run `gofmt` on changed Go files.
- Always run `golangci-lint run` and ensure zero lint issues before committing.

## Behavioral Contracts

### Aspect Ratio

Ladders preserve source display aspect ratio.

Do not reintroduce fixed 16:9 output sizing for square, portrait, or non-standard inputs.

### Orientation

Probe and ladder logic use display dimensions, not raw stored dimensions.

When normalization is enabled:

- frames are physically rotated when needed
- rotation metadata is cleared
- output is verified

Do not remove output rotation metadata clearing unless replacing it with a tested equivalent.

### Output Directories

HLS and DASH encoders create `outDir` automatically.

Tests should use `t.TempDir()` rather than root-level paths like `/output`.

### Tests

Prefer mock executor tests for command construction.

Use real FFmpeg smoke tests only when necessary and keep generated files under `/tmp`.

## Standard Commands (Must all pass before commit)

```bash
gofmt -w <changed-go-files>
GOCACHE=/tmp/go-build go test -v -race ./...
GOCACHE=/tmp/go-build go vet ./...
golangci-lint run
```

## Documentation Requirements

Update docs when behavior changes.

Common mappings:

| Change Type | Docs To Update |
|-------------|----------------|
| Public API | `README.md`, `docs/API.md`, `CHANGELOG.md` |
| Encode behavior | `README.md`, `docs/ENCODING.md`, `CHANGELOG.md` |
| Package structure | `STRUCTURE.md`, `docs/ARCHITECTURE.md` |
| Test workflow | `docs/TESTING.md`, `CONTRIBUTING.md` |
| User-facing bug fix | `CHANGELOG.md`, relevant troubleshooting docs |

## Git Hygiene

- Do not stage unrelated local files.
- Check `git status --short` before committing.
- If unrelated files are dirty, leave them alone.
- Commit only files required for the requested change.

## Useful Smoke Tests

Square input should produce square rungs:

```text
1080x1080 -> 1080x1080, 720x720, 360x360
```

Rotated portrait input with stored `1920x1080` and rotation `-90` should produce portrait rungs:

```text
608x1080, 404x720, 202x360
```

Non-standard landscape input `1280x718` should not be forced to `640x360`; it should preserve ratio approximately:

```text
642x360
```

## Common Mistakes

- Adding HLS `pad` back into the filter graph.
- Assuming raw `Width` and `Height` are display dimensions.
- Creating tests that write to `/output`.
- Depending on real FFmpeg in a unit test when a mock executor is enough.
- Updating README but forgetting `CHANGELOG.md`.
- Committing changes without running `golangci-lint run`.
