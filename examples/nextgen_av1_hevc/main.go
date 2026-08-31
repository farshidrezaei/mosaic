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

	fmt.Println("--- 1. Packaging AV1 HLS Stream with Capped-CRF 28 ---")
	jobAV1 := mosaic.Job{
		Input:     input,
		OutputDir: "./output/hls_av1",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[AV1]  [%5.1f%%] time=%s bitrate=%s",
				info.Percentage, info.CurrentTime, info.Bitrate)
		},
	}

	_, err := mosaic.EncodeHls(
		context.Background(),
		jobAV1,
		mosaic.WithAV1(),
		mosaic.WithCRF(28),
		mosaic.WithThumbnails(),
	)
	if err != nil {
		log.Fatalf("AV1 encoding failed: %v", err)
	}

	fmt.Println("\n\n--- 2. Packaging HEVC HLS Stream with Capped-CRF 24 ---")
	jobHEVC := mosaic.Job{
		Input:     input,
		OutputDir: "./output/hls_hevc",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[HEVC] [%5.1f%%] time=%s bitrate=%s",
				info.Percentage, info.CurrentTime, info.Bitrate)
		},
	}

	_, err = mosaic.EncodeHls(
		context.Background(),
		jobHEVC,
		mosaic.WithHEVC(),
		mosaic.WithCRF(24),
		mosaic.WithThumbnails(),
	)
	if err != nil {
		log.Fatalf("HEVC encoding failed: %v", err)
	}

	fmt.Println("\nAll next-gen streams packaged successfully!")
}
