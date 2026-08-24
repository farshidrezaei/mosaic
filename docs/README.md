# Mosaic

<p align="center">
  <img src="assets/logo.jpg" width="180" alt="Mosaic Logo">
</p>

<p align="center">
  <b>Predictable, Production-Ready Adaptive Bitrate (ABR) Video Packaging for Go</b>
</p>

<p align="center">
  <a href="https://farshidrezaei.github.io/mosaic/"><img src="https://img.shields.io/badge/Documentation-GitHub%20Pages-blue?style=for-the-badge&logo=github" alt="Documentation Portal"></a>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/farshidrezaei/mosaic"><img src="https://pkg.go.dev/badge/github.com/farshidrezaei/mosaic.svg" alt="Go Reference"></a>
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go Version"></a>
  <a href="https://github.com/farshidrezaei/mosaic/actions/workflows/go.yml"><img src="https://github.com/farshidrezaei/mosaic/actions/workflows/go.yml/badge.svg" alt="Build Status"></a>
  <a href="https://github.com/farshidrezaei/mosaic/releases"><img src="https://img.shields.io/github/v/release/farshidrezaei/mosaic?include_prereleases" alt="Latest Release"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
</p>

---

`mosaic` is a robust Go library for adaptive bitrate video packaging. It probes input media with FFprobe, computes an aspect-preserving ABR ladder, applies bitrate optimizations, and generates standardized **HLS (fMP4)** and **DASH CMAF** streams using FFmpeg.

📖 **Full Online Documentation & Guides**: [https://farshidrezaei.github.io/mosaic/](https://farshidrezaei.github.io/mosaic/)

Designed for server-side encoding workloads, background workers, and transcoding pipelines where predictability, clean abstractions, and zero external dependencies are critical.

---

## ⚡ Highlights

- **Standard CMAF Output**: Generates HLS (`master.m3u8`, variant playlists, `fMP4` segments) and DASH (`manifest.mpd`, `init.m4s`, `chunk.m4s`) streams.
- **Aspect-Preserving ABR Ladders**: Automatically preserves the source display aspect ratio — landscape, square (1:1), portrait (9:16), or ultra-wide inputs never get distorted or letterboxed with black bars.
- **Orientation Normalization**: Probes display matrices and rotation tags (`90°`, `180°`, `270°`), physically transposes frames when needed, and resets output metadata so mobile videos display correctly everywhere.
- **Real-Time Progress Tracking**: Accurately computes encoding percentage (`0.0%` to `100.0%`), encoded time, current bitrate, and speed.
- **Hardware Acceleration**: Out-of-the-box support for NVIDIA NVENC, Intel/AMD VAAPI, and Apple VideoToolbox.
- **Single-Pass Filter Complex**: Both HLS and DASH use unified `filter_complex` graphs (`split -> scale -> setsar=1`) for optimal 1-pass encoding performance and SAR consistency.
- **High Framerate Bitrate Scaling**: Optional automatic bitrate adjustments for high-framerate content (>30 FPS).
- **Configurable B-Frames**: Tune B-frame counts across profiles for maximum compression efficiency.
- **Zero Third-Party Dependencies**: Built strictly with Go standard library + FFmpeg/FFprobe CLI tooling.
- **Fully Testable Architecture**: Interface-driven command executor allows 100% unit testing without calling live FFmpeg.

---

## 📋 Requirements

- **Go**: `1.25+`
- **FFmpeg**: `4.4+` (with `libx264` and `aac` support)
- **FFprobe**: Typically installed alongside FFmpeg

---

## 📦 Installation

```bash
go get github.com/farshidrezaei/mosaic
```

---

## 🚀 Quick Start

### 1. HLS Packaging

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
		Input:     "input.mp4",
		OutputDir: "./output/hls",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[%5.1f%%] time=%s bitrate=%s speed=%s",
				info.Percentage, info.CurrentTime, info.Bitrate, info.Speed)
		},
	}

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(), // Handles mobile/rotated video
		mosaic.WithThreads(4),
	)
	if err != nil {
		log.Fatalf("Encoding failed: %v", err)
	}

	fmt.Printf("\nDone! CPU User Time: %.2fs | Peak RSS: %d KB\n", usage.UserTime, usage.MaxMemory)
}
```

### 2. DASH CMAF Packaging

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
		Input:     "input.mp4",
		OutputDir: "./output/dash",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[%5.1f%%] time=%s bitrate=%s speed=%s",
				info.Percentage, info.CurrentTime, info.Bitrate, info.Speed)
		},
	}

	_, err := mosaic.EncodeDash(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithBFrames(2),
		mosaic.WithScaleBitrateWithFPS(),
	)
	if err != nil {
		log.Fatalf("DASH encoding failed: %v", err)
	}

	fmt.Println("\nDASH packaging complete -> ./output/dash/manifest.mpd")
}
```

---

## 🧭 Core Workflow

```text
Input Media ──► probe ──► ladder ──► optimize ──► encoder ──► FFmpeg (CMAF)
```

1. **Probe**: `probe.Input` extracts video dimensions, framerate, duration, audio presence, and rotation metadata.
2. **Ladder**: `ladder.Build` constructs a ladder preserving the original display aspect ratio.
3. **Optimize**: `optimize.Apply` caps bitrates based on resolution/FPS and trims redundant, closely-spaced renditions.
4. **Encoder**: Generates an optimal single-pass FFmpeg command graph and streams real-time progress.

---

## 🎛️ Functional Options

Mosaic provides composable functional options to tailor the encoding process:

| Option | Description |
|---|---|
| `mosaic.WithNormalizeOrientation(bool...)` | Probes rotation metadata, transposes video if rotated, and clears output rotation tags. |
| `mosaic.WithThreads(n)` | Sets CPU encoding thread count (`0` = FFmpeg auto-detection). |
| `mosaic.WithBFrames(n)` | Sets number of B-frames for non-baseline profiles (default `0`). |
| `mosaic.WithScaleBitrateWithFPS(bool...)` | Proportionally scales bitrate caps for high-framerate videos (>30 FPS). |
| `mosaic.WithNVENC()` | Uses NVIDIA hardware encoding (`h264_nvenc`). |
| `mosaic.WithVAAPI()` | Uses Intel/AMD hardware encoding (`h264_vaapi`). |
| `mosaic.WithVideoToolbox()` | Uses Apple VideoToolbox hardware encoding (`h264_videotoolbox`). |
| `mosaic.WithGPU(config.GPUType)` | Selects a specific GPU backend explicitly. |
| `mosaic.WithLogLevel(level)` | Sets FFmpeg log level (`quiet`, `error`, `warning`, `info`, `debug`). |
| `mosaic.WithLogger(logger)` | Sets a custom `*slog.Logger` for internal library logs. |

---

## 📐 Aspect Ratio & Ladder Preservation

Unlike legacy pipelines that letterbox non-16:9 videos into fixed frames, Mosaic calculates each rendition's width dynamically based on display dimensions:

| Input Resolution | Aspect Ratio | Generated Renditions |
|---|---|---|
| `1920x1080` | 16:9 Landscape | `1920x1080` (5000k), `1280x720` (3000k), `640x360` (1000k) |
| `1080x1080` | 1:1 Square | `1080x1080` (5000k), `720x720` (3000k), `360x360` (1000k) |
| `1080x1920` | 9:16 Portrait | `608x1080` (5000k), `404x720` (3000k), `202x360` (1000k) |
| `1280x718` | Custom Landscape | `642x360` (1000k) |
| `426x240` | Low Resolution | `426x240` (1000k) *(no upscaling)* |

---

## 📊 Real-Time Progress Monitoring

The `ProgressHandler` receives parsed FFmpeg progress information on every tick:

```go
type ProgressInfo struct {
	Percentage  float64 // Exact percentage (0.0% to 100.0%)
	CurrentTime string  // Encoded timestamp (e.g., "00:01:23.456000")
	Bitrate     string  // Current encoding bitrate (e.g., "2450.3kbits/s")
	Speed       string  // Encoding speed factor (e.g., "1.85x")
}
```

---

## 📂 Examples

Complete, runnable examples are available in the [`examples/`](./examples) directory:

- [`examples/simple_hls`](./examples/simple_hls): Standard HLS VOD packaging with progress reporting.
- [`examples/advanced_dash`](./examples/advanced_dash): DASH CMAF with B-Frames, FPS scaling, and custom thread control.
- [`examples/live_streaming`](./examples/live_streaming): Low-latency live streaming profile (2s segments) for HLS & DASH.
- [`examples/orientation_normalization`](./examples/orientation_normalization): Standalone and pipeline rotation normalization for mobile videos.
- [`examples/progress_monitoring`](./examples/progress_monitoring): Terminal progress bar with percentage, speed, bitrate, and resource usage.
- [`examples/multi_gpu`](./examples/multi_gpu): Multi-backend GPU hardware acceleration (NVENC / VAAPI / VideoToolbox).

---

## 🧪 Testing & Quality Assurance

Mosaic is tested with a 100% dependency-injected architecture, enforcing strict code hygiene and race detection:

```bash
# Run all tests with race detector
GOCACHE=/tmp/go-build go test -v -race ./...

# Static analysis
GOCACHE=/tmp/go-build go vet ./...

# Linter (Mandatory - zero issues policy)
golangci-lint run
```

---

## 📚 Documentation Map

- [🌐 Online Documentation Portal](https://farshidrezaei.github.io/mosaic/): Full interactive guide, API reference, and searchable docs.
- [docs/API.md](./docs/API.md): Complete public API reference and struct definitions.
- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md): Package boundaries and internal execution flow.
- [docs/ENCODING.md](./docs/ENCODING.md): Ladder generation, orientation, HLS, DASH, and FFmpeg filter graphs.
- [docs/TESTING.md](./docs/TESTING.md): Mock executor guide, test strategies, and smoke tests.
- [docs/TROUBLESHOOTING.md](./docs/TROUBLESHOOTING.md): Common errors and debugging tips.
- [STRUCTURE.md](STRUCTURE.md): Repository layout and package responsibilities.
- [CONTRIBUTING.md](CONTRIBUTING.md): Contribution workflow and development contracts.
- [CHANGELOG.md](CHANGELOG.md): Complete version history and release notes.
- [SECURITY.md](SECURITY.md): Vulnerability reporting policy.

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.

