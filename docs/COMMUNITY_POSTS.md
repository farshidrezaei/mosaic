# Community Launch & Promotion Templates

Use these ready-to-copy templates to share Mosaic with the global Go and Video Engineering communities.

---

## 1. Reddit (r/golang)

**Title**:  
`Show r/golang: Mosaic – Predictable ABR video packaging (HLS/DASH CMAF) in Go with aspect preservation & mobile rotation fix`

**Body**:
```markdown
Hey r/golang! 👋

We built and open-sourced **Mosaic** (https://github.com/farshidrezaei/mosaic), a Go library and CLI tool for production-ready Adaptive Bitrate (ABR) video packaging into HLS (fMP4) and DASH CMAF.

### Why we built it:
Anyone who has worked with FFmpeg pipelines in production knows the pain:
1. **Letterboxing/Black Bars**: Legacy ladders assume 16:9 and distort square (1:1), portrait (9:16), or ultrawide videos.
2. **Mobile Rotation Bugs**: Videos shot on iOS/Android have display matrices/rotation tags (`-90°`). If transcoded naively, web players show them rotated sideways or upside down.
3. **Complex Filter Graphs**: Hand-crafting multi-rendition single-pass `filter_complex` with SAR consistency, GOP alignment, and segment numbering is error-prone.

### Key Features:
- ⚡ **Standard CMAF Packaging**: HLS (fMP4) and DASH output with clean single-pass filter graphs.
- 📐 **Aspect-Preserving Ladders**: Automatic ladder calculation for any aspect ratio (landscape, 1:1, 9:16 portrait) with zero black bars.
- 🔄 **Orientation Normalization**: Probes display matrices, physically transposes frames, and clears rotation tags.
- 📊 **Real-Time Progress**: 0.0% to 100.0% percentage calculation, speed factor, and encoded timestamp.
- 🚀 **Hardware Acceleration**: Automatic GPU profiles for NVIDIA NVENC, Intel/AMD VAAPI, and Apple VideoToolbox.
- 🛠️ **CLI Tool & Docker**: Installable via `go install github.com/farshidrezaei/mosaic/cmd/mosaic@latest` or Docker container.
- 🧪 **100% Mock Executor Testing**: No live FFmpeg needed in unit tests. Zero external dependencies.

Documentation: https://farshidrezaei.github.io/mosaic/
GitHub: https://github.com/farshidrezaei/mosaic

We'd love your feedback, bug reports, and contributions!
```

---

## 2. Hacker News (Show HN)

**Title**:  
`Show HN: Mosaic – Production-ready HLS and DASH CMAF video packaging for Go`

**URL**: `https://github.com/farshidrezaei/mosaic`

**Body (as first comment)**:
```text
Hi HN,

Mosaic is an open-source Go library and CLI tool for adaptive bitrate video packaging to HLS (fMP4) and DASH CMAF.

It solves three common headaches in video transcoding:
1. Automatically computes aspect-preserving rendition ladders (no letterboxing for vertical or square videos).
2. Physically normalizes mobile video orientation metadata (fixing rotated smartphone footage).
3. Provides a clean, mockable Go API with real-time percentage progress streaming.

Docs: https://farshidrezaei.github.io/mosaic/
Repo: https://github.com/farshidrezaei/mosaic

Feedback and questions are welcome!
```

---

## 3. Awesome-Go Pull Request

**Target Repository**: `avelino/awesome-go`  
**File**: `README.md` (under `Video` or `Media` section)

**Markdown Entry**:
```markdown
* [Mosaic](https://github.com/farshidrezaei/mosaic) - Predictable, production-ready Adaptive Bitrate (ABR) video packaging for Go (HLS & DASH CMAF).
```

---

## 4. Golang Weekly Submission

**Submission URL**: https://golangweekly.com/submit

- **Title**: Mosaic: Predictable ABR Video Packaging for Go
- **URL**: https://github.com/farshidrezaei/mosaic
- **Description**: A production-ready Go library and CLI for packaging video into standardized HLS (fMP4) and DASH CMAF with aspect-ratio preservation, orientation normalization, and GPU acceleration.
