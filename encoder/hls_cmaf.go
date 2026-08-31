package encoder

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/farshidrezaei/mosaic/config"
	"github.com/farshidrezaei/mosaic/internal/executor"
	"github.com/farshidrezaei/mosaic/ladder"
	"github.com/farshidrezaei/mosaic/probe"
	"github.com/farshidrezaei/mosaic/watermark"
)

// EncoderOptions defines options for the encoder.
type EncoderOptions struct {
	Watermark      *watermark.Config
	KeyInfoFile    string
	GPU            config.GPUType
	Codec          config.VideoCodec
	LogLevel       string
	Threads        int
	CRF            int
	EnableIFrames  bool
	NormalizeAudio bool
}

// EncodeHLSCMAF encodes the input video to HLS with CMAF segments.
// It uses the default command executor and default options.
func EncodeHLSCMAF(
	ctx context.Context,
	input string,
	outDir string,
	info probe.VideoInfo,
	profile config.Profile,
	l []ladder.Rendition,
) (*executor.Usage, error) {
	return EncodeHLSCMAFWithExecutor(ctx, input, outDir, info, profile, l, executor.DefaultExecutor, nil, EncoderOptions{LogLevel: "warning"})
}

// EncodeHLSCMAFWithExecutor encodes the input video to HLS with CMAF segments using the provided executor.
// It constructs a complex FFmpeg command to generate multiple renditions in a single pass.
func EncodeHLSCMAFWithExecutor(
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

		// input safety
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
				"-map", "a:0",
				fmt.Sprintf("-c:a:%d", i), "aac",
				fmt.Sprintf("-b:a:%d", i), "96k",
				"-ac", "2",
			)
			if opts.NormalizeAudio {
				args = append(args, fmt.Sprintf("-af:a:%d", i), "loudnorm=I=-16:TP=-1.5:LRA=11")
			}
		}
	}

	// ---------- HLS / CMAF ----------
	args = append(args, "-f", "hls")
	segmentFilename := filepath.Join(outDir, "seg_%v_%d.m4s")

	if opts.KeyInfoFile != "" {
		args = append(args, "-hls_key_info_file", opts.KeyInfoFile)
		segmentFilename = filepath.Join(outDir, "seg_%v_%d.ts")
	} else {
		args = append(args, "-hls_segment_type", "fmp4")
	}

	if !profile.LowLatency {
		args = append(args, "-hls_playlist_type", "vod")
	}

	hlsFlags := "independent_segments"
	if profile.LowLatency {
		hlsFlags += "+split_by_time"
	}
	if opts.EnableIFrames {
		hlsFlags += "+iframes_only"
	}

	args = append(args,
		"-hls_time", strconv.Itoa(profile.SegmentDuration),
		"-hls_flags", hlsFlags,
		"-hls_segment_filename", segmentFilename,
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", buildVarStreamMap(len(l), info.HasAudio),
		filepath.Join(outDir, "stream_%v.m3u8"),
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
			return nil, fmt.Errorf("ffmpeg HLS failed: %w", err)
		}
		return usage, nil
	} else {
		_, usage, err := exec.Execute(ctx, "ffmpeg", args...)
		if err != nil {
			return nil, fmt.Errorf("ffmpeg HLS failed: %w", err)
		}
		return usage, nil
	}

}

// ---------- FILTER GRAPH ----------

func buildFilterGraph(l []ladder.Rendition, info probe.VideoInfo, wm *watermark.Config) string {
	var b strings.Builder

	// split
	b.WriteString("[0:v]")
	_, _ = fmt.Fprintf(&b, "split=%d", len(l))
	for i := range l {
		_, _ = fmt.Fprintf(&b, "[v%d]", i)
	}
	b.WriteString(";")

	dw := info.DisplayWidth()
	dh := info.DisplayHeight()

	if wm != nil && wm.Path != "" {
		wm.Normalize()
		b.WriteString("[1:v]")
		_, _ = fmt.Fprintf(&b, "split=%d", len(l))
		for i := range l {
			_, _ = fmt.Fprintf(&b, "[wm%d]", i)
		}
		b.WriteString(";")

		for i, r := range l {
			wmWidth := int(float64(r.Width) * wm.ScaleFraction)
			if wmWidth%2 != 0 {
				wmWidth++
			}
			if wmWidth < 16 {
				wmWidth = 16
			}

			if wm.Opacity < 1.0 {
				_, _ = fmt.Fprintf(&b,
					"[wm%d]scale=%d:-1,format=rgba,colorchannelmixer=aa=%.2f[wm%d_proc];",
					i, wmWidth, wm.Opacity, i,
				)
			} else {
				_, _ = fmt.Fprintf(&b,
					"[wm%d]scale=%d:-1[wm%d_proc];",
					i, wmWidth, i,
				)
			}

			if dw > 0 && dh > 0 {
				_, _ = fmt.Fprintf(&b,
					"[v%d]scale=%d:%d,setsar=1,setdar=%d/%d[v%d_scaled];",
					i, r.Width, r.Height, dw, dh, i,
				)
			} else {
				_, _ = fmt.Fprintf(&b,
					"[v%d]scale=%d:%d,setsar=1[v%d_scaled];",
					i, r.Width, r.Height, i,
				)
			}

			_, _ = fmt.Fprintf(&b,
				"[v%d_scaled][wm%d_proc]overlay=%s[v%do];",
				i, i, wm.OverlayExpression(), i,
			)
		}
	} else {
		for i, r := range l {
			if dw > 0 && dh > 0 {
				_, _ = fmt.Fprintf(&b,
					"[v%d]scale=%d:%d,setsar=1,setdar=%d/%d[v%do];",
					i, r.Width, r.Height, dw, dh, i,
				)
			} else {
				_, _ = fmt.Fprintf(&b,
					"[v%d]scale=%d:%d,setsar=1[v%do];",
					i, r.Width, r.Height, i,
				)
			}
		}
	}

	return strings.TrimSuffix(b.String(), ";")
}
