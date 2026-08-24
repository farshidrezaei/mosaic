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
	outputDirHLS := filepath.Join(cwd, "output", "live_hls")
	outputDirDASH := filepath.Join(cwd, "output", "live_dash")

	if _, statErr := os.Stat(inputPath); os.IsNotExist(statErr) {
		log.Printf("input file not found: %s", inputPath)
		log.Printf("place a video file named input.mp4 in %s", cwd)
		return
	}

	// 1. Low-Latency HLS Live Profile (2-second segments, split by time)
	fmt.Println("--- 1. Starting Low-Latency HLS Live Encoding ---")
	jobHLS := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDirHLS,
		Profile:   mosaic.ProfileLive,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[HLS LIVE] [%5.1f%%] time=%s speed=%s bitrate=%s", info.Percentage, info.CurrentTime, info.Speed, info.Bitrate)
		},
	}

	startHLS := time.Now()
	usageHLS, err := mosaic.EncodeHls(
		context.Background(),
		jobHLS,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithThreads(4),
	)
	fmt.Println()
	if err != nil {
		log.Fatalf("HLS Live encoding failed: %v", err)
	}
	if usageHLS != nil {
		fmt.Printf("HLS usage: user=%.2fs system=%.2fs maxrss=%d KB\n", usageHLS.UserTime, usageHLS.SystemTime, usageHLS.MaxMemory)
	}
	fmt.Printf("HLS completed in %s -> %s\n\n", time.Since(startHLS).Round(time.Second), outputDirHLS)

	// 2. Low-Latency DASH Live Profile
	fmt.Println("--- 2. Starting Low-Latency DASH Live Encoding ---")
	jobDASH := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDirDASH,
		Profile:   mosaic.ProfileLive,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[DASH LIVE] [%5.1f%%] time=%s speed=%s bitrate=%s", info.Percentage, info.CurrentTime, info.Speed, info.Bitrate)
		},
	}

	startDASH := time.Now()
	usageDASH, err := mosaic.EncodeDash(
		context.Background(),
		jobDASH,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithThreads(4),
	)
	fmt.Println()
	if err != nil {
		log.Fatalf("DASH Live encoding failed: %v", err)
	}
	if usageDASH != nil {
		fmt.Printf("DASH usage: user=%.2fs system=%.2fs maxrss=%d KB\n", usageDASH.UserTime, usageDASH.SystemTime, usageDASH.MaxMemory)
	}
	fmt.Printf("DASH completed in %s -> %s\n", time.Since(startDASH).Round(time.Second), outputDirDASH)
}
