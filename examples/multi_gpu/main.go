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
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v", err)
	}

	inputPath := filepath.Join(cwd, "input.mp4")
	baseOutputDir := filepath.Join(cwd, "output", "multi_gpu")

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		log.Printf("input file not found: %s", inputPath)
		log.Printf("place a video file named input.mp4 in %s", cwd)
		return
	}

	if err := os.MkdirAll(baseOutputDir, 0o755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	backends := []struct {
		opt  mosaic.Option
		name string
		dir  string
	}{
		{name: "NVENC", dir: "nvenc", opt: mosaic.WithNVENC()},
		{name: "VAAPI", dir: "vaapi", opt: mosaic.WithVAAPI()},
		{name: "VideoToolbox", dir: "videotoolbox", opt: mosaic.WithVideoToolbox()},
	}

	for _, b := range backends {
		outDir := filepath.Join(baseOutputDir, b.dir)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			fmt.Printf("%s: skip, cannot create output dir: %v\n", b.name, err)
			continue
		}

		job := mosaic.Job{
			Input:     inputPath,
			OutputDir: outDir,
			Profile:   mosaic.ProfileLive,
			ProgressHandler: func(info mosaic.ProgressInfo) {
				fmt.Printf("\r[%s] time=%s speed=%s", b.name, info.CurrentTime, info.Speed)
			},
		}

		fmt.Printf("\n--- %s ---\n", b.name)
		usage, err := mosaic.EncodeHls(
			context.Background(),
			job,
			mosaic.WithNormalizeOrientation(),
			b.opt,
			mosaic.WithLogLevel("warning"),
		)
		fmt.Println()
		if err != nil {
			fmt.Printf("%s failed: %v\n", b.name, err)
			continue
		}

		if usage != nil {
			fmt.Printf("%s usage: user=%.2fs system=%.2fs maxrss=%d\n", b.name, usage.UserTime, usage.SystemTime, usage.MaxMemory)
		}
		fmt.Printf("%s output: %s\n", b.name, outDir)
	}
}
