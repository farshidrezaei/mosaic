# Functional Options Reference

Mosaic provides composable, type-safe functional options to configure every aspect of the adaptive bitrate packaging pipeline.

---

## 🎨 Feature Options

### Next-Gen Codecs & CRF

```go
func WithCodec(codec VideoCodec) Option
func WithHEVC() Option
func WithAV1() Option
func WithCRF(crf int) Option
```

- **`WithCodec(codec)`**: Sets the target video encoder standard:
  - `mosaic.CodecH264`: H.264 / AVC (`libx264`, `h264_nvenc`, `h264_vaapi`, `h264_videotoolbox`).
  - `mosaic.CodecHEVC`: H.265 / HEVC (`libx265`, `hevc_nvenc`, `hevc_vaapi`, `hevc_videotoolbox`).
  - `mosaic.CodecAV1`: Next-generation AV1 (`libsvtav1`, `av1_nvenc`, `av1_vaapi`).
- **`WithCRF(crf)`**: Enables **Capped-CRF (Content-Aware Bitrate)**. Instead of fixed bitrate ladders, FFmpeg targets a constant visual quality level while still strictly capping peaks at `-maxrate` and `-bufsize` (VBV buffer). Recommended: `23` for H264/HEVC, `28` for AV1.

```go
mosaic.WithAV1(),
mosaic.WithCRF(28),
```

---

### Thumbnails & Storyboards

```go
func WithThumbnails(cfg ...ThumbnailConfig) Option
```

- **Behavior**: Generates a compact JPEG sprite sheet (`thumbnails_0.jpg`) and a standard WebVTT cue file (`thumbnails.vtt`) containing precise `#xywh=x,y,w,h` spatial coordinates for video player timeline scrubber previews.
- **Config**:
  ```go
  mosaic.WithThumbnails(mosaic.ThumbnailConfig{
      IntervalSeconds: 2,   // Capture 1 frame every 2 seconds
      TileWidth:       160, // Width per tile in pixels
      TileHeight:      90,  // Height per tile in pixels
      Columns:         5,   // 5x5 grid per sprite image
      Rows:            5,
      Quality:         3,   // High JPEG quality
  })
  ```

---

### Subtitles & Audio Normalization

```go
func WithSubtitles(tracks ...SubtitleTrack) Option
func WithNormalizeAudio(enabled ...bool) Option
```

- **`WithSubtitles`**: Accepts one or more subtitle tracks (`.srt` or `.vtt`). SRT files are automatically converted to clean WebVTT format, written alongside the media output, and registered into HLS playlists (`#EXT-X-MEDIA:TYPE=SUBTITLES`) and DASH manifests (`<AdaptationSet contentType="text">`).
- **`WithNormalizeAudio`**: Normalizes audio levels across the entire stream using the **EBU R128 (`loudnorm=I=-16:TP=-1.5:LRA=11`)** broadcast standard, preventing sudden volume jumps between videos or commercials.

```go
mosaic.WithNormalizeAudio(),
mosaic.WithSubtitles(
    mosaic.SubtitleTrack{
        Path:     "./subs/en.srt",
        Language: "en",
        Label:    "English",
        Default:  true,
    },
    mosaic.SubtitleTrack{
        Path:     "./subs/fa.srt",
        Language: "fa",
        Label:    "Persian",
    },
),
```

---

### Dynamic Watermarking

```go
func WithWatermark(cfg WatermarkConfig) Option
```

- **Behavior**: Dynamically overlays a logo or watermark image (PNG / WebP) onto every video rendition without distorting aspect ratios or requiring pre-rendered video assets.
- **Positions**: `PositionTopRight`, `PositionTopLeft`, `PositionBottomRight`, `PositionBottomLeft`, `PositionCenter`.
- **Config**:
  ```go
  mosaic.WithWatermark(mosaic.WatermarkConfig{
      Path:     "./assets/logo.png",
      Position: mosaic.PositionTopRight,
      OffsetX:  20,
      OffsetY:  20,
      Opacity:  0.80, // 80% opacity
  })
  ```

---

### HLS AES-128 Encryption

```go
func WithAES128Encryption(cfg ...EncryptionConfig) Option
```

- **Behavior**: Generates a cryptographically secure 16-byte random key (`enc.key`), creates the `enc.keyinfo` file, encrypts all HLS media segments (`.ts` / `.m4s`), and writes `#EXT-X-KEY:METHOD=AES-128,URI="enc.key"` in all variant playlists.
- **Custom Key URI**: You can point `KeyURI` to your authorization server endpoint (e.g. `https://api.example.com/v1/keys/video123`).

```go
mosaic.WithAES128Encryption(mosaic.EncryptionConfig{
    KeyURI: "https://api.yourdomain.com/keys/session.key",
})
```

---

### S3 & Cloud Storage Upload

```go
func WithS3Upload(cfg S3Config) Option
```

- **Behavior**: Uploads all generated stream files (`.m3u8`, `.mpd`, `.m4s`, `.ts`, `.vtt`, `.jpg`, `.key`) directly to AWS S3, MinIO, or Cloudflare R2 using a **zero-dependency pure Go AWS SigV4 signer** and a concurrent worker pool.
- **Headers**: Automatically sets optimal `Content-Type` headers and cache policies (`Cache-Control: public, max-age=31536000, immutable` for media segments, `no-cache` for manifests).

```go
mosaic.WithS3Upload(mosaic.S3Config{
    Endpoint:        "https://s3.us-east-1.amazonaws.com",
    Bucket:          "my-vod-streams",
    Region:          "us-east-1",
    KeyPrefix:       "videos/vid-9901",
    AccessKey:       "AWS_ACCESS_KEY",
    SecretKey:       "AWS_SECRET_KEY",
    ConcurrentFiles: 8,
})
```

---

### Trick-Play (I-Frames Only)

```go
func WithIFrames(enabled ...bool) Option
```

- **Behavior**: Generates `#EXT-X-I-FRAMES-ONLY` playlists in HLS to allow smooth high-speed scrubbing, trick-play, and fast-forward in Apple AVPlayer, Roku, and Smart TVs.

---

### Orientation Normalization

```go
func WithNormalizeOrientation(enabled ...bool) Option
```

- **Behavior**: Probes container rotation metadata (`90°`, `180°`, `270°`), applies physical frame transposition, clears output rotation tags, and cleans up temporary files after encoding.

---

## ⚙️ Performance & Engine Options

### CPU Threading

```go
func WithThreads(n int) Option
```
- **Default**: `0` (FFmpeg auto-selection).
- **Behavior**: Sets FFmpeg `-threads`.

---

### B-Frames

```go
func WithBFrames(n int) Option
```
- **Default**: `0`.
- **Behavior**: Configures the number of consecutive B-frames for non-baseline renditions (Baseline renditions like 360p always maintain `0` B-frames for H.264 Baseline compliance).

---

### High-Framerate Bitrate Scaling

```go
func WithScaleBitrateWithFPS(enabled ...bool) Option
```
- **Default**: Disabled.
- **Behavior**: Scales bitrate caps upward for high-framerate content (>30 FPS, e.g. 50fps, 60fps) by `(FPS / 30.0)` up to a max factor of `1.5x`.

---

### Hardware Acceleration (GPU)

```go
func WithNVENC() Option          // NVIDIA NVENC (h264_nvenc, hevc_nvenc, av1_nvenc)
func WithVAAPI() Option          // Intel / AMD VAAPI (h264_vaapi, hevc_vaapi, av1_vaapi)
func WithVideoToolbox() Option   // Apple Silicon / macOS (h264_videotoolbox, hevc_videotoolbox)
func WithGPU(t config.GPUType) Option
```

---

### Logging & Diagnostics

```go
func WithLogLevel(level string) Option
func WithLogger(logger *slog.Logger) Option
```

- **LogLevel values**: `"quiet"`, `"error"`, `"warning"`, `"info"`, `"debug"`.
- **Logger**: Sets a standard Go `*slog.Logger` for internal library logs.
