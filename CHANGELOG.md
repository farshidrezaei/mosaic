# Changelog

All notable changes to this project are documented here.

## [Unreleased]

### Added

- Orientation metadata support in probing (`rotation` from FFprobe side data/tags).
- Orientation-aware helpers on `probe.VideoInfo` (`DisplayWidth`, `DisplayHeight`, `IsPortrait`).
- Portrait/rotated portrait ladder handling in `ladder.Build`.
- New tests for orientation detection and portrait ladder generation.
- Documentation freshness policy in `CONTRIBUTING.md`.

### Changed

- Refreshed `README.md`, `STRUCTURE.md`, `ROADMAP.md`, and `CONTRIBUTING.md` to match current API and behavior.
- Updated documented Go baseline to align with module declaration (`go 1.25`).

### Fixed

- Removed stale or incorrect API/docs statements (notably return signatures and outdated feature claims).
