# Encoding Behavior

This document describes how Mosaic chooses renditions and builds FFmpeg outputs.

## Probe Phase

Mosaic first calls FFprobe for the primary video stream.

Collected values:

- stored width
- stored height
- average frame rate
- rotation metadata from side data or tags

It then performs a second audio probe to determine if an audio stream exists.

## Display Dimensions

Stored dimensions are not always display dimensions.

Example:

```text
stored:   1920x1080
rotation: -90
display:  1080x1920
```

Mosaic builds the ladder from display dimensions.

## Ladder Generation

Base quality rungs are selected by display height:

| Source Display Height | Candidate Rungs |
|----------------------:|-----------------|
| `>= 1080`             | `1080`, `720`, `360` |
| `>= 720`              | `720`, `360` |
| `>= 360`              | `360` |
| `< 360`               | source display height |

The target width is computed from the source display aspect ratio:

```text
target_width = round(target_height * source_display_width / source_display_height)
```

Both width and height are made even.

## Aspect Ratio Examples

| Input Display Size | Initial Rungs |
|--------------------|---------------|
| `1920x1080`        | `1920x1080`, `1280x720`, `640x360` |
| `1080x1080`        | `1080x1080`, `720x720`, `360x360` |
| `1080x1920`        | `608x1080`, `404x720`, `202x360` |
| `1280x718`         | `642x360` |
| `426x240`          | `426x240` |

## Bitrate Optimization

After ladder generation, `optimize.Apply` adjusts bitrates and buffer sizes.

Current cap rules:

| Rung Height | Max Bitrate Cap |
|------------:|----------------:|
| `>= 1080`   | `5000k` |
| `>= 720`    | `3000k` |
| lower       | `1000k` |

The VBV buffer size is set to `MaxRate * 2`.

The optimizer also trims rungs where the height ratio relative to the previous kept rung is not meaningfully different.

## HLS CMAF

HLS output uses FFmpeg's HLS muxer:

```text
-f hls
-hls_segment_type fmp4
-master_pl_name master.m3u8
-var_stream_map ...
```

Video processing:

```text
split=N
scale=width:height
setsar=1
```

Mosaic does not pad HLS outputs into a fixed frame. The frame dimensions are the aspect-preserving ladder dimensions.

HLS VOD profile:

```text
-hls_time 5
-hls_flags independent_segments
```

HLS live profile:

```text
-hls_time 2
-hls_part_size 0.5
-hls_flags independent_segments+split_by_time
```

## DASH CMAF

DASH output uses FFmpeg's DASH muxer:

```text
-f dash
-seg_duration <profile duration>
-use_template 1
-use_timeline 1
```

Video renditions are mapped from the original video stream and sized with per-stream `-s:v:N`.

Generated manifest:

```text
manifest.mpd
```

## Audio Handling

If an audio stream exists, Mosaic maps the first audio stream for each variant:

```text
-map a:0
-c:a:N aac
-b:a:N 96k
-ac 2
```

If no audio stream exists, audio mapping is omitted.

## Orientation Normalization

When `WithNormalizeOrientation()` is enabled:

1. Mosaic probes orientation metadata.
2. If rotation is `90`, `180`, or `270`, it runs FFmpeg with `-noautorotate` and an explicit transpose filter.
3. The normalized temporary file is verified to have rotation `0`.
4. Encoding proceeds from the temporary normalized file.
5. The temporary input is removed after encoding.

Rotation filter mapping:

| Normalized Rotation | Filter |
|--------------------:|--------|
| `90`                | `transpose=2` |
| `180`               | `transpose=1,transpose=1` |
| `270` or `-90`      | `transpose=1` |

Mosaic also clears output video rotation metadata with:

```text
-metadata:s:v:N rotate=0
```

## Hardware Encoders

Default video codec:

```text
libx264
```

Option mappings:

| Option | FFmpeg Encoder |
|--------|----------------|
| `WithNVENC()` | `h264_nvenc` |
| `WithVAAPI()` | `h264_vaapi` |
| `WithVideoToolbox()` | `h264_videotoolbox` |

Hardware encoder options do not add upload filters or platform-specific setup.

## GOP

GOP size is based on frame rate and segment duration:

```text
gop = round(fps * segment_duration)
```

Rules:

- GOP is rounded up to an even value.
- Minimum GOP is `24`.
- `-keyint_min` matches GOP.
- `-sc_threshold` is `0`.

## Output Metadata

Mosaic clears rotation metadata for generated video streams to avoid double-rotation in players.

It sets sample aspect ratio to `1` in HLS filter graphs.

## Known Limitations

- Default codec is H.264.
- Audio is always AAC at `96k` when present.
- Per-title encoding is not implemented.
- Percentage progress is not computed yet.
- DASH does not currently use the same filter graph path as HLS.
