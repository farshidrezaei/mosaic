package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/farshidrezaei/mosaic"
	"github.com/farshidrezaei/mosaic/preview"
)

var version = "1.8.0"

func printUsage() {
	fmt.Fprintf(os.Stderr, `Mosaic CLI - Predictable Adaptive Bitrate (ABR) Video Packaging (v%s)

Usage:
  mosaic [options] -i <input-path-or-url> -o <output-dir>
  mosaic preview [options] [output-dir]

Commands:
  preview                   Start local web preview player to test HLS/DASH streams

Options:
  -i, --input string        Input video path or remote URL (required)
  -o, --output string       Output directory for generated stream files (required)
  -f, --format string       Packaging format: "hls" or "dash" (default "hls")
  -p, --profile string      Encoding profile: "vod" (5s segments) or "live" (2s segments) (default "vod")
      --codec string        Video codec: "h264", "hevc" (H.265), "av1" (default "h264")
      --crf int             Constant Rate Factor for quality optimization (e.g. 23 for H264/HEVC, 28 for AV1)
      --normalize           Auto-probe and physically normalize mobile/rotated video (default true)
      --normalize-audio     Normalize audio volume using EBU R128 standard (loudnorm)
      --watermark string    Path to watermark image overlay (PNG/WebP)
      --encrypt-aes128      Encrypt HLS segments with AES-128 standard
      --thumbnails          Generate storyboard sprite sheet and WebVTT cue file
      --iframes             Generate I-frame-only trick-play playlists for HLS
      --threads int         FFmpeg encoding CPU threads (default 0 = auto)
      --bframes int         Number of consecutive B-frames for non-baseline streams (default 0)
      --fps-scale           Scale bitrate caps upward for high-framerate (>30 FPS) videos
      --gpu string          Hardware acceleration backend: "nvenc", "vaapi", "videotoolbox"
      --log-level string    FFmpeg log level: "quiet", "warning", "error", "info" (default "warning")
  -v, --version             Display Mosaic version and exit
  -h, --help                Display this help message and exit

Examples:
  # Package an MP4 or remote URL into HLS fMP4 with orientation normalization and thumbnails:
  mosaic -i input.mp4 -o ./output/hls --thumbnails

  # Package into DASH CMAF with 4 threads and NVENC GPU acceleration:
  mosaic -i input.mp4 -o ./output/dash -f dash --threads 4 --gpu nvenc

  # Start local web preview player to inspect generated streams:
  mosaic preview ./output/hls

Documentation:
  https://farshidrezaei.github.io/mosaic/
`, version)
}

func handlePreview(args []string) {
	previewFlags := flag.NewFlagSet("preview", flag.ExitOnError)
	var port int
	previewFlags.IntVar(&port, "p", 8080, "Preview server port")
	previewFlags.IntVar(&port, "port", 8080, "Preview server port")

	previewFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  mosaic preview [options] [stream-directory]

Options:
  -p, --port int    Port for local HTTP preview server (default 8080)
`)
	}

	_ = previewFlags.Parse(args)

	dir := "."
	if previewFlags.NArg() > 0 {
		dir = previewFlags.Arg(0)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := preview.Serve(ctx, dir, port); err != nil {
		fmt.Fprintf(os.Stderr, "Preview server error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "preview" {
		handlePreview(os.Args[2:])
		return
	}

	var (
		input          string
		outputDir      string
		format         string
		profileStr     string
		codecStr       string
		crf            int
		normalize      bool
		normalizeAudio bool
		watermarkPath  string
		encryptAES128  bool
		thumbnails     bool
		iframes        bool
		s3Bucket       string
		s3Endpoint     string
		s3Region       string
		s3Prefix       string
		s3AccessKey    string
		s3SecretKey    string
		threads        int
		bframes        int
		fpsScale       bool
		gpuType        string
		logLevel       string
		showVersion    bool
		showHelp       bool
	)

	flag.StringVar(&input, "i", "", "Input video path or URL")
	flag.StringVar(&input, "input", "", "Input video path or URL")
	flag.StringVar(&outputDir, "o", "", "Output directory")
	flag.StringVar(&outputDir, "output", "", "Output directory")
	flag.StringVar(&format, "f", "hls", "Packaging format (hls, dash)")
	flag.StringVar(&format, "format", "hls", "Packaging format (hls, dash)")
	flag.StringVar(&profileStr, "p", "vod", "Profile (vod, live)")
	flag.StringVar(&profileStr, "profile", "vod", "Profile (vod, live)")
	flag.StringVar(&codecStr, "codec", "h264", "Video codec (h264, hevc, av1)")
	flag.IntVar(&crf, "crf", 0, "Constant Rate Factor (e.g. 23 for H264/HEVC, 28 for AV1)")
	flag.BoolVar(&normalize, "normalize", true, "Normalize mobile orientation")
	flag.BoolVar(&normalizeAudio, "normalize-audio", false, "Normalize audio volume using EBU R128 standard (loudnorm)")
	flag.StringVar(&watermarkPath, "watermark", "", "Path to watermark image overlay (PNG/WebP)")
	flag.BoolVar(&encryptAES128, "encrypt-aes128", false, "Encrypt HLS segments with AES-128 standard")
	flag.BoolVar(&thumbnails, "thumbnails", false, "Generate storyboard sprite sheet and WebVTT cues")
	flag.BoolVar(&iframes, "iframes", false, "Generate I-frame only playlists for HLS")
	flag.StringVar(&s3Bucket, "s3-bucket", "", "S3 bucket name to auto-upload stream files")
	flag.StringVar(&s3Endpoint, "s3-endpoint", "", "S3/MinIO endpoint URL (e.g. https://s3.amazonaws.com)")
	flag.StringVar(&s3Region, "s3-region", "us-east-1", "S3 region")
	flag.StringVar(&s3Prefix, "s3-prefix", "", "S3 object key prefix")
	flag.StringVar(&s3AccessKey, "s3-access-key", "", "S3 access key (defaults to AWS_ACCESS_KEY_ID env)")
	flag.StringVar(&s3SecretKey, "s3-secret-key", "", "S3 secret key (defaults to AWS_SECRET_ACCESS_KEY env)")
	flag.IntVar(&threads, "threads", 0, "CPU worker threads")
	flag.IntVar(&bframes, "bframes", 0, "B-frame count")
	flag.BoolVar(&fpsScale, "fps-scale", false, "Scale bitrate for >30 FPS content")
	flag.StringVar(&gpuType, "gpu", "", "GPU type (nvenc, vaapi, videotoolbox)")
	flag.StringVar(&logLevel, "log-level", "warning", "FFmpeg log level")
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	flag.Usage = printUsage
	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("mosaic version %s\n", version)
		os.Exit(0)
	}

	if input == "" || outputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: both -i/--input and -o/--output are required.")
		fmt.Fprintln(os.Stderr, "Run 'mosaic --help' for usage instructions.")
		os.Exit(1)
	}

	profile := mosaic.ProfileVOD
	if strings.ToLower(profileStr) == "live" {
		profile = mosaic.ProfileLive
	}

	// Setup graceful interrupt handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Progress bar renderer
	progressHandler := func(info mosaic.ProgressInfo) {
		const barWidth = 25
		percent := info.Percentage
		if percent > 100 {
			percent = 100
		}
		filled := int((percent / 100.0) * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		unfilled := barWidth - filled

		bar := strings.Repeat("█", filled) + strings.Repeat("░", unfilled)
		timeStr := info.CurrentTime
		if timeStr == "" {
			timeStr = "N/A"
		}
		speedStr := info.Speed
		if speedStr == "" {
			speedStr = "N/A"
		}
		bitrateStr := info.Bitrate
		if bitrateStr == "" {
			bitrateStr = "N/A"
		}

		fmt.Printf("\r[%s] %5.1f%% | Time: %-11s | Speed: %-6s | Bitrate: %-12s",
			bar, percent, timeStr, speedStr, bitrateStr)
	}

	job := mosaic.Job{
		Input:           input,
		OutputDir:       outputDir,
		Profile:         profile,
		ProgressHandler: progressHandler,
	}

	var opts []mosaic.Option
	if normalize {
		opts = append(opts, mosaic.WithNormalizeOrientation())
	}
	if normalizeAudio {
		opts = append(opts, mosaic.WithNormalizeAudio())
	}
	if watermarkPath != "" {
		opts = append(opts, mosaic.WithWatermark(mosaic.WatermarkConfig{Path: watermarkPath}))
	}
	if thumbnails {
		opts = append(opts, mosaic.WithThumbnails())
	}
	if iframes {
		opts = append(opts, mosaic.WithIFrames())
	}
	if encryptAES128 {
		opts = append(opts, mosaic.WithAES128Encryption())
	}
	if threads > 0 {
		opts = append(opts, mosaic.WithThreads(threads))
	}
	if bframes > 0 {
		opts = append(opts, mosaic.WithBFrames(bframes))
	}
	if fpsScale {
		opts = append(opts, mosaic.WithScaleBitrateWithFPS())
	}
	if logLevel != "" {
		opts = append(opts, mosaic.WithLogLevel(logLevel))
	}

	switch strings.ToLower(codecStr) {
	case "hevc", "h265":
		opts = append(opts, mosaic.WithHEVC())
	case "av1":
		opts = append(opts, mosaic.WithAV1())
	default:
		opts = append(opts, mosaic.WithCodec(mosaic.CodecH264))
	}
	if crf > 0 {
		opts = append(opts, mosaic.WithCRF(crf))
	}

	if s3Bucket != "" {
		ak := s3AccessKey
		if ak == "" {
			ak = os.Getenv("AWS_ACCESS_KEY_ID")
		}
		sk := s3SecretKey
		if sk == "" {
			sk = os.Getenv("AWS_SECRET_ACCESS_KEY")
		}
		opts = append(opts, mosaic.WithS3Upload(mosaic.S3Config{
			Bucket:    s3Bucket,
			Endpoint:  s3Endpoint,
			Region:    s3Region,
			KeyPrefix: s3Prefix,
			AccessKey: ak,
			SecretKey: sk,
		}))
	}

	switch strings.ToLower(gpuType) {
	case "nvenc":
		opts = append(opts, mosaic.WithNVENC())
	case "vaapi":
		opts = append(opts, mosaic.WithVAAPI())
	case "videotoolbox":
		opts = append(opts, mosaic.WithVideoToolbox())
	}

	fmt.Printf("⚡ Starting %s encoding (%s profile)\n", strings.ToUpper(format), profileStr)
	fmt.Printf("   Input:      %s\n", input)
	fmt.Printf("   Output:     %s\n", outputDir)
	fmt.Printf("   Thumbnails: %t\n\n", thumbnails)

	startTime := time.Now()
	var usage *mosaic.Usage
	var err error

	if strings.ToLower(format) == "dash" {
		usage, err = mosaic.EncodeDash(ctx, job, opts...)
	} else {
		usage, err = mosaic.EncodeHls(ctx, job, opts...)
	}

	fmt.Println() // newline after progress bar
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Encoding failed: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(startTime).Round(time.Second)
	fmt.Printf("\n✨ Packaging completed in %s!\n", elapsed)
	if usage != nil {
		fmt.Printf("   Peak Memory: %.1f MB | CPU Time: %.2fs user, %.2fs sys\n",
			float64(usage.MaxMemory)/1024.0, usage.UserTime, usage.SystemTime)
	}
	fmt.Printf("   Output: %s\n", outputDir)
	fmt.Printf("   💡 Tip: Run 'mosaic preview %s' to test the stream in your browser!\n", outputDir)
}
