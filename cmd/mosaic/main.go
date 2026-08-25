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
)

const version = "1.7.2"

func printUsage() {
	fmt.Fprintf(os.Stderr, `Mosaic CLI - Predictable Adaptive Bitrate (ABR) Video Packaging (v%s)

Usage:
  mosaic [options] -i <input-path-or-url> -o <output-dir>

Options:
  -i, --input string        Input video path or remote URL (required)
  -o, --output string       Output directory for generated stream files (required)
  -f, --format string       Packaging format: "hls" or "dash" (default "hls")
  -p, --profile string      Encoding profile: "vod" (5s segments) or "live" (2s segments) (default "vod")
      --normalize           Auto-probe and physically normalize mobile/rotated video (default true)
      --threads int         FFmpeg encoding CPU threads (default 0 = auto)
      --bframes int         Number of consecutive B-frames for non-baseline streams (default 0)
      --fps-scale           Scale bitrate caps upward for high-framerate (>30 FPS) videos
      --gpu string          Hardware acceleration backend: "nvenc", "vaapi", "videotoolbox"
      --log-level string    FFmpeg log level: "quiet", "warning", "error", "info" (default "warning")
  -v, --version             Display Mosaic version and exit
  -h, --help                Display this help message and exit

Examples:
  # Package an MP4 or remote URL into HLS fMP4 with orientation normalization:
  mosaic -i input.mp4 -o ./output/hls

  # Package into DASH CMAF with 4 threads and NVENC GPU acceleration:
  mosaic -i input.mp4 -o ./output/dash -f dash --threads 4 --gpu nvenc

Documentation:
  https://farshidrezaei.github.io/mosaic/
`, version)
}

func main() {
	var (
		input       string
		outputDir   string
		format      string
		profileStr  string
		normalize   bool
		threads     int
		bframes     int
		fpsScale    bool
		gpuType     string
		logLevel    string
		showVersion bool
		showHelp    bool
	)

	flag.StringVar(&input, "i", "", "Input video path or URL")
	flag.StringVar(&input, "input", "", "Input video path or URL")
	flag.StringVar(&outputDir, "o", "", "Output directory")
	flag.StringVar(&outputDir, "output", "", "Output directory")
	flag.StringVar(&format, "f", "hls", "Packaging format (hls, dash)")
	flag.StringVar(&format, "format", "hls", "Packaging format (hls, dash)")
	flag.StringVar(&profileStr, "p", "vod", "Profile (vod, live)")
	flag.StringVar(&profileStr, "profile", "vod", "Profile (vod, live)")
	flag.BoolVar(&normalize, "normalize", true, "Normalize mobile orientation")
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

	switch strings.ToLower(gpuType) {
	case "nvenc":
		opts = append(opts, mosaic.WithNVENC())
	case "vaapi":
		opts = append(opts, mosaic.WithVAAPI())
	case "videotoolbox":
		opts = append(opts, mosaic.WithVideoToolbox())
	}

	fmt.Printf("⚡ Starting %s encoding (%s profile)\n", strings.ToUpper(format), profileStr)
	fmt.Printf("   Input:  %s\n", input)
	fmt.Printf("   Output: %s\n\n", outputDir)

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
}
