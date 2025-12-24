package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	// 1. Define input and output
	cwd, _ := os.Getwd()
	inputPath := filepath.Join(cwd, "input.mp4")
	outputDir := filepath.Join(cwd, "output", "hls")

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// 2. Create the job
	job := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDir,
		Profile:   mosaic.ProfileVOD, // Use ProfileLive for lower latency
	}

	log.Printf("Starting HLS encoding for %s...", inputPath)

	// 3. Run the encoding
	if err := mosaic.EncodeHls(context.Background(), job); err != nil {
		log.Fatalf("Encoding failed: %v", err)
	}

	log.Println("✅ Encoding completed successfully!")
	log.Printf("📂 Output available at: %s", outputDir)
}
