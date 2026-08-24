# Mosaic
<p align="center">
  <img src="assets/logo.jpg" width="180">
</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/farshidrezaei/mosaic)](https://goreportcard.com/report/github.com/farshidrezaei/mosaic)
[![Go](https://github.com/farshidrezaei/mosaic/actions/workflows/go.yml/badge.svg)](https://github.com/farshidrezaei/mosaic/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`mosaic` is a Go library for adaptive bitrate video packaging. It probes input media with FFprobe, builds an optimized ABR ladder, and generates HLS or DASH CMAF outputs with FFmpeg.

The library is designed for applications that need predictable server-side encoding without shelling out manually for every FFmpeg command.

## Highlights

- HLS CMAF output with `master.m3u8`, variant playlists, init segments, and fMP4 media segments.
- DASH CMAF output with `manifest.mpd`, init segments, and media segments.
- Aspect-preserving ABR ladders. Square inputs stay square, portrait inputs stay portrait, and non-standard dimensions keep their display ratio.
- Orientation-aware probing using FFprobe display matrix and rotate metadata.
- Optional orientation normalization that physically rotates frames and clears rotation metadata.
- Automatic output directory creation for HLS and DASH jobs.
- Audio stream detection with conditional audio mapping.
- Progress callbacks from FFmpeg `-progress` output.
- Functional options for threads, hardware encoder backend, FFmpeg log level, and logger.
- Hardware encoder selection for NVENC, VAAPI, and VideoToolbox.
- Dependency-injected command execution for fast unit tests and controlled integrations.

## Requirements

- Go `1.25+`
- FFmpeg `4.4+`
- FFprobe, usually installed with FFmpeg
- FFmpeg builds must include H.264 and AAC support for the default outputs

Hardware acceleration requires the corresponding FFmpeg encoder to be present:

- `h264_nvenc` for NVIDIA NVENC
- `h264_vaapi` for VAAPI
- `h264_videotoolbox` for Apple VideoToolbox

## Installation

```bash
go get github.com/farshidrezaei/mosaic
```

## Quick Start: HLS

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	job := mosaic.Job{
		Input:     "/path/to/input.mp4",
		OutputDir: "/tmp/mosaic-hls",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("time=%s bitrate=%s speed=%s\n", info.CurrentTime, info.Bitrate, info.Speed)
		},
	}

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithThreads(4),
		mosaic.WithLogLevel("warning"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if usage != nil {
		fmt.Printf("user=%.2fs system=%.2fs maxrss=%d\n", usage.UserTime, usage.SystemTime, usage.MaxMemory)
	}
}
```

The generated HLS directory contains:

```text
master.m3u8
stream_0.m3u8
stream_1.m3u8
stream_2.m3u8
init_0.mp4 or init.mp4
seg_0_0.m4s
...
```

The exact number of variants depends on the input height and optimizer decisions.

## Quick Start: DASH

```go
package main

import (
	"context"
	"log"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	job := mosaic.Job{
		Input:     "/path/to/input.mp4",
		OutputDir: "/tmp/mosaic-dash",
		Profile:   mosaic.ProfileVOD,
	}

	_, err := mosaic.EncodeDash(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithLogLevel("warning"),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

The generated DASH directory contains:

```text
manifest.mpd
init-stream$RepresentationID$.m4s
chunk-stream$RepresentationID$-$Number$.m4s
...
```

## Input Sources

`Job.Input` can be a local file path or a public URL accepted by FFmpeg and FFprobe.

```go
job := mosaic.Job{
	Input:     "https://example.com/video.mp4",
	OutputDir: "/tmp/output",
	Profile:   mosaic.ProfileVOD,
}
```

For remote inputs, FFmpeg and FFprobe must be able to resolve and fetch the URL from the runtime environment.

## Output Directories

`mosaic` creates `Job.OutputDir` automatically before invoking FFmpeg.

If the process cannot write to the target directory, encoding fails with an error such as:

```text
create output dir: mkdir /path: permission denied
```

Use a writable output directory and ensure enough disk space for all generated variants and segments.

## Encoding Profiles

| Profile       | Segment Duration | Low Latency Flags |
|---------------|-----------------:|------------------:|
| `ProfileVOD`  |               5s |                No |
| `ProfileLive` |               2s |               Yes |

`ProfileLive` enables shorter segments and HLS low-latency related flags. It does not implement a full live ingest loop by itself.

## Aspect Ratio and Ladder Behavior

Mosaic builds quality rungs using target heights of `1080`, `720`, and `360`, but it computes each rung width from the input display aspect ratio.

Examples:

| Input Display Size | Output Rungs                       |
|--------------------|------------------------------------|
| `1920x1080`        | `1920x1080`, `1280x720`, `640x360` |
| `1080x1080`        | `1080x1080`, `720x720`, `360x360`  |
| `1080x1920`        | `608x1080`, `404x720`, `202x360`   |
| `1280x718`         | `642x360`                          |
| `426x240`          | `426x240`                          |

Notes:

- The ladder is based on effective display dimensions after rotation metadata is applied.
- Output dimensions are forced to even numbers for broad H.264 and `yuv420p` compatibility.
- Inputs below 360p are not upscaled. The fallback rung keeps the source display height.
- HLS no longer pads video into a fixed 16:9 frame. The output frame itself carries the preserved aspect ratio.

## Orientation Handling

Some phone videos store frames in one orientation and rely on container metadata such as a display matrix or `rotate` tag to tell players how to display them.

Mosaic handles this in two layers:

1. `probe.VideoInfo` exposes `DisplayWidth`, `DisplayHeight`, and `IsPortrait`, so ladders are built from display dimensions instead of raw stored dimensions.
2. `WithNormalizeOrientation()` physically rotates frames for `90`, `180`, and `270` degree metadata and clears rotation metadata from output streams.

Recommended default:

```go
usage, err := mosaic.EncodeHls(
	ctx,
	job,
	mosaic.WithNormalizeOrientation(),
)
```

Use orientation normalization when targeting browsers, mobile players, or mixed playback environments where rotation metadata support may differ.

## Progress Reporting

Set `Job.ProgressHandler` to receive parsed values from FFmpeg `-progress`.

```go
job.ProgressHandler = func(info mosaic.ProgressInfo) {
	fmt.Printf("time=%s bitrate=%s speed=%s\n", info.CurrentTime, info.Bitrate, info.Speed)
}
```

Fields:

- `CurrentTime`: FFmpeg `out_time`
- `Bitrate`: FFmpeg `bitrate`
- `Speed`: FFmpeg `speed`
- `Percentage`: currently reserved and reported as `0`

## Hardware Encoding

Use a convenience option:

```go
mosaic.WithNVENC()
mosaic.WithVAAPI()
mosaic.WithVideoToolbox()
```

Or select a backend explicitly:

```go
mosaic.WithGPU(config.GPU_NVENC)
```

Hardware options only select FFmpeg encoder names. They do not validate driver availability, device permissions, pixel format constraints, or platform-specific FFmpeg requirements.

## Public API

```go
type Job struct {
	Input           string
	OutputDir       string
	ProgressHandler ProgressHandler
	Profile         Profile
}

type ProgressInfo struct {
	CurrentTime string
	Bitrate     string
	Speed       string
	Percentage  float64
}

type Profile string

const (
	ProfileVOD  Profile = "vod"
	ProfileLive Profile = "live"
)

func EncodeHls(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error)
func EncodeHlsWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) (*executor.Usage, error)

func EncodeDash(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error)
func EncodeDashWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) (*executor.Usage, error)

func NormalizeVideoOrientation(ctx context.Context, inputPath, outputPath string) error

func WithThreads(n int) Option
func WithGPU(t ...config.GPUType) Option
func WithNormalizeOrientation(enabled ...bool) Option
func WithNVENC() Option
func WithVAAPI() Option
func WithVideoToolbox() Option
func WithLogLevel(level string) Option
func WithLogger(logger *slog.Logger) Option
```

See [docs/API.md](./docs/API.md) for detailed API notes.

## Examples

Runnable examples live under `examples/`.

```bash
cd examples/simple_hls
cp /path/to/input.mp4 input.mp4
GOCACHE=/tmp/go-build go run .
```

Available examples:

- `examples/simple_hls`: basic HLS encode
- `examples/advanced_dash`: DASH encode with options
- `examples/multi_gpu`: tries several hardware encoder backends

## Testing

```bash
# Unit and package tests
go test ./...

# If the default Go cache is not writable in your environment
GOCACHE=/tmp/go-build go test ./...

# Vet
GOCACHE=/tmp/go-build go vet ./...

# Coverage
GOCACHE=/tmp/go-build go test ./... -cover
```

Optional lint:

```bash
golangci-lint run
```

See [docs/TESTING.md](./docs/TESTING.md) for integration and smoke-test guidance.

## Documentation Map

- [docs/API.md](./docs/API.md): public API reference and usage notes
- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md): package boundaries and runtime flow
- [docs/ENCODING.md](./docs/ENCODING.md): ladder, orientation, HLS, DASH, and FFmpeg behavior
- [docs/TESTING.md](./docs/TESTING.md): test strategy, commands, and smoke tests
- [docs/TROUBLESHOOTING.md](./docs/TROUBLESHOOTING.md): common failures and debugging steps
- [STRUCTURE.md](STRUCTURE.md): quick repository layout
- [CONTRIBUTING.md](CONTRIBUTING.md): contribution workflow
- [ROADMAP.md](ROADMAP.md): planned work
- [CHANGELOG.md](CHANGELOG.md): release notes
- [SUPPORT.md](SUPPORT.md): support and bug-report guidance
- [SECURITY.md](SECURITY.md): security reporting policy
- [AGENTS.md](AGENTS.md): instructions for AI coding agents

## Repository Layout

```text
mosaic/
├── encode.go
├── job.go
├── orientation.go
├── config/
├── probe/
├── ladder/
├── optimize/
├── encoder/
├── internal/executor/
├── examples/
├── docs/
└── README.md
```

## Current Limitations

- H.264/AAC are the default codecs.
- HEVC and AV1 are planned but not implemented as public options yet.
- `ProgressInfo.Percentage` is reserved and not computed yet.
- Cloud uploads, DRM packaging, thumbnails, and sprites are planned but not implemented.
- Hardware encoder options assume FFmpeg and the host system are already configured.

## License

MIT. See [LICENSE](LICENSE).
