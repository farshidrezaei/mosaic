# Quick Start

Get started with **Mosaic** in minutes.

---

## 1. Installation

Ensure you have Go `1.25+` and FFmpeg `4.4+` installed:

```bash
go get github.com/farshidrezaei/mosaic
```

---

## 2. HLS Packaging Example

Here is a complete, production-ready example to package a video into HLS with CMAF segments:

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
		Input:     "input.mp4",
		OutputDir: "./output/hls",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[%5.1f%%] time=%s bitrate=%s speed=%s",
				info.Percentage, info.CurrentTime, info.Bitrate, info.Speed)
		},
	}

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(), // Automatically physically rotates mobile/portrait videos
		mosaic.WithThreads(4),             // 4 CPU encoding threads
	)
	if err != nil {
		log.Fatalf("HLS Encoding failed: %v", err)
	}

	fmt.Printf("\nEncoding completed!\n")
	if usage != nil {
		fmt.Printf("CPU Time: %.2fs user, %.2fs system | Peak RSS: %d KB\n",
			usage.UserTime, usage.SystemTime, usage.MaxMemory)
	}
}
```

### Generated Files

The output directory `./output/hls` will contain:

```text
master.m3u8          # Multi-variant master playlist
stream_0.m3u8        # 1080p stream playlist
stream_1.m3u8        # 720p stream playlist
stream_2.m3u8        # 360p stream playlist
init_0.mp4           # fMP4 initialization segment for variant 0
seg_0_0.m4s          # Media segment 0 of variant 0
seg_0_1.m4s          # Media segment 1 of variant 0
...
```

---

## 3. DASH CMAF Example

To package for MPEG-DASH:

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
		Input:     "input.mp4",
		OutputDir: "./output/dash",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[%5.1f%%] time=%s bitrate=%s speed=%s",
				info.Percentage, info.CurrentTime, info.Bitrate, info.Speed)
		},
	}

	_, err := mosaic.EncodeDash(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithBFrames(2),
		mosaic.WithScaleBitrateWithFPS(),
	)
	if err != nil {
		log.Fatalf("DASH Encoding failed: %v", err)
	}

	fmt.Println("\nDASH packaging complete -> ./output/dash/manifest.mpd")
}
```

### Generated Files

The output directory `./output/dash` will contain:

```text
manifest.mpd                              # MPD XML manifest
init-stream0.m4s                          # Init segment for stream 0
chunk-stream0-00001.m4s                   # Media chunks
...
```
