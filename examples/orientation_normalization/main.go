package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/farshidrezaei/mosaic"
	"github.com/farshidrezaei/mosaic/probe"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v", err)
	}

	inputPath := filepath.Join(cwd, "rotated_input.mp4")
	if _, statErr := os.Stat(inputPath); os.IsNotExist(statErr) {
		inputPath = filepath.Join(cwd, "input.mp4")
	}

	if _, statErr := os.Stat(inputPath); os.IsNotExist(statErr) {
		log.Printf("input file not found: %s", inputPath)
		log.Printf("place a video file named input.mp4 or rotated_input.mp4 in %s", cwd)
		return
	}

	ctx := context.Background()

	// 1. Probe input video orientation metadata
	info, err := probe.Input(ctx, inputPath)
	if err != nil {
		log.Fatalf("failed to probe video: %v", err)
	}

	fmt.Printf("Source Video Metadata:\n")
	fmt.Printf("  Stored Dimensions:  %dx%d\n", info.Width, info.Height)
	fmt.Printf("  Rotation Metadata:  %d degrees\n", info.Rotation)
	fmt.Printf("  Display Dimensions: %dx%d (IsPortrait: %t)\n", info.DisplayWidth(), info.DisplayHeight(), info.IsPortrait())
	fmt.Printf("  Duration:           %.2fs (FPS: %.2f)\n\n", info.Duration, info.FPS)

	// 2. Standalone Normalization Example:
	// Creates a physically rotated MP4 with metadata cleared (for web/mobile players).
	normalizedPath := filepath.Join(cwd, "output", "normalized.mp4")
	_ = os.MkdirAll(filepath.Dir(normalizedPath), 0o755)

	fmt.Printf("Running standalone orientation normalization...\n")
	startNorm := time.Now()
	if normErr := mosaic.NormalizeVideoOrientation(ctx, inputPath, normalizedPath); normErr != nil {
		log.Fatalf("NormalizeVideoOrientation failed: %v", normErr)
	}
	fmt.Printf("Normalized video saved in %s (%s)\n\n", normalizedPath, time.Since(startNorm).Round(time.Millisecond))

	// 3. Automated HLS Encoding with Orientation Normalization option
	outputDir := filepath.Join(cwd, "output", "hls_normalized")
	job := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDir,
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[HLS Normalized] [%5.1f%%] time=%s bitrate=%s", info.Percentage, info.CurrentTime, info.Bitrate)
		},
	}

	fmt.Printf("Starting HLS encoding with WithNormalizeOrientation()...\n")
	startHLS := time.Now()
	usage, err := mosaic.EncodeHls(
		ctx,
		job,
		mosaic.WithNormalizeOrientation(), // Automatically normalizes and clears rotation metadata
		mosaic.WithThreads(4),
	)
	fmt.Println()
	if err != nil {
		log.Fatalf("HLS encoding failed: %v", err)
	}

	if usage != nil {
		fmt.Printf("Usage: user=%.2fs system=%.2fs maxrss=%d KB\n", usage.UserTime, usage.SystemTime, usage.MaxMemory)
	}
	fmt.Printf("HLS stream created in %s (%s)\n", outputDir, time.Since(startHLS).Round(time.Second))
}
