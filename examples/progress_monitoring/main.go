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

	inputPath := "https://assets-ae.tupic.com/2026/08/02/6a6efa1d285c6.mp4"
	//inputPath := "https://api.tupic.com/assets/v1/assets/proxy?path=https%3A%2F%2Fstorage.tupic.com%2F2026%2F08%2F24%2F01a034de-b62d-7865-8c30-0cf805d5b002%2Fprivate%2Fhls%2Fuploaded%2Findex.m3u8%3FExpires%3D1787699912%26Signature%3DE2RsxdM8ZI-C8Arl4ueUKSkCrPYqBfU2YXEywL~5n~~WQHL0QCZcqPmAts3OmD9p1YMRVeg2Ac67X8N-jN4vtMDLd-JFO-GiA6s2VgwS~VddQYVqhqubarV9Bp4ucHl5Jm5XBWyYT1krtmVeFYcwCrk7oIAhsSF~Fc-D11aeJTXf~SU-K4HtnoJZdSX7QoqqH2uKlWNE6dSCOTGU0TjGzTxhV2ksvHhCnqfI6xCOnd2jftI9aKJiLPW4Y6611~MhpS8PmUssLFB4VYV1qnFnP91eJV7tlQgnEORgRi02NTJvmtvFTn2UknHWFxQS~tcKBitJWzYu5va7PNiVFaqPUw__%26Key-Pair-Id%3DK1R1M45CWKN77T"
	outputDir := filepath.Join(cwd, "./output", "progress_monitoring")

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
