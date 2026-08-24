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

	inputPath := "https://assets-ae.tupic.com/2026/08/02/6a6efa1d285c6.mp4"
	//inputPath := "https://api.tupic.com/assets/v1/assets/proxy?path=https%3A%2F%2Fstorage.tupic.com%2F2026%2F08%2F24%2F01a034de-b62d-7865-8c30-0cf805d5b002%2Fprivate%2Fhls%2Fuploaded%2Findex.m3u8%3FExpires%3D1787699912%26Signature%3DE2RsxdM8ZI-C8Arl4ueUKSkCrPYqBfU2YXEywL~5n~~WQHL0QCZcqPmAts3OmD9p1YMRVeg2Ac67X8N-jN4vtMDLd-JFO-GiA6s2VgwS~VddQYVqhqubarV9Bp4ucHl5Jm5XBWyYT1krtmVeFYcwCrk7oIAhsSF~Fc-D11aeJTXf~SU-K4HtnoJZdSX7QoqqH2uKlWNE6dSCOTGU0TjGzTxhV2ksvHhCnqfI6xCOnd2jftI9aKJiLPW4Y6611~MhpS8PmUssLFB4VYV1qnFnP91eJV7tlQgnEORgRi02NTJvmtvFTn2UknHWFxQS~tcKBitJWzYu5va7PNiVFaqPUw__%26Key-Pair-Id%3DK1R1M45CWKN77T"
	outputDir := filepath.Join(cwd, "./output", "hls_simple")

	if mkdirErr := os.MkdirAll(outputDir, 0o755); mkdirErr != nil {
		log.Fatalf("failed to create output directory: %v", mkdirErr)
	}

	job := mosaic.Job{
		Input:     inputPath,
		OutputDir: outputDir,
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[%5.1f%%] time=%-11s speed=%-6s bitrate=%-12s", info.Percentage, info.CurrentTime, info.Speed, info.Bitrate)
		},
	}

	fmt.Printf("starting HLS encoding: %s\n", inputPath)
	start := time.Now()

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
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
