# Testing Guide

This guide covers local testing, mock-based tests, integration tests, and smoke tests.

## Fast Test Suite

```bash
go test ./...
```

If the default Go build cache is not writable:

```bash
GOCACHE=/tmp/go-build go test ./...
```

## Vet

```bash
GOCACHE=/tmp/go-build go vet ./...
```

## Coverage

```bash
GOCACHE=/tmp/go-build go test ./... -cover
```

## Lint

```bash
golangci-lint run
```

Run lint before commits when `golangci-lint` is installed.

## Test Categories

### Unit Tests

Unit tests cover:

- FFprobe JSON parsing.
- FPS parsing.
- Orientation metadata parsing.
- Rotation filter selection.
- Ladder generation.
- Bitrate optimization.
- HLS filter graph generation.
- DASH and HLS FFmpeg argument construction.
- Progress parsing.
- Executor behavior.

### Mocked Command Tests

Most encode tests use a fake `executor.CommandExecutor`.

This avoids:

- requiring FFmpeg in every unit test
- network access
- slow media processing
- fixture media files

### Integration-Style Tests

Some tests replace `executor.DefaultExecutor` with a mock and exercise public wrappers.

Real FFmpeg and FFprobe smoke tests are best run manually because they depend on installed binaries and fixture media.

## Smoke Test: Local File

Create a short input:

```bash
ffmpeg -y \
  -f lavfi -i testsrc2=size=1080x1080:duration=3 \
  -c:v libx264 \
  /tmp/mosaic-square.mp4
```

Run a small Go program:

```go
package main

import (
	"context"
	"log"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	_, err := mosaic.EncodeHls(
		context.Background(),
		mosaic.Job{
			Input:     "/tmp/mosaic-square.mp4",
			OutputDir: "/tmp/mosaic-square-hls",
			Profile:   mosaic.ProfileVOD,
		},
		mosaic.WithNormalizeOrientation(),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

Validate dimensions:

```bash
for f in /tmp/mosaic-square-hls/init*.mp4; do
  echo "$f"
  ffprobe -v error -select_streams v:0 \
    -show_entries stream=width,height:stream_tags=rotate:stream_side_data=rotation \
    -of csv=p=0 "$f"
done
```

Expected dimensions for `1080x1080`:

```text
1080,1080
720,720
360,360
```

## Smoke Test: Rotated Portrait

Use a video with stored landscape dimensions and `-90` display rotation.

Expected behavior:

- Output is portrait.
- Rotation metadata is cleared.
- Frame is not upside down.
- Aspect ratio is preserved.

Example expected dimensions for display `1080x1920`:

```text
608,1080
404,720
202,360
```

## Smoke Test: Remote URL

Remote inputs require FFmpeg and FFprobe network access.

```go
job := mosaic.Job{
	Input:     "https://example.com/video.mp4",
	OutputDir: "/tmp/mosaic-url-hls",
	Profile:   mosaic.ProfileVOD,
}
_, err := mosaic.EncodeHls(context.Background(), job, mosaic.WithNormalizeOrientation())
```

If DNS or network is blocked, FFprobe fails before encoding:

```text
Failed to resolve hostname
Input/output error
```

## Inspecting Output

List files:

```bash
find /tmp/mosaic-output -maxdepth 1 -type f -printf '%f\n' | sort
```

Probe an init segment:

```bash
ffprobe -v error -select_streams v:0 \
  -show_entries stream=width,height,avg_frame_rate:stream_tags=rotate:stream_side_data=rotation \
  -of json /tmp/mosaic-output/init_0.mp4
```

Extract a frame:

```bash
ffmpeg -y -ss 2 -i /tmp/mosaic-output/stream_0.m3u8 \
  -frames:v 1 /tmp/mosaic-frame.png
```

## Common Test Environment Issue

If tests fail with:

```text
read-only file system
```

set `GOCACHE` to a writable directory:

```bash
GOCACHE=/tmp/go-build go test ./...
```
