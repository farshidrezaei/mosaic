# Video Orientation Normalization

Modern smartphones (iOS and Android) record video with fixed physical sensors and embed rotation metadata (such as a display matrix or a container `rotate` tag) indicating how the video should be oriented during playback.

---

## The Problem with Rotation Metadata

1. **Player Inconsistency**: Many web players, HTML5 `<video>` tags on older browsers, and streaming clients ignore rotation metadata or render the video sideways or upside down.
2. **Double Rotation**: If transcoding re-encodes frames without stripping the rotation metadata tag, modern players rotate the already-rotated video twice.
3. **Black Bars**: Legacy encoding ladders letterbox vertical video into a 16:9 box instead of building a native portrait ladder.

---

## Mosaic's Two-Layer Solution

### 1. Display-Aware Ladder Probing
Mosaic probes video using display dimensions, not raw matrix dimensions:
```go
info, err := probe.Input(ctx, "mobile_video.mp4")
fmt.Println(info.DisplayWidth(), info.DisplayHeight(), info.IsPortrait())
```
If a video is stored as `1920x1080` with `-90°` rotation, Mosaic treats it as `1080x1920` (portrait) and builds portrait rungs: `608x1080`, `404x720`, `202x360`.

### 2. Physical Normalization (`WithNormalizeOrientation`)
When `mosaic.WithNormalizeOrientation()` is enabled:
- Rotated frames are physically transposed with FFmpeg `-noautorotate` and transpose filters (`transpose=1`, `transpose=2`).
- The output rotation metadata tag is set to `rotate=0`.
- The temporary output is verified to ensure zero remaining rotation tags.

```go
usage, err := mosaic.EncodeHls(
    ctx,
    job,
    mosaic.WithNormalizeOrientation(),
)
```

---

## Standalone Normalization API

You can also use the standalone helper function outside of HLS/DASH packaging:

```go
err := mosaic.NormalizeVideoOrientation(ctx, "input.mp4", "output_normalized.mp4")
```
