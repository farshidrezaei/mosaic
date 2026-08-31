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
	if err := ensureOutputDir(outDir); err != nil {
		return nil, err
	}

	filter := buildFilterGraph(l, info, opts.Watermark)
	gop := calcGOP(info.FPS, profile.SegmentDuration)

	args := []string{
		"-y",
		"-loglevel", opts.LogLevel,

		"-analyzeduration", "100M",
		"-probesize", "100M",
		"-fflags", "+genpts",

		"-i", input,
	}

	if opts.Watermark != nil && opts.Watermark.Path != "" {
		args = append(args, "-i", opts.Watermark.Path)
	}

	if opts.Threads > 0 {
		args = append(args, "-threads", strconv.Itoa(opts.Threads))
	}

	args = append(args, "-filter_complex", filter)

	// ---------- VIDEO ----------
	encoderName := ResolveVideoEncoder(opts.Codec, opts.GPU)
	for i, r := range l {
		args = append(args,
			"-map", fmt.Sprintf("[v%do]", i),

			fmt.Sprintf("-c:v:%d", i), encoderName,
			"-pix_fmt", "yuv420p",
		)

		if opts.Codec == "" || opts.Codec == config.CodecH264 {
			args = append(args,
				fmt.Sprintf("-profile:v:%d", i), r.Profile,
				fmt.Sprintf("-level:v:%d", i), r.Level,
			)
			if encoderName == "libx264" {
				args = append(args, "-preset", "medium")
			}
		} else if opts.Codec == config.CodecAV1 && encoderName == "libsvtav1" {
			args = append(args, "-preset", "8")
		} else if opts.Codec == config.CodecHEVC && encoderName == "libx265" {
			args = append(args, "-x265-params", "no-info=1")
		}

		if opts.CRF > 0 {
			args = append(args, fmt.Sprintf("-crf:v:%d", i), strconv.Itoa(opts.CRF))
		}

		args = append(args,
			"-g", strconv.Itoa(gop),
			"-keyint_min", strconv.Itoa(gop),
			"-sc_threshold", "0",
			"-bf", fmt.Sprintf("%d", r.BFrames),

			fmt.Sprintf("-maxrate:v:%d", i), fmt.Sprintf("%dk", r.MaxRate),
			fmt.Sprintf("-bufsize:v:%d", i), fmt.Sprintf("%dk", r.BufSize),
		)
		args = appendVideoRotationMetadataReset(args, i)
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
			if opts.NormalizeAudio {
				args = append(args, fmt.Sprintf("-af:a:%d", i), "loudnorm=I=-16:TP=-1.5:LRA=11")
			}
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

		StreamProgress(progressChan, progressHandler)

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
