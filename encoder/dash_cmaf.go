package encoder

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/farshidrezaei/mosaic/config"
	"github.com/farshidrezaei/mosaic/internal/executor"
	"github.com/farshidrezaei/mosaic/ladder"
	"github.com/farshidrezaei/mosaic/probe"
)

// EncodeDASHCMAF encodes the input video to DASH with CMAF segments.
// It uses the default command executor and default options.
func EncodeDASHCMAF(
	ctx context.Context,
	input string,
	outDir string,
	info probe.VideoInfo,
	profile config.Profile,
	l []ladder.Rendition,
) (*executor.Usage, error) {
	return EncodeDASHCMAFWithExecutor(ctx, input, outDir, info, profile, l, executor.DefaultExecutor, nil, EncoderOptions{LogLevel: "warning"})
}

// EncodeDASHCMAFWithExecutor encodes the input video to DASH with CMAF segments using the provided executor.
// It generates a DASH-compliant manifest (.mpd) and fragmented MP4 segments.
func EncodeDASHCMAFWithExecutor(
	ctx context.Context,
	input string,
	outDir string,
	info probe.VideoInfo,
	profile config.Profile,
	l []ladder.Rendition,
	exec executor.CommandExecutor,
	progressHandler func(map[string]string),
	opts EncoderOptions,
) (*executor.Usage, error) {

	gop := calcGOP(info.FPS, profile.SegmentDuration)

	args := []string{
		"-y",
		"-loglevel", opts.LogLevel,

		"-analyzeduration", "100M",
		"-probesize", "100M",
		"-fflags", "+genpts",

		"-i", input,
	}

	if opts.Threads > 0 {
		args = append(args, "-threads", strconv.Itoa(opts.Threads))
	}

	// ---------- VIDEO ----------
	for i, r := range l {
		codec := "libx264"
		switch opts.GPU {
		case config.GPU_NVENC:
			codec = "h264_nvenc"
		case config.GPU_VAAPI:
			codec = "h264_vaapi"
		case config.GPU_VIDEOTOOLBOX:
			codec = "h264_videotoolbox"
		}

		args = append(args,
			"-map", "0:v:0",

			fmt.Sprintf("-c:v:%d", i), codec,
			fmt.Sprintf("-profile:v:%d", i), r.Profile,
			fmt.Sprintf("-level:v:%d", i), r.Level,

			"-pix_fmt", "yuv420p",
			"-preset", "medium",

			"-g", strconv.Itoa(gop),
			"-keyint_min", strconv.Itoa(gop),
			"-sc_threshold", "0",

			fmt.Sprintf("-maxrate:v:%d", i), fmt.Sprintf("%dk", r.MaxRate),
			fmt.Sprintf("-bufsize:v:%d", i), fmt.Sprintf("%dk", r.BufSize),

			fmt.Sprintf("-s:v:%d", i), fmt.Sprintf("%dx%d", r.Width, r.Height),
		)
	}

	// ---------- AUDIO ----------
	if info.HasAudio {
		for i := range l {
			args = append(args,
				"-map", "0:a:0",
				fmt.Sprintf("-c:a:%d", i), "aac",
				fmt.Sprintf("-b:a:%d", i), "96k",
				"-ac", "2",
			)
		}
	}

	// ---------- DASH ----------
	args = append(args,
		"-f", "dash",
		"-seg_duration", strconv.Itoa(profile.SegmentDuration),

		"-use_template", "1",
		"-use_timeline", "1",

		"-init_seg_name", "init-stream$RepresentationID$.m4s",
		"-media_seg_name", "chunk-stream$RepresentationID$-$Number$.m4s",

		"-adaptation_sets", func() string {
			if info.HasAudio {
				return "id=0,streams=v id=1,streams=a"
			}
			return "id=0,streams=v"
		}(),

		filepath.Join(outDir, "manifest.mpd"),
	)

	if progressHandler != nil {
		args = append(args, "-progress", "pipe:1")
		progressChan := make(chan string)
		errChan := make(chan error, 1)
		var usage *executor.Usage

		go func() {
			var err error
			_, usage, err = exec.ExecuteWithProgress(ctx, progressChan, "ffmpeg", args...)
			errChan <- err
		}()

		for raw := range progressChan {
			progressHandler(ParseProgress(raw))
		}

		if err := <-errChan; err != nil {
			return nil, fmt.Errorf("ffmpeg DASH failed: %w", err)
		}
		return usage, nil
	} else {
		_, usage, err := exec.Execute(ctx, "ffmpeg", args...)
		if err != nil {
			return nil, fmt.Errorf("ffmpeg DASH failed: %w", err)
		}
		return usage, nil
	}

}
