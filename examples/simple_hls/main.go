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
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v", err)
	}

	inputPath := filepath.Join(cwd, "input.mp4")
	outputDir := filepath.Join(cwd, "output", "hls_simple")

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		log.Printf("input file not found: %s", inputPath)
		log.Printf("place a video file named input.mp4 in %s", cwd)
		return
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	job := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDir,
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\rtime=%s speed=%s bitrate=%s", info.CurrentTime, info.Speed, info.Bitrate)
		},
	}

	fmt.Printf("starting HLS encoding: %s\n", inputPath)
	start := time.Now()

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithThreads(4),
		mosaic.WithLogLevel("warning"),
	)
	fmt.Println()
	if err != nil {
		log.Fatalf("HLS encoding failed: %v", err)
	}

	if usage != nil {
		fmt.Printf("usage: user=%.2fs system=%.2fs maxrss=%d\n", usage.UserTime, usage.SystemTime, usage.MaxMemory)
	}

	fmt.Printf("completed in %s\n", time.Since(start).Round(time.Second))
	fmt.Printf("output: %s\n", outputDir)
}
