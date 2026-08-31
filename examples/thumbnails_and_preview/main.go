package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/farshidrezaei/mosaic"
	"github.com/farshidrezaei/mosaic/preview"
)

func main() {
	input := "input.mp4"
	if len(os.Args) > 1 {
		input = os.Args[1]
	}

	outDir := "./output/hls_thumbnails"

	job := mosaic.Job{
		Input:     input,
		OutputDir: outDir,
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\rPackaging: [%5.1f%%] time=%s speed=%s",
				info.Percentage, info.CurrentTime, info.Speed)
		},
	}

	fmt.Printf("Packaging %s into HLS with Storyboard Thumbnails...\n", input)
	_, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithThumbnails(mosaic.ThumbnailConfig{
			IntervalSeconds: 2,
			TileWidth:       160,
			TileHeight:      90,
			Columns:         5,
			Rows:            5,
			Quality:         3,
		}),
	)
	if err != nil {
		log.Fatalf("Packaging failed: %v", err)
	}

	fmt.Printf("\nPackaging complete! Starting local web player on http://localhost:8080\n")
	server := preview.NewServer(outDir, 8080)

	if err := server.Start(); err != nil {
		log.Fatalf("Preview server failed: %v", err)
	}
}
