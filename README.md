# 🎬 Mosaic

[![Go Report Card](https://goreportcard.com/badge/github.com/farshidrezaei/mosaic)](https://goreportcard.com/report/github.com/farshidrezaei/mosaic)
[![Go](https://github.com/farshidrezaei/mosaic/actions/workflows/go.yml/badge.svg)](https://github.com/farshidrezaei/mosaic/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Mosaic** is a Go library for adaptive bitrate (ABR) video encoding that generates HLS and DASH streams with CMAF
segments. It automatically builds an optimized encoding ladder based on your source video and handles all the complexity
of FFmpeg command construction.

## ✨ Features

- 🎯 **Automatic Ladder Building** - Generates optimal renditions (1080p, 720p, 360p) based on source resolution
- 🔧 **Intelligent Optimization** - Bitrate capping and redundant rendition trimming
- 📦 **CMAF Support** - Industry-standard fMP4 segments compatible with both HLS and DASH
- ⚡ **Dual Profiles** - VOD (5s segments) and Live (2s low-latency) modes
- 🎨 **Smart Scaling** - Maintains aspect ratio with letterboxing when needed
- 🔊 **Audio Detection** - Automatically handles videos with or without audio
- 📊 **Progress Reporting** - Real-time updates on encoding status
- ⚙️ **Functional Options** - Flexible configuration for threads, GPU, and logging
- 🚀 **Hardware Acceleration** - Support for NVIDIA NVENC, Intel/AMD VAAPI, and Apple VideoToolbox
- 🛡️ **100% Test Coverage** - Comprehensive test suite with mocked dependencies

## 📋 Requirements

- **Go** 1.20 or later
- **FFmpeg** 4.4+ with libx264 and AAC support
- **FFprobe** (comes with FFmpeg)

## 📦 Installation

```bash
go get github.com/farshidrezaei/mosaic
```

## 🚀 Quick Start

> 💡 **See full examples in the [`examples/`](./examples) directory.**

### HLS Encoding

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
		OutputDir: "/output/hls",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("Progress: %s, Speed: %s\n", info.CurrentTime, info.Speed)
		},
	}

	// Use functional options for more control
	err := mosaic.EncodeHls(context.Background(), job, 
		mosaic.WithThreads(4),
		mosaic.WithGPU(),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

**Output:**

```
/output/hls/
├── master.m3u8           # Master playlist
├── stream_0.m3u8         # 1080p variant playlist
├── stream_1.m3u8         # 720p variant playlist
├── stream_2.m3u8         # 360p variant playlist
├── seg_0_0.m4s           # 1080p segments
├── seg_1_0.m4s           # 720p segments
└── seg_2_0.m4s           # 360p segments
```

### DASH Encoding

```go
job := mosaic.Job{
Input:     "/path/to/input.mp4", // or a url
OutputDir: "/output/dash",
Profile:   mosaic.ProfileLive,
}

if err := mosaic.EncodeDash(context.Background(), job); err != nil {
log.Fatal(err)
}
```

**Output:**

```
/output/dash/
├── manifest.mpd               # DASH manifest
├── init-stream0.m4s           # Initialization segments
├── chunk-stream0-00001.m4s    # Media chunks
└── ...
```

## 🎚️ Encoding Profiles

| Profile       | Segment Duration | Use Case          | Latency  |
|---------------|------------------|-------------------|----------|
| `ProfileVOD`  | 5 seconds        | On-demand content | Standard |
| `ProfileLive` | 2 seconds        | Live streaming    | Low      |

## 📐 Automatic Ladder

Mosaic intelligently builds an encoding ladder based on your source video:

| Source Height | Generated Renditions                      |
|---------------|-------------------------------------------|
| ≥ 1080p       | 1080p (5200k), 720p (3000k), 360p (1000k) |
| ≥ 720p        | 720p (3000k), 360p (1000k)                |
| ≥ 360p        | 360p (1000k)                              |

### Optimization Features

1. **Bitrate Capping** - Prevents excessive bandwidth:
    - 1080p max: 5000 kbps
    - 720p max: 3000 kbps
    - Others max: 1000 kbps

2. **Rendition Trimming** - Removes redundant rungs when height ratio < 0.7
    - Example: 1080p + 540p → 720p skipped

## 🏗️ Architecture

```
Input Video
    │
    ├─► Probe (FFprobe)
    │   └─► Resolution, FPS, Audio Detection
    │
    ├─► Ladder Builder
    │   └─► Generate Renditions (1080p, 720p, 360p)
    │
    ├─► Optimizer
    │   └─► Cap Bitrates & Trim Redundant Renditions
    │
    └─► Encoder (FFmpeg)
        ├─► HLS CMAF (fMP4 segments + master.m3u8)
        └─► DASH CMAF (fMP4 segments + manifest.mpd)
```

## 🔧 Under the Hood

### Video Encoding Settings

- **Codec**: H.264/AVC (libx264)
- **Pixel Format**: YUV 4:2:0
- **Preset**: Medium (balanced speed/quality)
- **Rate Control**: VBR with maxrate and bufsize
- **GOP Structure**: Aligned to segment boundaries (FPS × segment_duration)
- **Profiles**: Baseline (360p), Main (720p+)

### Audio Encoding Settings

- **Codec**: AAC
- **Bitrate**: 96 kbps
- **Channels**: Stereo (2.0)
- **Sample Rate**: Automatic

### FFmpeg Optimizations

```bash
-analyzeduration 100M    // Handle complex inputs
-probesize 100M          // Thorough stream analysis
-fflags +genpts          // Fix timestamp issues
-sc_threshold 0          // Disable scene detection (consistent GOPs)
```

## 🧪 Testing Architecture

Mosaic is built with testability in mind, achieving **100% code coverage** on all production logic.

### Dependency Injection
The library uses a `CommandExecutor` interface to abstract FFmpeg and FFprobe interactions. This allows for:
- **Mocked Testing**: Run tests without installing FFmpeg
- **Deterministic Results**: Simulate exact command outputs and errors
- **Safety**: No accidental external command execution during tests

### Running Tests
```bash
# Run all tests
go test ./...

# Check coverage
go test ./... -cover

# Run linter (if installed)
golangci-lint run
```

## 🧩 Package Structure

```
mosaic/
├── .golangci.yml            # Linter configuration
├── job.go                   # Job definition and profiles
├── encode.go                # Main encoding functions
├── probe/                   # Input video analysis
├── ladder/                  # Rendition ladder building
├── optimize/                # Bitrate optimization
├── config/                  # Encoding profiles (VOD, LIVE)
├── internal/                # Internal utilities (executor, mocks)
└── encoder/
    ├── common.go            # Shared utilities
    ├── hls_cmaf.go          # HLS encoder
    └── dash_cmaf.go         # DASH encoder
```

### Hardware Acceleration

Mosaic supports multiple hardware acceleration backends:

```go
// NVIDIA NVENC (Default)
mosaic.EncodeHls(ctx, job, mosaic.WithNVENC())

// Intel/AMD VAAPI
mosaic.EncodeHls(ctx, job, mosaic.WithVAAPI())

// Apple VideoToolbox
mosaic.EncodeHls(ctx, job, mosaic.WithVideoToolbox())

// Generic GPU option (defaults to NVENC)
mosaic.EncodeHls(ctx, job, mosaic.WithGPU())
```

## 🎯 API Reference

### Types

```go
type Job struct {
	Input           string          // Path to input video
	OutputDir       string          // Output directory for segments/manifests
	Profile         Profile         // ProfileVOD or ProfileLive
	ProgressHandler ProgressHandler // Optional progress callback
}

type ProgressInfo struct {
	Percentage  float64
	CurrentTime string
	Bitrate     string
	Speed       string
}

type ProgressHandler func(ProgressInfo)

type Profile string
const (
ProfileVOD  Profile = "vod"  // 5-second segments
ProfileLive Profile = "live" // 2-second segments
)
```

### Functions

```go
// Encode to HLS with CMAF segments
func EncodeHls(ctx context.Context, job Job, opts ...Option) error

// Encode to HLS with a custom command executor
func EncodeHlsWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) error

// Encode to DASH with CMAF segments
func EncodeDash(ctx context.Context, job Job, opts ...Option) error

// Encode to DASH with a custom command executor
func EncodeDashWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) error

// Functional Options
func WithThreads(n int) Option
func WithGPU() Option
func WithLogLevel(level string) Option
func WithLogger(logger *slog.Logger) Option
```

## 🤝 Contributing

Contributions are welcome! Areas for improvement:

- [ ] Parallel rendition encoding
- [ ] HDR/10-bit support
- [ ] Hardware acceleration (VAAPI, NVENC, QSV)
- [ ] Custom audio configurations
- [ ] AV1/VP9 codec support

## 📄 License

MIT License - see [LICENSE](./LICENSE) file for details

## 🙏 Acknowledgments

Built with [FFmpeg](https://ffmpeg.org/) - the Swiss Army knife of video processing.

---

**Made with ❤️ by [Farshid Rezaei](https://github.com/farshidrezaei)**
