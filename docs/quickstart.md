# Quick Start

Get started with **Mosaic** in minutes using either the standalone CLI or the Go library.

---

## 1. CLI Quick Start

### Install CLI
```bash
go install github.com/farshidrezaei/mosaic/cmd/mosaic@latest
```

### Basic Packaging with Thumbnails
```bash
mosaic -i input.mp4 -o ./output/hls --thumbnails
```

### Next-Gen AV1 with Watermark & Audio Normalization
```bash
mosaic -i input.mp4 -o ./output/hls_av1 \
  --codec av1 \
  --crf 28 \
  --watermark ./logo.png \
  --normalize-audio \
  --thumbnails
```

### Encrypt HLS Segments with AES-128
```bash
mosaic -i input.mp4 -o ./output/hls_encrypted --encrypt-aes128
```

### Test in Web Player (DevTools Preview)
```bash
mosaic preview ./output/hls
```
Open **`http://localhost:8080`** in your browser to inspect and test the stream with live rendition switching, audio selector, and scrub thumbnails!

---

## 2. Go Library Usage

### Install Go Package
```bash
go get github.com/farshidrezaei/mosaic
```

### Complete HLS Packaging Pipeline

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
		mosaic.WithNormalizeOrientation(), // Automatically physically transposes mobile 90°/270° videos
		mosaic.WithNormalizeAudio(),       // Broadcast-standard EBU R128 loudness leveling
		mosaic.WithThumbnails(),           // Generates storyboard sprite sheet & thumbnails.vtt
		mosaic.WithWatermark(mosaic.WatermarkConfig{
			Path:     "./branding/logo.png",
			Position: mosaic.PositionTopRight,
			Opacity:  0.85,
		}),
		mosaic.WithSubtitles(mosaic.SubtitleTrack{
			Path:     "./subtitles/en.srt", // Auto-converted to WebVTT
			Language: "en",
			Label:    "English",
			Default:  true,
		}),
		mosaic.WithThreads(4),
	)
	if err != nil {
		log.Fatalf("HLS Encoding failed: %v", err)
	}

	fmt.Printf("\nEncoding completed! CPU User Time: %.2fs | Peak RSS: %d KB\n",
		usage.UserTime, usage.MaxMemory)
}
```

### Generated Files

The output directory `./output/hls` will contain:

```text
master.m3u8          # Multi-variant master playlist with subtitles
stream_0.m3u8        # 1080p stream playlist
stream_1.m3u8        # 720p stream playlist
stream_2.m3u8        # 360p stream playlist
sub_0.m3u8           # English subtitle playlist
sub_0.vtt            # WebVTT subtitle file
thumbnails_0.jpg     # Storyboard sprite sheet
thumbnails.vtt       # WebVTT timeline cue coordinates
init_0.mp4           # fMP4 initialization segment for variant 0
seg_0_0.m4s          # Media segment 0 of variant 0
seg_0_1.m4s          # Media segment 1 of variant 0
...
```

---

## 3. DASH CMAF Packaging

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
