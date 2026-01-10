package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	// This example demonstrates how to use different hardware acceleration backends.
	// Note: You must have the corresponding hardware and FFmpeg support installed.

	cwd, _ := os.Getwd()
	inputPath := filepath.Join(cwd, "input.mp4")
	outputDir := filepath.Join(cwd, "output", "multi_gpu")

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		log.Printf("⚠️  Warning: %s not found.", inputPath)
		return
	}

	_ = os.MkdirAll(outputDir, 0755)

	job := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDir,
		Profile:   mosaic.ProfileLive, // Live profile for low-latency
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\rProgress: %.1f%% (Speed: %s)", info.Percentage, info.Speed)
		},
	}

	fmt.Println("🚀 Starting Multi-GPU Hardware Acceleration Example")

	// Example 1: NVIDIA NVENC
	fmt.Println("\n--- Using NVIDIA NVENC ---")
	_, err := mosaic.EncodeHls(context.Background(), job, mosaic.WithNVENC())
	if err != nil {
		fmt.Printf("NVENC failed (likely no hardware): %v\n", err)
	}

	// Example 2: Intel/AMD VAAPI
	fmt.Println("\n--- Using VAAPI ---")
	_, err = mosaic.EncodeHls(context.Background(), job, mosaic.WithVAAPI())
	if err != nil {
		fmt.Printf("VAAPI failed (likely no hardware): %v\n", err)
	}

	// Example 3: Apple VideoToolbox
	fmt.Println("\n--- Using VideoToolbox ---")
	_, err = mosaic.EncodeHls(context.Background(), job, mosaic.WithVideoToolbox())
	if err != nil {
		fmt.Printf("VideoToolbox failed (likely no hardware): %v\n", err)
	}

	fmt.Println("\n\n✅ Multi-GPU example finished.")
}
