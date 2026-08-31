# Examples Catalog

Explore all runnable examples located in the [`examples/`](https://github.com/farshidrezaei/mosaic/tree/main/examples) directory.

---

## 1. Simple HLS Packaging
**Path**: `examples/simple_hls/main.go`

Demonstrates straightforward HLS VOD packaging with real-time percentage progress tracking and automatic orientation normalization.

### Run with Go:
```bash
cd examples/simple_hls && go run main.go
```

### CLI Command Equivalent:
```bash
mosaic -i input.mp4 -o ./output/hls --normalize
```

---

## 2. Advanced DASH CMAF
**Path**: `examples/advanced_dash/main.go`

Demonstrates DASH CMAF packaging with custom B-Frames (`WithBFrames(2)`), high-framerate bitrate scaling (`WithScaleBitrateWithFPS()`), and custom thread count.

### Run with Go:
```bash
cd examples/advanced_dash && go run main.go
```

### CLI Command Equivalent:
```bash
mosaic -i input.mp4 -o ./output/dash --format=dash --bframes=2 --threads=4
```

---

## 3. Storyboard Thumbnails & Web Preview Player
**Path**: `examples/thumbnails_and_preview/main.go`

Demonstrates automatic generation of timeline scrubber sprite sheets (`thumbnails_0.jpg`), WebVTT cues (`thumbnails.vtt`), and starting the built-in dark-mode preview web player.

### Run with Go:
```bash
cd examples/thumbnails_and_preview && go run main.go
```

### CLI Command Equivalent:
```bash
# 1. Package with timeline thumbnails
mosaic -i input.mp4 -o ./output/hls --thumbnails

# 2. Launch instant web preview player
mosaic preview ./output/hls
```

---

## 4. Dynamic Watermarking & Subtitles
**Path**: `examples/watermark_and_subtitles/main.go`

Demonstrates responsive logo overlay with alpha opacity blending and automatic SRT-to-WebVTT conversion injected into HLS and DASH streams.

### Run with Go:
```bash
cd examples/watermark_and_subtitles && go run main.go
```

### CLI Command Equivalent:
```bash
mosaic -i input.mp4 -o ./output/hls_branded \
  --watermark ./branding/logo.png \
  --subtitles ./subtitles/fa.srt,./subtitles/en.srt \
  --normalize-audio
```

---

## 5. HLS AES-128 Content Encryption
**Path**: `examples/encryption_aes128/main.go`

Demonstrates generating cryptographic keys (`enc.key`), writing keyinfo metadata, encrypting HLS segments, and verifying `#EXT-X-KEY` playlist security.

### Run with Go:
```bash
cd examples/encryption_aes128 && go run main.go
```

### CLI Command Equivalent:
```bash
# Package with AES-128 cryptographic segment encryption
mosaic -i input.mp4 -o ./output/hls_encrypted --encrypt-aes128

# Play encrypted stream with local web preview
mosaic preview ./output/hls_encrypted
```

---

## 6. Next-Gen AV1 & HEVC Encoding (Capped-CRF)
**Path**: `examples/nextgen_av1_hevc/main.go`

Demonstrates high-efficiency video coding using **AV1** (`libsvtav1`) and **HEVC** (`libx265`) with Capped-CRF content-aware bitrate control for 30–50% bandwidth reduction.

### Run with Go:
```bash
cd examples/nextgen_av1_hevc && go run main.go
```

### CLI Command Equivalent:
```bash
# Encode with AV1 and Capped-CRF 28
mosaic -i input.mp4 -o ./output/hls_av1 --codec=av1 --crf=28

# Encode with HEVC (H.265) and Capped-CRF 23
mosaic -i input.mp4 -o ./output/hls_hevc --codec=hevc --crf=23
```

---

## 7. Direct S3 / MinIO Cloud Storage Upload
**Path**: `examples/s3_cloud_upload/main.go`

Demonstrates zero-dependency direct uploading of streaming assets (`.m3u8`, `.mpd`, `.m4s`, `.vtt`, `.jpg`) to S3/MinIO with AWS SigV4 signing, concurrent workers, and immutable cache headers.

### Run with Go:
```bash
cd examples/s3_cloud_upload && go run main.go
```

### CLI Command Equivalent:
```bash
mosaic -i input.mp4 -o ./output/hls \
  --s3-bucket my-stream-bucket \
  --s3-prefix vod/movie-101 \
  --s3-region us-east-1 \
  --s3-endpoint https://s3.us-east-1.amazonaws.com
```

---

## 8. Live Streaming (Low Latency)
**Path**: `examples/live_streaming/main.go`

Demonstrates packaging video using `ProfileLive` with short 2-second segments and low-latency CMAF flags for both HLS and DASH.

### Run with Go:
```bash
cd examples/live_streaming && go run main.go
```

### CLI Command Equivalent:
```bash
mosaic -i input.mp4 -o ./output/hls_live --profile=live
```

---

## 9. Video Orientation Normalization
**Path**: `examples/orientation_normalization/main.go`

Demonstrates probing rotation metadata, running standalone physical video normalization with `NormalizeVideoOrientation`, and HLS packaging with `WithNormalizeOrientation()`.

### Run with Go:
```bash
cd examples/orientation_normalization && go run main.go
```

### CLI Command Equivalent:
```bash
mosaic -i mobile_video.mp4 -o ./output/hls_rotated --normalize
```

---

## 10. Visual Terminal Progress Bar
**Path**: `examples/progress_monitoring/main.go`

Demonstrates rendering an interactive Unicode progress bar in the terminal (`[██████░░░░] 60.0%`) with speed, bitrate, elapsed time, and peak memory (RSS) summary.

### Run with Go:
```bash
cd examples/progress_monitoring && go run main.go
```

### CLI Command Equivalent:
```bash
mosaic -i input.mp4 -o ./output/hls --log-level=info
```

---

## 11. Multi-GPU Hardware Acceleration
**Path**: `examples/multi_gpu/main.go`

Tests and compares GPU-accelerated encoding using NVIDIA NVENC, Intel/AMD VAAPI, and Apple VideoToolbox.

### Run with Go:
```bash
cd examples/multi_gpu && go run main.go
```

### CLI Command Equivalent:
```bash
# NVIDIA GPU
mosaic -i input.mp4 -o ./output/hls --gpu=nvenc

# Intel / AMD VAAPI GPU
mosaic -i input.mp4 -o ./output/hls --gpu=vaapi

# Apple VideoToolbox (macOS)
mosaic -i input.mp4 -o ./output/hls --gpu=videotoolbox
```
