# Functional Options Reference

Mosaic uses composable functional options to configure encoding pipelines.

---

## Complete Options List

### Orientation Normalization

```go
func WithNormalizeOrientation(enabled ...bool) Option
```
- **Default**: Disabled (unless explicitly passed).
- **Behavior**: Probes container rotation metadata (`90°`, `180°`, `270°`), applies physical frame transposition, clears output rotation tags, and cleans up temporary files after encoding.

```go
mosaic.WithNormalizeOrientation()      // Enables normalization
mosaic.WithNormalizeOrientation(false) // Explicitly disables
```

---

### CPU Threading

```go
func WithThreads(n int) Option
```
- **Default**: `0` (FFmpeg auto-selection).
- **Behavior**: Sets the `-threads` flag in FFmpeg.

```go
mosaic.WithThreads(8)
```

---

### B-Frames

```go
func WithBFrames(n int) Option
```
- **Default**: `0`.
- **Behavior**: Configures the number of consecutive B-frames for non-baseline profiles (Baseline renditions like 360p always maintain `0` B-frames for H.264 Baseline compliance).

```go
mosaic.WithBFrames(2)
```

---

### High-Framerate Bitrate Scaling

```go
func WithScaleBitrateWithFPS(enabled ...bool) Option
```
- **Default**: Disabled.
- **Behavior**: Scales bitrate caps upward for high-framerate content (>30 FPS, e.g. 50fps, 60fps) by `(FPS / 30.0)` up to a max factor of `1.5x`.

```go
mosaic.WithScaleBitrateWithFPS()
```

---

### Hardware Acceleration (GPU)

```go
func WithNVENC() Option          // NVIDIA NVENC (h264_nvenc)
func WithVAAPI() Option          // Intel / AMD VAAPI (h264_vaapi)
func WithVideoToolbox() Option   // Apple Silicon / macOS (h264_videotoolbox)
func WithGPU(t config.GPUType) Option
```

```go
mosaic.WithNVENC()
```

---

### Logging & Diagnostics

```go
func WithLogLevel(level string) Option
func WithLogger(logger *slog.Logger) Option
```

- **LogLevel values**: `"quiet"`, `"panic"`, `"fatal"`, `"error"`, `"warning"`, `"info"`, `"verbose"`, `"debug"`.
- **Logger**: Sets a standard Go `*slog.Logger` for internal library logs.

```go
mosaic.WithLogLevel("error")
mosaic.WithLogger(slog.Default())
```
