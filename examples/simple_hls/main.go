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
	// 1. Setup paths
	// In a real application, you would use absolute paths.
	// For this example, we assume an 'input.mp4' exists in the current directory.
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	inputPath := filepath.Join(cwd, "input.mp4")
	outputDir := filepath.Join(cwd, "output", "hls_simple")

	// Ensure input exists for a better error message
	if _, serr := os.Stat(inputPath); os.IsNotExist(serr) {
		log.Printf("⚠️  Warning: %s not found. Please place a video file named 'input.mp4' in the current directory.", inputPath)
		return
	}

	// Create output directory
	if merr := os.MkdirAll(outputDir, 0755); merr != nil {
		log.Fatalf("Failed to create output directory: %v", merr)
	}

	// 2. Configure the encoding job
	job := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDir,
		Profile:   mosaic.ProfileVOD, // VOD profile uses 5-second segments
		ProgressHandler: func(info mosaic.ProgressInfo) {
			// This callback is triggered as FFmpeg reports progress
			fmt.Printf("\r🚀 Progress: %.1f%% | Time: %s | Speed: %s | Bitrate: %s",
				info.Percentage, info.CurrentTime, info.Speed, info.Bitrate)
		},
	}

	fmt.Printf("🎬 Starting HLS encoding for: %s\n", inputPath)
	start := time.Now()

	// 3. Execute encoding with options
	// We use 4 threads and set the log level to warning to keep the console clean.
	stats, err := mosaic.EncodeHls(context.Background(), job, mosaic.WithLogLevel("warning"))

	fmt.Println() // New line after progress reporting

	if err != nil {
		log.Fatalf("❌ Encoding failed: %v", err)
	}

	if stats != nil {
		fmt.Printf("User Time: %f, System Time: %f, Max Memory: %d\n", stats.UserTime, stats.SystemTime, stats.MaxMemory)
	}

	duration := time.Since(start)
	fmt.Printf("✅ Encoding completed in %v!\n", duration.Round(time.Second))
	fmt.Printf("📂 Output saved to: %s\n", outputDir)
}
