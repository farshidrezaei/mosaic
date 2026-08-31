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

`mosaic` is a comprehensive, production-grade Go library for adaptive bitrate (ABR) video packaging. It probes input media, computes an aspect-preserving rendition ladder, applies quality and bitrate optimizations, and generates standardized **HLS (fMP4 / TS)** and **DASH CMAF** streams using FFmpeg.

📖 **Full Online Documentation & Guides**: [https://farshidrezaei.github.io/mosaic/](https://farshidrezaei.github.io/mosaic/)

Designed for server-side video infrastructure, background workers, and media pipelines where predictability, performance, clean abstractions, and zero external dependencies are critical.

---

## ⚡ Highlights

- **Standard CMAF Output**: Generates HLS (`master.m3u8`, variant playlists, `fMP4` segments) and DASH (`manifest.mpd`, `init.m4s`, `chunk.m4s`) streams.
- **Next-Gen Codecs**: First-class support for **AV1** (`libsvtav1`), **HEVC / H.265** (`libx265`), and **H.264 / AVC** (`libx264`) with software and GPU hardware acceleration.
- **Smart Quality & Capped-CRF**: Content-Aware Bitrate optimization combining Constant Rate Factor (CRF) with VBV maxrate caps to save 30–50% bandwidth without quality loss.
- **Storyboard Thumbnails (Trick-Play Preview)**: Automatically generates sprite sheets (`thumbnails_0.jpg`) and standard WebVTT cue files (`thumbnails.vtt`) for timeline scrubber previews in modern video players.
- **Built-in Web Preview DevTools Player**: Instantly test and inspect streams locally via `mosaic preview [dir]` featuring a dark-mode web player (Hls.js / Dash.js), audio/rendition switcher, and live diagnostics.
- **Subtitles & Multi-Track Audio**: Converts SRT to WebVTT automatically and injects subtitle tracks into HLS master playlists (`#EXT-X-MEDIA:TYPE=SUBTITLES`) and DASH AdaptationSets.
- **EBU R128 Audio Normalization**: Automatic broadcast-standard audio normalization (`loudnorm=I=-16:TP=-1.5:LRA=11`) ensuring uniform volume levels.
- **Dynamic Watermarking & Branding**: Configurable overlay placement (`top-right`, `top-left`, `bottom-right`, `bottom-left`, `center`), alpha opacity blending, and auto-scaling relative to rendition widths.
- **HLS AES-128 Segment Encryption**: Automated 16-byte key generation and playlist tagging (`#EXT-X-KEY:METHOD=AES-128`) for content protection.
- **Zero-Dependency Cloud Storage Upload (S3 / MinIO / R2)**: Concurrent streaming asset upload with pure Go AWS SigV4 signing and automatic `Content-Type` / `Cache-Control: immutable` headers.
- **Aspect-Preserving ABR Ladders**: Automatically preserves the source display aspect ratio — landscape, square (1:1), portrait (9:16), or ultra-wide inputs never get distorted or letterboxed with black bars.
- **Orientation Normalization**: Probes display matrices and rotation tags (`90°`, `180°`, `270°`), physically transposes frames when needed, and resets output metadata so mobile videos display correctly everywhere.
- **Real-Time Progress Tracking**: Accurately computes encoding percentage (`0.0%` to `100.0%`), encoded time, current bitrate, and speed.
- **Zero Third-Party Dependencies**: Built strictly with Go standard library + FFmpeg/FFprobe CLI tooling.
- **Fully Testable Architecture**: Interface-driven command executor allows 100% unit testing without calling live FFmpeg.

---

## 📋 Requirements

- **Go**: `1.25+`
- **FFmpeg**: `4.4+` (with `libx264`, `libx265`, `libsvtav1`, and `aac` support)
- **FFprobe**: Typically installed alongside FFmpeg

---

## 📦 Installation

### As a Go Library
```bash
go get github.com/farshidrezaei/mosaic
```

### As a Standalone CLI Tool
```bash
# Install directly via Go
go install github.com/farshidrezaei/mosaic/cmd/mosaic@latest

# Or run via Docker (FFmpeg pre-installed)
docker run --rm -v $(pwd):/workspace ghcr.io/farshidrezaei/mosaic -i input.mp4 -o ./output/hls --thumbnails
```

---

## 🛠️ CLI Quick Start

```bash
# 1. Package video into HLS with mobile normalization and thumbnail scrubber:
mosaic -i input.mp4 -o ./output/hls --thumbnails

# 2. Package with Next-Gen AV1 codec, Capped-CRF, Watermark, and Audio Normalization:
mosaic -i input.mp4 -o ./output/hls_av1 \
  --codec av1 \
  --crf 28 \
  --watermark ./logo.png \
  --normalize-audio \
  --thumbnails

# 3. Encrypt HLS segments with AES-128:
mosaic -i input.mp4 -o ./output/hls_secure --encrypt-aes128

# 4. Package and auto-upload directly to S3 / MinIO:
mosaic -i input.mp4 -o ./output/hls_s3 \
  --s3-bucket my-stream-bucket \
  --s3-prefix videos/movie1 \
  --s3-region us-east-1

# 5. Launch local web player to preview generated streams:
mosaic preview ./output/hls
```

---

## 🚀 Go Library Usage

### 1. Complete Production HLS Workflow (Thumbnails, Watermark, Subtitles, AES-128)

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
		Input:     "movie.mp4",
		OutputDir: "./output/hls_stream",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[%5.1f%%] time=%s bitrate=%s speed=%s",
				info.Percentage, info.CurrentTime, info.Bitrate, info.Speed)
		},
	}

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(), // Correct mobile 90°/270° orientation
		mosaic.WithNormalizeAudio(),       // EBU R128 broadcast audio leveling
		mosaic.WithThumbnails(),           // Generates thumbnails.vtt & sprite sheet
		mosaic.WithWatermark(mosaic.WatermarkConfig{
			Path:     "./branding/logo.png",
			Position: mosaic.PositionTopRight,
			Opacity:  0.85,
		}),
		mosaic.WithSubtitles(mosaic.SubtitleTrack{
			Path:     "./subtitles/en.srt", // Auto-converted to WebVTT
			Language: "en",
			Label:    "English",
			Default:  true,
		}),
		mosaic.WithAES128Encryption(),     // Generates enc.key and encrypts segments
		mosaic.WithThreads(4),
	)
	if err != nil {
		log.Fatalf("Encoding failed: %v", err)
	}

	fmt.Printf("\nDone! CPU User Time: %.2fs | Peak RSS: %d KB\n", usage.UserTime, usage.MaxMemory)
}
```

### 2. Next-Gen AV1 & HEVC Encoding with Capped-CRF

```go
package main

import (
	"context"
	"log"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	job := mosaic.Job{
		Input:     "input.mp4",
		OutputDir: "./output/av1_hls",
		Profile:   mosaic.ProfileVOD,
	}

	// High-efficiency AV1 ABR ladder with Capped-CRF 28
	_, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithAV1(),
		mosaic.WithCRF(28),
		mosaic.WithThumbnails(),
	)
	if err != nil {
		log.Fatalf("AV1 encoding failed: %v", err)
	}
}
```

### 3. Direct Cloud Upload to S3 / MinIO / Cloudflare R2

```go
package main

import (
	"context"
	"log"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	job := mosaic.Job{
		Input:     "input.mp4",
		OutputDir: "./output/hls",
		Profile:   mosaic.ProfileVOD,
	}

	_, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithThumbnails(),
		mosaic.WithS3Upload(mosaic.S3Config{
			Endpoint:  "https://s3.us-east-1.amazonaws.com",
			Bucket:    "my-media-bucket",
			Region:    "us-east-1",
			KeyPrefix: "content/video-101",
			AccessKey: "YOUR_ACCESS_KEY",
			SecretKey: "YOUR_SECRET_KEY",
		}),
	)
	if err != nil {
		log.Fatalf("Packaging and upload failed: %v", err)
	}
}
```

---

## 📚 Documentation

For complete architecture details, API references, tutorials, and benchmark results:

- [Documentation Portal (GitHub Pages)](https://farshidrezaei.github.io/mosaic/)
- [System Architecture](docs/ARCHITECTURE.md)
- [Encoding & Filter Graph](docs/ENCODING.md)
- [Public API Reference](docs/API.md)
- [Functional Options Reference](docs/options.md)
- [Examples Catalog](docs/EXAMPLES.md)
- [Testing & Quality Guide](docs/TESTING.md)
- [Troubleshooting & FAQ](docs/TROUBLESHOOTING.md)

---

## 🤝 Contributing

Contributions are welcome! Please review [CONTRIBUTING.md](CONTRIBUTING.md) and [AGENTS.md](AGENTS.md) before submitting pull requests.

```bash
# Verify code formatting, tests, and linting
gofmt -w .
GOCACHE=/tmp/go-build go test -v -race ./...
golangci-lint run
```

---

## 📄 License

Mosaic is licensed under the [MIT License](LICENSE).
