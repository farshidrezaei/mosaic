# Changelog

All notable changes to this project are documented here.

## [v1.8.0] - 2026-08-31

### Added

- Next-Gen VitePress documentation portal with SSG sitemap generation, full English & Persian (RTL) dual-language translations, interactive Vue 3 ABR ladder calculator component, and interactive in-browser stream player.
- Added `storage` package, `WithS3Upload()` functional option, and `--s3-*` CLI flags for zero-dependency direct streaming asset uploads to S3, MinIO, and Cloudflare R2 using pure Go AWS SigV4 signer, concurrent worker pool, and optimal MIME/cache headers.
- Added Next-Gen Codec support with `WithCodec()`, `WithHEVC()`, `WithAV1()` options and `--codec` CLI flag (supporting `libsvtav1`, `libx265`, `libx264` and hardware encoders).
- Added `WithCRF()` option and `--crf` CLI flag for capped-CRF content-aware bitrate optimization.
- Added `watermark` package, `WithWatermark()` option, and `--watermark` CLI flag to dynamically overlay logos/watermarks with customized positioning (top-right, top-left, bottom-right, bottom-left, center), auto-scaling per rendition, and alpha opacity.
- Added `encryption` package, `WithAES128Encryption()` option, and `--encrypt-aes128` CLI flag for automated cryptographic key generation and HLS AES-128 segment envelope encryption (`#EXT-X-KEY:METHOD=AES-128`).
- Added `subtitles` package and `WithSubtitles()` functional option for WebVTT and SRT subtitle conversion and injection into HLS `#EXT-X-MEDIA:TYPE=SUBTITLES` and DASH AdaptationSets.
- Added `WithNormalizeAudio()` option and `--normalize-audio` CLI flag applying EBU R128 (`loudnorm=I=-16:TP=-1.5:LRA=11`) broadcast audio volume standardization.
- Added `thumbnail` package and `WithThumbnails()` functional option to automatically generate storyboard sprite sheets and standard WebVTT cue files (`thumbnails.vtt`) for timeline scrubber previews in video players.
- Added `preview` package and `mosaic preview [dir]` CLI subcommand launching a local dark-mode HTTP preview player (supporting HLS.js, Dash.js, quality switching, and live stream telemetry).
- Added `WithIFrames()` functional option and `--iframes` CLI flag to generate I-frame-only trick-play playlists (`#EXT-X-I-FRAMES-ONLY`) for HLS.
- Added `docs/ROADMAP_PROGRESS.md` to track implementation progress across all roadmap phases.

### Fixed

- Fixed DASH conflicting stream aspect ratios error on portrait and non-standard videos by explicitly configuring DAR (`setdar`) across all ladder rungs in the filter complex.
- Fixed HLS live profile failure caused by unrecognized `-hls_part_size` option in modern FFmpeg versions.
- Removed deprecated `-vsync vfr` option from thumbnail generation command to ensure compatibility with modern FFmpeg (v7, v8, v9).

## [v1.7.3] - 2026-08-25

### Added

- Added GoReleaser configuration (`.goreleaser.yaml`) and GitHub Actions release workflow (`.github/workflows/release.yml`) for automated multi-platform cross-compilation (Linux, macOS, Windows on amd64 and arm64) and Release Asset generation on tag push.
- Added dynamic version injection support via ldflags in `cmd/mosaic`.

## [v1.7.2] - 2026-08-25

### Fixed

- Fixed real-time progress reporting by accumulating FFmpeg key-value progress lines into complete blocks (`StreamProgress`) emitted at `progress=continue` and `progress=end`. Prevents partial single-line emission where metrics (`out_time`, `speed`, `bitrate`, `Percentage`) appeared empty.
- Preserved last known valid metrics when FFmpeg emits `N/A` at `progress=end`, ensuring final percentage accurately reflects `100.0%`.

## [v1.7.1] - 2026-08-25

### Fixed

- Fixed `file name too long` error when normalizing orientation on complex remote URLs with URL-encoded query strings (e.g. proxy URLs, signed URLs, CDN tokens). `normalizedInputExt` now cleanly parses URL paths and restricts intermediate file extensions to safe media types.

## [v1.7.0] - 2026-08-25

### Added

- Real-time progress percentage calculation in `ProgressInfo.Percentage` computed from probed media duration and encoded timestamps.
- Added `Duration` metadata field to `probe.VideoInfo`.
- Added `BFrames` configuration to `config.Profile` (defaults to `0`) and `WithBFrames(n)` functional option.
- Added optional `WithScaleBitrateWithFPS()` option to dynamically adjust bitrate caps for high-framerate (>30 FPS) content.
- Added `AddSequentialResponse` and queue support to `executor.MockCommandExecutor` for multi-step mock testing.
- Added maintainer contact email and GitHub Security Advisories link in `SECURITY.md`.

### Changed

- Unified DASH encoding pipeline to use single-pass `filter_complex` (`split -> scale -> setsar=1`) matching HLS behavior.
- Added `-bf` flag to DASH video stream arguments.
- Enforced `golangci-lint run` as a mandatory pre-commit contract across all documentation.

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
