package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	// This example demonstrates advanced DASH encoding with:
	// 1. Low-latency live profile
	// 2. Hardware acceleration (NVENC)
	// 3. Custom progress reporting
	// 4. Thread optimization

	cwd, _ := os.Getwd()
	inputPath := filepath.Join(cwd, "input.mp4")
	outputDir := filepath.Join(cwd, "output", "dash_advanced")

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		log.Printf("⚠️  Warning: %s not found. Please place a video file named 'input.mp4' in the current directory.", inputPath)
		return
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	job := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDir,
		Profile:   mosaic.ProfileLive, // Live profile uses 2-second segments for lower latency
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r📦 DASH Encoding: %.1f%% | Time: %s | Speed: %s",
				info.Percentage, info.CurrentTime, info.Speed)
		},
	}

	fmt.Printf("🎬 Starting Advanced DASH encoding for: %s\n", inputPath)
	start := time.Now()

	// Execute DASH encoding with advanced options
	_, err := mosaic.EncodeDash(context.Background(), job,
		mosaic.WithNVENC(),    // Use NVIDIA hardware acceleration
		mosaic.WithThreads(0), // 0 means auto-detect optimal thread count
		mosaic.WithLogLevel("warning"),
	)

	fmt.Println()

	if err != nil {
		log.Fatalf("❌ DASH Encoding failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("✅ DASH Encoding completed in %v!\n", duration.Round(time.Second))
	fmt.Printf("📂 Output available at: %s\n", outputDir)
}
