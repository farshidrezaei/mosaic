package mosaic_test

import (
	"context"
	"fmt"
	"log"

	"github.com/farshidrezaei/mosaic"
)

// ExampleEncodeHls demonstrates packaging a video file into multi-rendition HLS (fMP4 CMAF).
func ExampleEncodeHls() {
	job := mosaic.Job{
		Input:     "input.mp4",
		OutputDir: "./output/hls",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("Progress: %.1f%%\n", info.Percentage)
		},
	}

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithThreads(4),
	)
	if err != nil {
		log.Fatalf("HLS encoding failed: %v", err)
	}

	fmt.Printf("Peak Memory: %d KB\n", usage.MaxMemory)
}

// ExampleEncodeDash demonstrates packaging a video file into MPEG-DASH CMAF format.
func ExampleEncodeDash() {
	job := mosaic.Job{
		Input:     "input.mp4",
		OutputDir: "./output/dash",
		Profile:   mosaic.ProfileVOD,
	}

	_, err := mosaic.EncodeDash(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithBFrames(2),
		mosaic.WithScaleBitrateWithFPS(),
	)
	if err != nil {
		log.Fatalf("DASH encoding failed: %v", err)
	}
}

// ExampleNormalizeVideoOrientation demonstrates standalone video orientation normalization.
func ExampleNormalizeVideoOrientation() {
	err := mosaic.NormalizeVideoOrientation(
		context.Background(),
		"input_rotated.mp4",
		"output_normalized.mp4",
	)
	if err != nil {
		log.Fatalf("Normalization failed: %v", err)
	}
}
