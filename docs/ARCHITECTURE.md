# Mosaic Architecture

Mosaic is built around a clean, modular architecture. The root package orchestrates the workflow, while specialized domain packages own focused tasks without circular dependencies.

---

## Complete Runtime Flow

```text
Job
 └─ encode.go (Orchestration Engine)
    ├─ prepareInputForEncoding (Optional NormalizeVideoOrientation)
    ├─ probe.InputWithExecutor (FFprobe video streams + audio streams + rotation)
    ├─ ladder.Build (Aspect-preserving quality rungs)
    ├─ optimize.Apply (Bitrate caps and redundant rung trimming)
    ├─ encryption.SetupKeyInfo (Optional AES-128 key generation & keyinfo)
    ├─ encoder.Encode{HLS|DASH}CMAFWithExecutor (FFmpeg single-pass packaging)
    ├─ thumbnail.GenerateWithExecutor (Optional storyboard sprite + VTT)
    ├─ subtitles.ProcessTracks (Optional SRT-to-VTT + manifest injection)
    └─ storage.UploadDirectory (Optional S3 / MinIO direct sync)
```

---

## Domain Packages & Responsibilities

### 1. `root` (`mosaic`)
- **Files**: `encode.go`, `job.go`, `orientation.go`.
- **Role**: Public API, functional options wiring, and end-to-end execution pipeline orchestration.

### 2. `probe`
- **Files**: `probe/probe.go`.
- **Role**: Runs FFprobe to extract dimensions, average framerate, audio presence, duration, and display rotation metadata. Exposes helper methods `DisplayWidth()`, `DisplayHeight()`, and `IsPortrait()`.

### 3. `ladder`
- **Files**: `ladder/ladder.go`, `ladder/types.go`.
- **Role**: Generates candidate rendition rungs based on source display height while computing width dynamically to strictly preserve the input display aspect ratio (even pixels for H.264/HEVC/AV1).

### 4. `optimize`
- **Files**: `optimize/optimize.go`, `optimize/cost.go`.
- **Role**: Applies resolution-based bitrate caps, sets VBV buffer sizes (`MaxRate * 2`), scales bitrates for high-FPS content (>30 FPS), and trims redundant rungs.

### 5. `encoder`
- **Files**: `encoder/hls_cmaf.go`, `encoder/dash_cmaf.go`, `encoder/codec.go`, `encoder/common.go`.
- **Role**: Assembles optimized single-pass FFmpeg commands for HLS and DASH CMAF, resolves software/hardware codecs (`libx264`, `libx265`, `libsvtav1`, `nvenc`, `vaapi`, `videotoolbox`), applies Capped-CRF, builds watermark filter chains, and manages progress streams.

### 6. `thumbnail`
- **Files**: `thumbnail/thumbnail.go`.
- **Role**: Generates compact JPEG sprite sheets and standard WebVTT cue files (`thumbnails.vtt`) with `#xywh=x,y,w,h` spatial tags for player scrubber previews.

### 7. `preview`
- **Files**: `preview/server.go`.
- **Role**: Embedded local HTTP server featuring a modern dark-mode HTML5 player (Hls.js / Dash.js), audio track selector, quality switcher, and live stream telemetry.

### 8. `subtitles`
- **Files**: `subtitles/subtitles.go`.
- **Role**: Auto-converts SRT subtitles to WebVTT format and injects `#EXT-X-MEDIA:TYPE=SUBTITLES` into HLS master playlists and `<AdaptationSet contentType="text">` into DASH manifests.

### 9. `watermark`
- **Files**: `watermark/watermark.go`.
- **Role**: Constructs responsive FFmpeg overlay coordinates and opacity mixers for dynamic logo watermarking across different rendition resolutions.

### 10. `encryption`
- **Files**: `encryption/encryption.go`.
- **Role**: Generates cryptographically secure 16-byte AES keys and creates standard `enc.keyinfo` files for HLS AES-128 segment protection.

### 11. `storage`
- **Files**: `storage/storage.go`.
- **Role**: Pure Go (Zero-Dependency) AWS Signature Version 4 client for concurrent direct streaming asset uploads to S3, MinIO, or Cloudflare R2 with optimal MIME and cache headers.

### 12. `internal/executor`
- **Files**: `internal/executor/executor.go`, `internal/executor/mock.go`.
- **Role**: Command execution abstraction capturing process execution time, peak memory usage (RSS), and providing a fast Mock Executor for 100% unit test coverage.

### 13. `config`
- **Files**: `config/profiles.go`.
- **Role**: Defines `Profile` presets (`VOD`, `LIVE`), `VideoCodec` enums (`CodecH264`, `CodecHEVC`, `CodecAV1`), and GPU backend constants (`GPU_NVENC`, `GPU_VAAPI`, `GPU_VIDEOTOOLBOX`).

---

## Architectural Principles

1. **Zero External Dependencies**: Mosaic relies solely on the Go standard library and standard FFmpeg/FFprobe binaries.
2. **Deterministic & Predictable**: Encoding ladders and filter graphs are calculated mathematically without guesswork.
3. **Aspect-Ratio Integrity**: Output streams never letterbox or distort source aspect ratios.
4. **Testability First**: All FFmpeg/FFprobe operations run through `executor.CommandExecutor`, allowing lightning-fast unit tests with zero real FFmpeg process overhead.
