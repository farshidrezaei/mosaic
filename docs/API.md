# Mosaic API Reference

This document explains the public API surface of `github.com/farshidrezaei/mosaic`.

## Package

```go
import "github.com/farshidrezaei/mosaic"
```

## Job

```go
type Job struct {
	Input           string
	OutputDir       string
	ProgressHandler ProgressHandler
	Profile         Profile
}
```

### Input

`Input` is the source media path or URL.

Supported examples:

```text
/var/media/input.mp4
./input.mp4
https://example.com/input.mp4
```

The value is passed to FFprobe and FFmpeg. Any protocol support depends on the installed FFmpeg build and runtime network permissions.

### OutputDir

`OutputDir` is where playlists, manifests, init segments, and media segments are written.

Mosaic creates this directory automatically with `0755` permissions. Existing directories are reused.

### Profile

```go
const (
	ProfileVOD  Profile = "vod"
	ProfileLive Profile = "live"
)
```

`ProfileVOD` uses 5-second segments.

`ProfileLive` uses 2-second segments and HLS low-latency related flags. Mosaic does not implement a long-running live ingest loop.

### ProgressHandler

`ProgressHandler` is optional.

```go
type ProgressHandler func(ProgressInfo)
```

It receives parsed FFmpeg `-progress` output.

```go
type ProgressInfo struct {
	CurrentTime string
	Bitrate     string
	Speed       string
	Percentage  float64
}
```

`Percentage` is calculated automatically based on the probed media duration and encoded progress (0.0 to 100.0).

## HLS Encoding

```go
func EncodeHls(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error)
```

`EncodeHls` uses the default command executor.

Generated output includes:

```text
master.m3u8
stream_%v.m3u8
init*.mp4
seg_%v_%d.m4s
```

The exact init segment naming is controlled by FFmpeg's HLS muxer.

## HLS Encoding With Executor

```go
func EncodeHlsWithExecutor(
	ctx context.Context,
	job Job,
	exec executor.CommandExecutor,
	opts ...Option,
) (*executor.Usage, error)
```

Use this variant for tests or advanced command execution control.

The executor is used for all FFprobe and FFmpeg invocations inside the operation.

## DASH Encoding

```go
func EncodeDash(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error)
```

Generated output includes:

```text
manifest.mpd
init-stream$RepresentationID$.m4s
chunk-stream$RepresentationID$-$Number$.m4s
```

## DASH Encoding With Executor

```go
func EncodeDashWithExecutor(
	ctx context.Context,
	job Job,
	exec executor.CommandExecutor,
	opts ...Option,
) (*executor.Usage, error)
```

Use this variant for tests or advanced command execution control.

## Orientation Normalization

```go
func NormalizeVideoOrientation(ctx context.Context, inputPath, outputPath string) error
```

This standalone helper:

1. Probes source orientation metadata.
2. Physically rotates frames when metadata indicates `90`, `180`, or `270` degrees.
3. Clears rotation metadata on output.
4. Verifies the resulting output has no remaining rotation metadata.

For encoding jobs, prefer:

```go
mosaic.WithNormalizeOrientation()
```

## Options

### WithNormalizeOrientation

```go
func WithNormalizeOrientation(enabled ...bool) Option
```

Called without arguments, it enables normalization.

```go
mosaic.WithNormalizeOrientation()
```

Explicit disable:

```go
mosaic.WithNormalizeOrientation(false)
```

### WithThreads

```go
func WithThreads(n int) Option
```

Sets FFmpeg `-threads`.

`0` means FFmpeg auto-selects thread behavior.

### WithBFrames

```go
func WithBFrames(n int) Option
```

Sets the number of B-frames for non-baseline renditions (default is `0`).

### WithScaleBitrateWithFPS

```go
func WithScaleBitrateWithFPS(enabled ...bool) Option
```

Enables optional bitrate scaling for high-framerate inputs (>30 FPS).

### WithLogLevel

```go
func WithLogLevel(level string) Option
```

Sets FFmpeg `-loglevel`.

Common values:

```text
quiet
error
warning
info
debug
```

### WithLogger

```go
func WithLogger(logger *slog.Logger) Option
```

Sets the logger used for Mosaic internal logging.

### Hardware Options

```go
func WithGPU(t ...config.GPUType) Option
func WithNVENC() Option
func WithVAAPI() Option
func WithVideoToolbox() Option
```

Supported GPU constants:

```go
config.GPU_NVENC
config.GPU_VAAPI
config.GPU_VIDEOTOOLBOX
```

These options select FFmpeg encoder names only. They do not configure drivers, devices, upload filters, or platform-specific FFmpeg requirements.

## Usage Return Value

Encoding functions return `*executor.Usage`.

```go
type Usage struct {
	UserTime   float64
	SystemTime float64
	MaxMemory  int64
}
```

The values come from the OS process state after FFmpeg exits.

## Error Handling

Errors may come from:

- FFprobe not installed
- FFmpeg not installed
- invalid input path or URL
- unsupported codec or encoder
- missing hardware encoder
- output directory permission failure
- FFmpeg process failure
- context cancellation

Always check returned errors.

```go
usage, err := mosaic.EncodeHls(ctx, job)
if err != nil {
	return err
}
_ = usage
```

## Context Cancellation

All external commands are created with `exec.CommandContext`.

Canceling the context should terminate the running FFprobe or FFmpeg command.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

_, err := mosaic.EncodeHls(ctx, job)
```

## Testing Hooks

Use `EncodeHlsWithExecutor` and `EncodeDashWithExecutor` with `internal/executor.MockCommandExecutor` in package tests.

For external users, implement the same interface:

```go
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) ([]byte, *Usage, error)
	ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, *Usage, error)
}
```
