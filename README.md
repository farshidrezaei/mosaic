# Mosaic

[![Go Report Card](https://goreportcard.com/badge/github.com/farshidrezaei/mosaic)](https://goreportcard.com/report/github.com/farshidrezaei/mosaic)
[![Go](https://github.com/farshidrezaei/mosaic/actions/workflows/go.yml/badge.svg)](https://github.com/farshidrezaei/mosaic/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`mosaic` is a Go library for adaptive bitrate (ABR) encoding to HLS and DASH with CMAF segments.
It probes input media, builds an encoding ladder, optimizes renditions, and runs FFmpeg command generation/execution for
you.

## Features

- Automatic ABR ladder generation (`1080p`, `720p`, `360p` profiles by source capability)
- Ladder optimization (bitrate capping and redundant rung trimming)
- HLS CMAF output (`master.m3u8`, variant playlists, fMP4 segments)
- DASH CMAF output (`manifest.mpd`, init/media segments)
- Orientation-aware ladder selection (portrait/rotated input support)
- Audio stream detection and conditional audio mapping
- Progress callbacks from FFmpeg `-progress` output
- Functional options for threads, GPU backend, log level, logger
- Hardware acceleration options: NVENC, VAAPI, VideoToolbox
- Testable architecture via dependency-injected command executor

## Requirements

- Go `1.25+` (module currently declares `go 1.25`)
- FFmpeg `4.4+` with H.264 and AAC support
- FFprobe (usually bundled with FFmpeg)

## Installation

```bash
go get github.com/farshidrezaei/mosaic
```

## Quick Start

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
		OutputDir: "/tmp/hls_output",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("time=%s speed=%s bitrate=%s\n", info.CurrentTime, info.Speed, info.Bitrate)
		},
	}

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithThreads(4),
		mosaic.WithNVENC(),
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

## Orientation Handling

`mosaic` detects video orientation from FFprobe metadata (`side_data rotation` and `tags.rotate`) and uses effective
display dimensions when building the ladder.

- Natural portrait input (for example `720x1280`) produces portrait renditions.
- Rotated portrait metadata (for example `1920x1080` with rotation `90`) is treated as portrait for ladder decisions.

## Encoding Profiles

| Profile       | Segment Duration | Low Latency |
|---------------|-----------------:|------------:|
| `ProfileVOD`  |               5s |          No |
| `ProfileLive` |               2s |         Yes |

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

func WithThreads(n int) Option
func WithGPU(t ...config.GPUType) Option
func WithNVENC() Option
func WithVAAPI() Option
func WithVideoToolbox() Option
func WithLogLevel(level string) Option
func WithLogger(logger *slog.Logger) Option
```

## Testing

```bash
# Unit + package tests
go test ./...

# Coverage (environment-dependent; may require writable GOCACHE)
GOCACHE=/tmp/go-build go test ./... -cover

# Lint
# golangci-lint run
```

## Repository Layout

```text
mosaic/
├── encode.go
├── job.go
├── config/
├── probe/
├── ladder/
├── optimize/
├── encoder/
├── internal/executor/
└── examples/
```

For deeper package mapping, see `STRUCTURE.md`.
For contribution workflow, see `CONTRIBUTING.md`.
For planned work, see `ROADMAP.md`.

## License

MIT. See `LICENSE`.
