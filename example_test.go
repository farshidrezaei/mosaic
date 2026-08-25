package mosaic_test

import (
	"context"
	"fmt"

	"github.com/farshidrezaei/mosaic"
)

// ExampleJob demonstrates constructing an encoding job configuration.
func ExampleJob() {
	job := mosaic.Job{
		Input:     "input.mp4",
		OutputDir: "./output/hls",
		Profile:   mosaic.ProfileVOD,
	}

	fmt.Println("Input:", job.Input)
	fmt.Println("OutputDir:", job.OutputDir)
	fmt.Println("Profile:", job.Profile)
	// Output:
	// Input: input.mp4
	// OutputDir: ./output/hls
	// Profile: vod
}

// ExampleEncodeHls demonstrates packaging a video file into multi-rendition HLS (fMP4 CMAF).
func ExampleEncodeHls() {
	ctx := context.Background()
	job := mosaic.Job{
		Input:     "input.mp4",
		OutputDir: "./output/hls",
		Profile:   mosaic.ProfileVOD,
	}

	// Invoke EncodeHls with options
	_, err := mosaic.EncodeHls(
		ctx,
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithThreads(4),
	)
	if err != nil {
		// In a real environment with input.mp4, this packages HLS streams
		_ = err
	}
}

// ExampleEncodeDash demonstrates packaging a video file into MPEG-DASH CMAF format.
func ExampleEncodeDash() {
	ctx := context.Background()
	job := mosaic.Job{
		Input:     "input.mp4",
		OutputDir: "./output/dash",
		Profile:   mosaic.ProfileVOD,
	}

	// Invoke EncodeDash with options
	_, err := mosaic.EncodeDash(
		ctx,
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithBFrames(2),
		mosaic.WithScaleBitrateWithFPS(),
	)
	if err != nil {
		// In a real environment with input.mp4, this packages DASH streams
		_ = err
	}
}

// ExampleNormalizeVideoOrientation demonstrates standalone video orientation normalization.
func ExampleNormalizeVideoOrientation() {
	ctx := context.Background()
	err := mosaic.NormalizeVideoOrientation(
		ctx,
		"input_rotated.mp4",
		"output_normalized.mp4",
	)
	if err != nil {
		// In a real environment with input_rotated.mp4, this transposes frames
		_ = err
	}
}
