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
)

// EncodeHLSCMAF encodes the input video to HLS with CMAF segments.
// It uses the default command executor.
func EncodeHLSCMAF(
	ctx context.Context,
	input string,
	outDir string,
	info probe.VideoInfo,
	profile config.Profile,
	l []ladder.Rendition,
) error {
	return EncodeHLSCMAFWithExecutor(ctx, input, outDir, info, profile, l, executor.DefaultExecutor)
}

// EncodeHLSCMAFWithExecutor encodes the input video to HLS with CMAF segments using the provided executor.
func EncodeHLSCMAFWithExecutor(
	ctx context.Context,
	input string,
	outDir string,
	info probe.VideoInfo,
	profile config.Profile,
	l []ladder.Rendition,
	exec executor.CommandExecutor,
) error {

	filter := buildFilterGraph(l)
	gop := calcGOP(info.FPS, profile.SegmentDuration)

	args := []string{
		"-y",
		"-loglevel", "warning",

		// input safety
		"-analyzeduration", "100M",
		"-probesize", "100M",
		"-fflags", "+genpts",

		"-i", input,
		"-filter_complex", filter,
	}

	// ---------- VIDEO ----------
	for i, r := range l {
		args = append(args,
			"-map", fmt.Sprintf("[v%do]", i),

			fmt.Sprintf("-c:v:%d", i), "libx264",
			fmt.Sprintf("-profile:v:%d", i), r.Profile,
			fmt.Sprintf("-level:v:%d", i), r.Level,

			"-pix_fmt", "yuv420p",
			"-preset", "medium",

			"-g", strconv.Itoa(gop),
			"-keyint_min", strconv.Itoa(gop),
			"-sc_threshold", "0",

			fmt.Sprintf("-maxrate:v:%d", i), fmt.Sprintf("%dk", r.MaxRate),
			fmt.Sprintf("-bufsize:v:%d", i), fmt.Sprintf("%dk", r.BufSize),
		)
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
		}
	}

	// ---------- HLS / CMAF ----------
	args = append(args,
		"-f", "hls",
		"-hls_segment_type", "fmp4",
		"-hls_playlist_type", "vod",
	)

	if profile.LowLatency {
		args = append(args,
			"-hls_time", strconv.Itoa(profile.SegmentDuration),
			"-hls_part_size", "0.5",
			"-hls_flags", "independent_segments+split_by_time",
		)
	} else {
		args = append(args,
			"-hls_time", strconv.Itoa(profile.SegmentDuration),
			"-hls_flags", "independent_segments",
		)
	}

	args = append(args,
		"-hls_segment_filename",
		filepath.Join(outDir, "seg_%v_%d.m4s"),

		"-master_pl_name", "master.m3u8",
		"-var_stream_map", buildVarStreamMap(len(l), info.HasAudio),

		filepath.Join(outDir, "stream_%v.m3u8"),
	)

	_, err := exec.Execute(ctx, "ffmpeg", args...)
	if err != nil {
		return fmt.Errorf("ffmpeg HLS failed: %w", err)
	}

	return nil
}

// ---------- FILTER GRAPH ----------

func buildFilterGraph(l []ladder.Rendition) string {
	var b strings.Builder

	// split
	b.WriteString("[0:v]")
	b.WriteString(fmt.Sprintf("split=%d", len(l)))
	for i := range l {
		b.WriteString(fmt.Sprintf("[v%d]", i))
	}
	b.WriteString(";")

	// scale + pad + SAR
	for i, r := range l {
		b.WriteString(fmt.Sprintf(
			"[v%d]scale=%d:%d:force_original_aspect_ratio=decrease,"+
				"pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1[v%do];",
			i,
			r.Width, r.Height,
			r.Width, r.Height,
			i,
		))
	}

	return strings.TrimSuffix(b.String(), ";")
}
