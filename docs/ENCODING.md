# Encoding & Pipeline Architecture

This document describes how Mosaic constructs rendition ladders, calculates display aspect ratios, configures next-gen video codecs, applies overlays, and builds optimal single-pass FFmpeg pipelines.

---

## 1. Probe & Display Dimensions

Mosaic queries FFprobe to inspect input stream metadata:

- Display width & height (accounting for container rotation tags)
- Average frame rate (FPS)
- Media duration
- Audio stream presence and channels
- Orientation metadata (matrix side data and tags)

### Orientation vs Stored Dimensions

Mobile and modern smartphone cameras often record rotated frames (e.g., stored `1920x1080` with `-90°` rotation). Mosaic computes all ladders from **true display dimensions** (`1080x1920`), ensuring correct portrait/landscape resolution tiers without distortion.

---

## 2. Aspect-Preserving Ladder Generation

Base rendition candidates are determined by display height:

| Source Display Height | Candidate Tiers |
|----------------------:|-----------------|
| `>= 1080`             | `1080p`, `720p`, `360p` |
| `>= 720`              | `720p`, `360p` |
| `>= 360`              | `360p` |
| `< 360`               | Source display height |

The width of each rendition is computed from the exact display aspect ratio:

\[
\text{target\_width} = 2 \times \text{round}\left(\frac{\text{target\_height} \times \text{display\_width}}{2 \times \text{display\_height}}\right)
\]

---

## 3. Video Codecs & Hardware Acceleration

Mosaic supports software and hardware encoders across all major codecs:

| Codec Constant | Software Encoder | NVIDIA NVENC | Intel/AMD VAAPI | Apple VideoToolbox |
|---|---|---|---|---|
| `CodecH264` (`"h264"`) | `libx264` | `h264_nvenc` | `h264_vaapi` | `h264_videotoolbox` |
| `CodecHEVC` (`"hevc"`) | `libx265` | `hevc_nvenc` | `hevc_vaapi` | `hevc_videotoolbox` |
| `CodecAV1` (`"av1"`) | `libsvtav1` | `av1_nvenc` | `av1_vaapi` | *(N/A)* |

### Capped-CRF (Content-Aware Bitrate Control)

When `WithCRF(crf)` is enabled, Mosaic injects Constant Rate Factor parameters combined with VBV maximum bitrate and buffer size constraints (`-maxrate` and `-bufsize`). This ensures:
- Simple scenes (talking heads, static backgrounds) consume significantly less bandwidth.
- Complex scenes with high motion are strictly capped to prevent buffer overruns and network congestion.

---

## 4. Single-Pass Filter Complex Graph

Mosaic generates a unified `filter_complex` graph to process all ladder tiers in a single FFmpeg pass:

```text
[0:v]split=3[v0][v1][v2];
[v0]scale=1920:1080,setsar=1,setdar=16/9[v0o];
[v1]scale=1280:720,setsar=1,setdar=16/9[v1o];
[v2]scale=640:360,setsar=1,setdar=16/9[v2o]
```

### Display Aspect Ratio (`setdar`) Consistency

In DASH Adaptation Sets, FFmpeg strictly enforces identical Display Aspect Ratio (DAR) across all representations. Mosaic automatically calculates and injects `setdar=sourceW/sourceH` into every scale chain, eliminating any rounding mismatch errors across different resolutions.

### Dynamic Watermark Overlay Graph

When `WithWatermark()` is configured, Mosaic injects the watermark image into the filter graph:

```text
[1:v]format=rgba,colorchannelmixer=aa=0.85[wm];
[0:v]split=3[v0][v1][v2];
[wm]split=3[wm0][wm1][wm2];
[v0]scale=1920:1080,setsar=1,setdar=16/9[v0s];
[wm0]scale=288:-1[wm0s];
[v0s][wm0s]overlay=main_w-overlay_w-16:16[v0o];
...
```

---

## 5. Audio Normalization (EBU R128)

When `WithNormalizeAudio()` is enabled, Mosaic applies the ITU-R BS.1770 / EBU R128 loudness filter:

```text
-filter:a loudnorm=I=-16:TP=-1.5:LRA=11
```

- **Integrated Loudness (`I`)**: `-16.0 LUFS` (optimal target for web & mobile streaming).
- **True Peak (`TP`)**: `-1.5 dBFS` (prevents inter-sample clipping on DAC decoders).
- **Loudness Range (`LRA`)**: `11.0 LU`.

---

## 6. Storyboard Thumbnails (Sprite Sheet + WebVTT)

When `WithThumbnails()` is enabled, Mosaic extracts frames at regular intervals and packs them into a grid sprite:

```bash
ffmpeg -y -i input.mp4 -vf "select=not(mod(n\,48)),scale=160:90,tile=5x5" -start_number 0 -q:v 3 thumbnails_%d.jpg
```

Mosaic automatically generates the matching `thumbnails.vtt` file with precise spatial cue tags:

```text
WEBVTT

00:00:00.000 --> 00:00:02.000
thumbnails_0.jpg#xywh=0,0,160,90

00:00:02.000 --> 00:00:04.000
thumbnails_0.jpg#xywh=160,0,160,90
```

---

## 7. Security & HLS AES-128 Encryption

When `WithAES128Encryption()` is enabled:

1. Mosaic creates a 16-byte random key (`enc.key`) via `crypto/rand`.
2. Mosaic creates `enc.keyinfo` containing the playlist URI and local key path.
3. FFmpeg encrypts each media segment with AES-128-CBC and writes `#EXT-X-KEY:METHOD=AES-128,URI="..."` into every variant playlist.
