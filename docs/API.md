# Mosaic API Reference

This document provides a comprehensive reference for the public API surface of `github.com/farshidrezaei/mosaic`.

## Package Import

```go
import "github.com/farshidrezaei/mosaic"
```

---

## Core Types

### Job

```go
type Job struct {
	Input           string
	OutputDir       string
	ProgressHandler ProgressHandler
	Profile         Profile
}
```

- **`Input`**: Local file path (e.g. `./video.mp4`) or remote URL (e.g. `https://example.com/video.mp4`).
- **`OutputDir`**: Target directory where manifests, playlists, init segments, and media segments are written. Created automatically (`0755`).
- **`Profile`**: `ProfileVOD` (5s segments) or `ProfileLive` (2s segments).
- **`ProgressHandler`**: Callback function receiving real-time progress events.

### ProgressInfo

```go
type ProgressInfo struct {
	CurrentTime string  // e.g. "00:01:23.456000"
	Bitrate     string  // e.g. "2450.3kbits/s"
	Speed       string  // e.g. "2.15x"
	Percentage  float64 // Exact percentage (0.0 to 100.0)
}
```

### VideoCodec

```go
type VideoCodec = config.VideoCodec

const (
	CodecH264 VideoCodec = "h264" // H.264 / AVC (libx264, nvenc, vaapi, videotoolbox)
	CodecHEVC VideoCodec = "hevc" // H.265 / HEVC (libx265, hevc_nvenc, hevc_vaapi, hevc_videotoolbox)
	CodecAV1  VideoCodec = "av1"  // AV1 (libsvtav1, av1_nvenc, av1_vaapi)
)
```

### ThumbnailConfig

```go
type ThumbnailConfig struct {
	SpriteFilename  string // Default: "thumbnails_%d.jpg"
	VTTFilename     string // Default: "thumbnails.vtt"
	IntervalSeconds int    // Interval between thumbnails in seconds (default: 2)
	TileWidth       int    // Thumbnail tile width in pixels (default: 160)
	TileHeight      int    // Thumbnail tile height in pixels (default: 90)
	Columns         int    // Columns in sprite grid (default: 5)
	Rows            int    // Rows in sprite grid (default: 5)
	Quality         int    // JPEG compression quality (1-31, default: 3)
}
```

### WatermarkConfig & WatermarkPosition

```go
type WatermarkConfig struct {
	Path     string            // Absolute or relative path to image file (PNG/WebP)
	Position WatermarkPosition // Overlay position (default: PositionTopRight)
	OffsetX  int               // Horizontal padding in pixels (default: 20)
	OffsetY  int               // Vertical padding in pixels (default: 20)
	Opacity  float64           // Alpha opacity level (0.0 - 1.0, default: 1.0)
}

type WatermarkPosition string

const (
	PositionTopRight    WatermarkPosition = "top-right"
	PositionTopLeft     WatermarkPosition = "top-left"
	PositionBottomRight WatermarkPosition = "bottom-right"
	PositionBottomLeft  WatermarkPosition = "bottom-left"
	PositionCenter      WatermarkPosition = "center"
)
```

### SubtitleTrack

```go
type SubtitleTrack struct {
	Path     string // Input subtitle file path (.vtt or .srt)
	Language string // ISO language code (e.g. "en", "fa")
	Label    string // Display label in video player (e.g. "English", "Persian")
	Default  bool   // Whether this track is the default active subtitle
	Forced   bool   // Whether this is a forced subtitle track
}
```

### EncryptionConfig (HLS AES-128)

```go
type EncryptionConfig struct {
	KeyURI string // URI placed inside the HLS playlist (default: "enc.key")
	IV     string // Optional 32-character hex Initialization Vector
	Key    []byte // 16-byte raw AES key (if nil, cryptographically generated automatically)
}
```

### S3Config

```go
type S3Config struct {
	Endpoint        string // S3 / MinIO / R2 endpoint URL (e.g. "https://s3.amazonaws.com")
	Bucket          string // Target bucket name
	Region          string // AWS region (default: "us-east-1")
	AccessKey       string // AWS Access Key ID
	SecretKey       string // AWS Secret Access Key
	KeyPrefix       string // Object key prefix (e.g. "streams/video-123")
	ConcurrentFiles int    // Parallel upload workers (default: 5)
	UseSSL          bool   // Use HTTPS connection (default: true)
}
```

---

## Primary Packaging Functions

### EncodeHls

```go
func EncodeHls(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error)
func EncodeHlsWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) (*executor.Usage, error)
```

Encodes the media into an HLS adaptive bitrate stream (`master.m3u8`, variant playlists, `seg_*.m4s` / `seg_*.ts`).

### EncodeDash

```go
func EncodeDash(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error)
func EncodeDashWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) (*executor.Usage, error)
```

Encodes the media into DASH CMAF (`manifest.mpd`, `init-stream*.m4s`, `chunk-stream*.m4s`).

---

## Functional Options Reference

| Functional Option | Description |
|---|---|
| `WithCodec(codec VideoCodec)` | Sets the video codec (`CodecH264`, `CodecHEVC`, `CodecAV1`). |
| `WithHEVC()` | Shortcut for `WithCodec(CodecHEVC)`. |
| `WithAV1()` | Shortcut for `WithCodec(CodecAV1)`. |
| `WithCRF(crf int)` | Enables Capped-CRF content-aware bitrate optimization (e.g. `23` for H.264/HEVC, `28` for AV1). |
| `WithThumbnails(cfg ...ThumbnailConfig)` | Generates storyboard sprite sheet and `thumbnails.vtt` for timeline scrubber previews. |
| `WithSubtitles(tracks ...SubtitleTrack)` | Automatically converts SRT to WebVTT and injects subtitle tracks into HLS master playlists and DASH manifests. |
| `WithNormalizeAudio(enabled ...bool)` | Normalizes audio volume using EBU R128 broadcast standard (`loudnorm=I=-16:TP=-1.5:LRA=11`). |
| `WithWatermark(cfg WatermarkConfig)` | Overlays dynamic logo/watermark with responsive scaling, custom placement, and alpha opacity. |
| `WithAES128Encryption(cfg ...EncryptionConfig)` | Generates 16-byte key and encrypts HLS segments with AES-128 (`#EXT-X-KEY:METHOD=AES-128`). |
| `WithS3Upload(cfg S3Config)` | Uploads packaged stream files directly to S3 / MinIO / R2 using pure Go AWS SigV4 with concurrent worker pool. |
| `WithIFrames(enabled ...bool)` | Generates I-frame-only trick-play playlists (`#EXT-X-I-FRAMES-ONLY`) for fast seeking in HLS. |
| `WithNormalizeOrientation(enabled ...bool)` | Probes orientation metadata, transposes video if rotated, and clears output rotation tags. |
| `WithThreads(n int)` | Sets CPU encoding thread count (`0` = auto-detect). |
| `WithBFrames(n int)` | Sets number of consecutive B-frames (default `0`). |
| `WithScaleBitrateWithFPS(enabled ...bool)` | Proportionally scales bitrate caps upward for high-framerate (>30 FPS) videos. |
| `WithNVENC()` | Uses NVIDIA hardware acceleration (`h264_nvenc`, `hevc_nvenc`, `av1_nvenc`). |
| `WithVAAPI()` | Uses Intel/AMD hardware acceleration (`h264_vaapi`, `hevc_vaapi`, `av1_vaapi`). |
| `WithVideoToolbox()` | Uses Apple VideoToolbox hardware acceleration on macOS (`h264_videotoolbox`, `hevc_videotoolbox`). |
| `WithLogLevel(level string)` | Sets FFmpeg log level (`quiet`, `warning`, `error`, `info`, `debug`). |
| `WithLogger(logger *slog.Logger)` | Injects custom structured logger for Mosaic internal logs. |

---

## Standalone Utilities

### NormalizeVideoOrientation

```go
func NormalizeVideoOrientation(ctx context.Context, inputPath, outputPath string) error
```

Physically transposes rotated video frames and removes rotation metadata.

### Preview Server

```go
import "github.com/farshidrezaei/mosaic/preview"

server := preview.NewServer("./output/hls_stream", 8080)
err := server.Start()
```

Launches the built-in dark-mode web player with Hls.js / Dash.js, live quality switcher, and telemetry.
