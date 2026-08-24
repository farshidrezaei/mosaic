# Changelog

All notable changes to this project are documented here.

## [Unreleased]

### Added

### Changed

### Fixed

## [v1.6.0] - 2026-08-24

### Added

- Job input validation with clear error messages for empty `Input` or `OutputDir`.

### Changed

- Replaced raw byte-chunk progress reading with `bufio.Scanner` for correct line-buffered FFmpeg progress parsing.
- Made `-preset medium` conditional on `libx264` codec only; GPU encoders (NVENC, VAAPI, VideoToolbox) no longer receive an incompatible preset.
- HLS playlist type is now `vod` only for VOD profiles; Live profiles omit `hls_playlist_type` to allow proper low-latency behavior.
- Audio probe errors on context cancellation now propagate instead of being silently ignored.
- Improved FPS parsing to validate both numerator and denominator before division.
- Added safe type assertion for `syscall.Rusage` in executor to prevent panic on non-Linux platforms.
- Updated CI workflow to `actions/checkout@v4`, `actions/setup-go@v5`, `golangci-lint-action@v6`, added `go vet` and `-race` flag.
- Expanded `.gitignore` with OS, editor, and temp file patterns.

### Fixed

- Fixed progress data corruption caused by byte-level reads splitting FFmpeg output lines mid-stream.
- Fixed FFmpeg failure when using GPU hardware encoders due to invalid `-preset medium` argument.
- Fixed HLS Live profile incorrectly setting `hls_playlist_type vod`.
- Fixed wrong error variable in `examples/advanced_dash` error message.

## [v1.5.2] - 2026-06-13

### Added

### Changed

### Fixed

- Restricted orientation-normalization audio mapping to the first audio stream to avoid including malformed/unsupported auxiliary audio tracks in MP4 output.

## [v1.5.1] - 2026-06-13

### Added

### Changed

### Fixed

- Prevented orientation-normalization MP4 failures on inputs with unsupported extra streams (for example data/timecode tracks) by mapping only video/audio during normalization remux/rotate steps.

## [v1.5.0] - 2026-06-08

### Added

- Orientation metadata support in probing (`rotation` from FFprobe side data/tags).
- Orientation-aware helpers on `probe.VideoInfo` (`DisplayWidth`, `DisplayHeight`, `IsPortrait`).
- Portrait/rotated portrait ladder handling in `ladder.Build`.
- New tests for orientation detection and portrait ladder generation.
- Documentation freshness policy in `CONTRIBUTING.md`.
- Complete human and AI documentation set (`docs/`, `AGENTS.md`, `SUPPORT.md`, and `SECURITY.md`).

### Changed

- Preserved source display aspect ratio when building output ladders instead of forcing fixed 16:9 frame sizes.
- Refreshed `README.md`, `STRUCTURE.md`, `ROADMAP.md`, and `CONTRIBUTING.md` to match current API and behavior.
- Updated documented Go baseline to align with module declaration (`go 1.25`).

### Fixed

- Cleared video rotation metadata on generated HLS/DASH renditions to prevent players from applying source rotation twice.
- Corrected orientation normalization for +/-90 degree display matrices so normalized HLS output is not rotated upside down.
- Created HLS/DASH output directories automatically before invoking FFmpeg.
- Removed stale or incorrect API/docs statements (notably return signatures and outdated feature claims).
