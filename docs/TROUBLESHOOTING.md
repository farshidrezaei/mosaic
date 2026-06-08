# Troubleshooting

This document lists common failures and how to debug them.

## FFprobe Not Found

Symptom:

```text
exec: "ffprobe": executable file not found in $PATH
```

Fix:

Install FFmpeg and ensure `ffprobe` is on `PATH`.

```bash
ffprobe -version
ffmpeg -version
```

## FFmpeg Not Found

Symptom:

```text
exec: "ffmpeg": executable file not found in $PATH
```

Fix:

Install FFmpeg and ensure `ffmpeg` is on `PATH`.

## Remote URL Fails

Symptoms:

```text
Failed to resolve hostname
Input/output error
```

Likely causes:

- DNS blocked.
- Network egress blocked.
- URL expired.
- FFmpeg build lacks protocol support.

Debug:

```bash
ffprobe -v error -show_format -show_streams "https://example.com/video.mp4"
```

## Output Directory Permission Denied

Symptom:

```text
create output dir: mkdir /path: permission denied
```

Fix:

Use a writable output directory.

```bash
mkdir -p /tmp/mosaic-output
```

Mosaic creates directories automatically, but the process still needs permissions.

## Video Is Upside Down

Likely cause:

- Source rotation metadata was interpreted differently by the player.
- Orientation normalization was disabled.

Recommended option:

```go
mosaic.WithNormalizeOrientation()
```

Debug source metadata:

```bash
ffprobe -v error -select_streams v:0 \
  -show_entries stream=width,height:stream_tags=rotate:stream_side_data=rotation \
  -of json input.mp4
```

Debug output init segment:

```bash
ffprobe -v error -select_streams v:0 \
  -show_entries stream=width,height:stream_tags=rotate:stream_side_data=rotation \
  -of json /tmp/output/init_0.mp4
```

Output rotation metadata should be absent or `0`.

## Output Aspect Ratio Looks Wrong

Mosaic preserves display aspect ratio when building ladder dimensions.

Check source display dimensions:

```bash
ffprobe -v error -select_streams v:0 \
  -show_entries stream=width,height:stream_side_data=rotation:stream_tags=rotate \
  -of json input.mp4
```

Check output dimensions:

```bash
for f in /tmp/output/init*.mp4; do
  echo "$f"
  ffprobe -v error -select_streams v:0 \
    -show_entries stream=width,height \
    -of csv=p=0 "$f"
done
```

For a square input, outputs should be square.

For a portrait input, outputs should be portrait.

For a non-standard input, dimensions may differ by one or two pixels from the exact mathematical ratio because Mosaic forces even dimensions.

## Hardware Encoder Fails

Symptoms:

```text
Unknown encoder 'h264_nvenc'
No capable devices found
Error initializing output stream
```

Fix:

- Confirm FFmpeg supports the encoder:

```bash
ffmpeg -hide_banner -encoders | grep h264
```

- Confirm drivers and devices are available.
- Try CPU encoding by removing the GPU option.

## HLS Player Cannot Load Output

Check files:

```bash
find /tmp/output -maxdepth 1 -type f -printf '%f\n' | sort
```

Expected HLS files:

```text
master.m3u8
stream_0.m3u8
init*.mp4
seg_*.m4s
```

Serve the directory over HTTP for browser playback. Many browser/player combinations do not reliably load HLS from `file://`.

## DASH Player Cannot Load Output

Expected DASH files:

```text
manifest.mpd
init-stream*.m4s
chunk-stream*.m4s
```

Use an HTTP server for playback tests.

## Tests Fail Because Go Cache Is Read-Only

Symptom:

```text
open /home/user/.cache/go-build/...: read-only file system
```

Fix:

```bash
GOCACHE=/tmp/go-build go test ./...
```

## Collecting a Useful Bug Report

Include:

- input file or URL
- exact code snippet or command
- OS and architecture
- Go version
- FFmpeg version
- full error output
- source FFprobe JSON
- output init segment FFprobe JSON
- expected output dimensions
- actual output dimensions
