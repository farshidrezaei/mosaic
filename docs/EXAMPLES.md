# Examples Catalog

Explore all runnable examples located in the [`examples/`](https://github.com/farshidrezaei/mosaic/tree/main/examples) directory.

---

## 1. Simple HLS Packaging

**Path**: `examples/simple_hls/main.go`

Demonstrates straightforward HLS VOD packaging with real-time percentage progress tracking and automatic orientation normalization.

```bash
cd examples/simple_hls
go run main.go
```

---

## 2. Advanced DASH CMAF

**Path**: `examples/advanced_dash/main.go`

Demonstrates DASH CMAF packaging with custom B-Frames (`WithBFrames(2)`), high-framerate bitrate scaling (`WithScaleBitrateWithFPS()`), and custom thread count.

```bash
cd examples/advanced_dash
go run main.go
```

---

## 3. Live Streaming (Low Latency)

**Path**: `examples/live_streaming/main.go`

Demonstrates packaging video using `ProfileLive` with short 2-second segments and low-latency CMAF flags for both HLS and DASH.

```bash
cd examples/live_streaming
go run main.go
```

---

## 4. Video Orientation Normalization

**Path**: `examples/orientation_normalization/main.go`

Demonstrates probing rotation metadata, running standalone physical video normalization with `NormalizeVideoOrientation`, and HLS packaging with `WithNormalizeOrientation()`.

```bash
cd examples/orientation_normalization
go run main.go
```

---

## 5. Visual Terminal Progress Bar

**Path**: `examples/progress_monitoring/main.go`

Demonstrates rendering an interactive Unicode progress bar in the terminal (`[██████░░░░] 60.0%`) with speed, bitrate, elapsed time, and peak memory (RSS) summary.

```bash
cd examples/progress_monitoring
go run main.go
```

---

## 6. Multi-GPU Hardware Acceleration

**Path**: `examples/multi_gpu/main.go`

Tests and compares GPU-accelerated encoding using NVIDIA NVENC, Intel/AMD VAAPI, and Apple VideoToolbox.

```bash
cd examples/multi_gpu
go run main.go
```
