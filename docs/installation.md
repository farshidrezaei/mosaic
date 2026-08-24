# Installation & Prerequisites

## Requirements

| Requirement | Minimum Version | Notes |
|---|---|---|
| **Go** | `1.25+` | Required for building applications using Mosaic |
| **FFmpeg** | `4.4+` | Must include `libx264` and `aac` encoders |
| **FFprobe** | `4.4+` | Used for media stream and metadata probing |

---

## Installing Go Module

Add `mosaic` to your Go module dependencies:

```bash
go get github.com/farshidrezaei/mosaic
```

---

## Installing FFmpeg & FFprobe

### Ubuntu / Debian

```bash
sudo apt update
sudo apt install -y ffmpeg
```

### macOS (Homebrew)

```bash
brew install ffmpeg
```

### Arch Linux

```bash
sudo pacman -S ffmpeg
```

### Docker

You can use the official `golang` image and install `ffmpeg`:

```dockerfile
FROM golang:1.25-alpine

RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server .

CMD ["./server"]
```

---

## Hardware Acceleration Setup

If you plan to use GPU encoders, ensure the respective drivers and FFmpeg encoder modules are present:

- **NVIDIA GPU**: `h264_nvenc` (requires NVIDIA CUDA drivers)
- **Intel / AMD GPU**: `h264_vaapi` (requires `libva` and DRM access)
- **Apple Silicon / macOS**: `h264_videotoolbox` (native on macOS)
