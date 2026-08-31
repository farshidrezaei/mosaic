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

	job := mosaic.Job{
		Input:     input,
		OutputDir: "./output/hls_encrypted",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\rEncrypting & Packaging: [%5.1f%%] speed=%s",
				info.Percentage, info.Speed)
		},
	}

	_, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithAES128Encryption(mosaic.EncryptionConfig{
			KeyURI: "enc.key", // In production: "https://api.yourdomain.com/keys/session.key"
		}),
		mosaic.WithThumbnails(),
	)
	if err != nil {
		log.Fatalf("Packaging failed: %v", err)
	}

	fmt.Printf("\nEncrypted HLS stream generated successfully in ./output/hls_encrypted\n")
	fmt.Printf("Play with: mosaic preview ./output/hls_encrypted\n")
}
