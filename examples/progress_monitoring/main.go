package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/farshidrezaei/mosaic"
)

func renderProgressBar(percentage float64, width int) string {
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}
	completed := int(float64(width) * (percentage / 100.0))
	remaining := width - completed

	return fmt.Sprintf("[%s%s]", strings.Repeat("█", completed), strings.Repeat("░", remaining))
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v", err)
	}

	inputPath := filepath.Join(cwd, "input.mp4")
	outputDir := filepath.Join(cwd, "output", "hls_monitored")

	if _, statErr := os.Stat(inputPath); os.IsNotExist(statErr) {
		log.Printf("input file not found: %s", inputPath)
		log.Printf("place a video file named input.mp4 in %s", cwd)
		return
	}

	job := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDir,
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			bar := renderProgressBar(info.Percentage, 25)
			fmt.Printf("\r%s %5.1f%% | Time: %-11s | Speed: %-5s | Bitrate: %-12s",
				bar,
				info.Percentage,
				info.CurrentTime,
				info.Speed,
				info.Bitrate,
			)
		},
	}

	fmt.Printf("Starting progress-monitored HLS encoding...\n")
	start := time.Now()

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithBFrames(2),
		mosaic.WithThreads(4),
	)
	fmt.Println()
	if err != nil {
		log.Fatalf("Encoding failed: %v", err)
	}

	fmt.Printf("\nEncoding Summary:\n")
	fmt.Printf("  Total Wall Time:  %s\n", time.Since(start).Round(time.Second))
	if usage != nil {
		fmt.Printf("  CPU User Time:    %.2fs\n", usage.UserTime)
		fmt.Printf("  CPU System Time:  %.2fs\n", usage.SystemTime)
		fmt.Printf("  Peak Memory (RSS): %d KB (%.1f MB)\n", usage.MaxMemory, float64(usage.MaxMemory)/1024.0)
	}
	fmt.Printf("  Output Directory: %s\n", outputDir)
}
