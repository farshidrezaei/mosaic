package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	input := "input.mp4"
	if len(os.Args) > 1 {
		input = os.Args[1]
	}

	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "my-stream-bucket"
	}

	job := mosaic.Job{
		Input:     input,
		OutputDir: "./output/hls_s3",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\rPackaging: [%5.1f%%] speed=%s",
				info.Percentage, info.Speed)
		},
	}

	fmt.Printf("Packaging %s and uploading to S3 bucket '%s'...\n", input, bucket)
	_, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithThumbnails(),
		mosaic.WithS3Upload(mosaic.S3Config{
			Bucket:          bucket,
			Endpoint:        os.Getenv("S3_ENDPOINT"),
			Region:          "us-east-1",
			KeyPrefix:       "vod/video-1",
			AccessKey:       os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretKey:       os.Getenv("AWS_SECRET_ACCESS_KEY"),
			ConcurrentFiles: 6,
		}),
	)
	if err != nil {
		log.Fatalf("Packaging or upload failed: %v", err)
	}

	fmt.Printf("\nStream packaged and uploaded to S3 successfully!\n")
}
