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
