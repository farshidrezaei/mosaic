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

- Preserved source display aspect ratio when building output ladders instead of forcing fixed 16:9 frame sizes.
- Refreshed `README.md`, `STRUCTURE.md`, `ROADMAP.md`, and `CONTRIBUTING.md` to match current API and behavior.
- Updated documented Go baseline to align with module declaration (`go 1.25`).

### Fixed

- Cleared video rotation metadata on generated HLS/DASH renditions to prevent players from applying source rotation twice.
- Corrected orientation normalization for +/-90 degree display matrices so normalized HLS output is not rotated upside down.
- Created HLS/DASH output directories automatically before invoking FFmpeg.
- Removed stale or incorrect API/docs statements (notably return signatures and outdated feature claims).
