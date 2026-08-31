package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	input := "input.mp4"
	if len(os.Args) > 1 {
		input = os.Args[1]
	}

	job := mosaic.Job{
		Input:     input,
		OutputDir: "./output/hls_watermarked",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\rPackaging: [%5.1f%%] time=%s bitrate=%s",
				info.Percentage, info.CurrentTime, info.Bitrate)
		},
	}

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithNormalizeAudio(), // EBU R128 loudness leveling
		mosaic.WithWatermark(mosaic.WatermarkConfig{
			Path:     "./branding/logo.png",
			Position: mosaic.PositionTopRight,
			OffsetX:  16,
			OffsetY:  16,
			Opacity:  0.85,
		}),
		mosaic.WithSubtitles(
			mosaic.SubtitleTrack{
				Path:     "./subtitles/en.srt", // Auto converted to WebVTT
				Language: "en",
				Label:    "English",
				Default:  true,
			},
			mosaic.SubtitleTrack{
				Path:     "./subtitles/fa.srt",
				Language: "fa",
				Label:    "Persian",
			},
		),
	)
	if err != nil {
		log.Fatalf("Packaging failed: %v", err)
	}

	fmt.Printf("\nDone! Stream generated with watermark and subtitles. CPU User Time: %.2fs\n", usage.UserTime)
}
