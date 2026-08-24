package mosaic

import (
	"context"
	"fmt"

	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/farshidrezaei/mosaic/config"
	"github.com/farshidrezaei/mosaic/encoder"
	"github.com/farshidrezaei/mosaic/internal/executor"
	"github.com/farshidrezaei/mosaic/ladder"
	"github.com/farshidrezaei/mosaic/optimize"
	"github.com/farshidrezaei/mosaic/probe"
)

// Option defines a functional option for configuring encoding jobs.
// It allows for flexible and extensible configuration of the encoding process.
type Option func(*options)

type options struct {
	logger               *slog.Logger
	gpu                  config.GPUType
	logLevel             string
	threads              int
	bframes              int
	normalizeOrientation bool
	scaleBitrateWithFPS  bool
}

func defaultOptions() *options {
	return &options{
		threads:  0, // auto
		gpu:      "",
		logLevel: "warning",
		logger:   slog.Default(),
		bframes:  0,
	}
}

// WithNormalizeOrientation enables pre-encoding orientation normalization.
// If called without arguments, it enables normalization.
func WithNormalizeOrientation(enabled ...bool) Option {
	return func(o *options) {
		if len(enabled) == 0 {
			o.normalizeOrientation = true
			return
		}
		o.normalizeOrientation = enabled[0]
	}
}

// WithThreads sets the number of CPU threads to use for encoding.
// Set to 0 (default) to let FFmpeg auto-detect the optimal number of threads.
func WithThreads(n int) Option {
	return func(o *options) {
		o.threads = n
	}
}

// WithBFrames sets the number of B-frames to use for encoding non-baseline renditions (default 0).
func WithBFrames(n int) Option {
	return func(o *options) {
		o.bframes = n
	}
}

// WithScaleBitrateWithFPS enables optional bitrate scaling for high framerate videos (>30 FPS).
func WithScaleBitrateWithFPS(enabled ...bool) Option {
	return func(o *options) {
		if len(enabled) == 0 {
			o.scaleBitrateWithFPS = true
			return
		}
		o.scaleBitrateWithFPS = enabled[0]
	}
}

// WithGPU enables hardware acceleration for the encoding process.
// If no specific GPUType is provided, it defaults to NVIDIA NVENC.
func WithGPU(t ...config.GPUType) Option {
	return func(o *options) {
		if len(t) > 0 {
			o.gpu = t[0]
		} else {
			o.gpu = config.GPU_NVENC
		}
	}
}

// WithNVENC enables NVIDIA NVENC hardware acceleration.
func WithNVENC() Option {
	return func(o *options) {
		o.gpu = config.GPU_NVENC
	}
}

// WithVAAPI enables VAAPI (Intel/AMD) hardware acceleration.
func WithVAAPI() Option {
	return func(o *options) {
		o.gpu = config.GPU_VAAPI
	}
}

// WithVideoToolbox enables Apple VideoToolbox hardware acceleration.
func WithVideoToolbox() Option {
	return func(o *options) {
		o.gpu = config.GPU_VIDEOTOOLBOX
	}
}

// WithLogLevel sets the FFmpeg log level (e.g., "quiet", "error", "warning", "info", "debug").
// The default is "warning".
func WithLogLevel(level string) Option {
	return func(o *options) {
		o.logLevel = level
	}
}

// WithLogger sets a custom slog.Logger for internal library logging.
// By default, it uses slog.Default().
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func initialize(ctx context.Context, job Job, opts *options) (probe.VideoInfo, config.Profile, []ladder.Rendition, error) {
	return initializeWithExecutor(ctx, job, executor.DefaultExecutor, opts)
}

func initializeWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts *options) (probe.VideoInfo, config.Profile, []ladder.Rendition, error) {
	// 1. Probe
	info, err := probe.InputWithExecutor(ctx, job.Input, exec)
	if err != nil {
		return probe.VideoInfo{}, config.Profile{}, []ladder.Rendition{}, err
	}

	// profile
	var profile config.Profile
	switch job.Profile {
	case ProfileLive:
		profile = config.LIVE
	default:
		profile = config.VOD
	}

	if opts.bframes > 0 {
		profile.BFrames = opts.bframes
	}

	// build ladder
	l := ladder.Build(info, profile.BFrames)

	// cost optimizer
	if opts.scaleBitrateWithFPS {
		l = optimize.Apply(l, optimize.WithFPS(info.FPS))
	} else {
		l = optimize.Apply(l)
	}

	opts.logger.Info("encoding variants", "count", len(l))

	return info, profile, l, err
}

func buildProgressCallback(handler ProgressHandler, duration float64) func(map[string]string) {
	if handler == nil {
		return nil
	}
	return func(m map[string]string) {
		var percentage float64
		if duration > 0 {
			currentSec := encoder.ParseOutTimeSeconds(m)
			percentage = (currentSec / duration) * 100.0
			if percentage > 100.0 {
				percentage = 100.0
			} else if percentage < 0.0 {
				percentage = 0.0
			}
		}
		handler(ProgressInfo{
			Percentage:  percentage,
			CurrentTime: m["out_time"],
			Bitrate:     m["bitrate"],
			Speed:       m["speed"],
		})
	}
}

// EncodeHls encodes the given job into HLS format with CMAF segments.
// It automatically builds an optimized encoding ladder and generates a master playlist.
// Functional options can be provided to customize the encoding process.
func EncodeHls(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error) {
	return EncodeHlsWithExecutor(ctx, job, executor.DefaultExecutor, opts...)
}

// EncodeHlsWithExecutor is like EncodeHls but allows providing a custom CommandExecutor.
// This is primarily used for testing or advanced command execution scenarios.
func EncodeHlsWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) (*executor.Usage, error) {
	if err := job.validate(); err != nil {
		return nil, err
	}

	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	effectiveInput, cleanupInput, err := prepareInputForEncoding(ctx, job.Input, exec, o)
	if err != nil {
		return nil, err
	}
	defer cleanupInput()

	effectiveJob := job
	effectiveJob.Input = effectiveInput

	info, profile, l, err := initializeWithExecutor(ctx, effectiveJob, exec, o)
	if err != nil {
		return nil, err
	}
	// 2. Encode
	return encoder.EncodeHLSCMAFWithExecutor(
		ctx,
		effectiveInput,
		job.OutputDir,
		info,
		profile,
		l,
		exec,
		buildProgressCallback(job.ProgressHandler, info.Duration),
		encoder.EncoderOptions{
			Threads:  o.threads,
			GPU:      o.gpu,
			LogLevel: o.logLevel,
		},
	)
}

// EncodeDash encodes the given job into DASH format with CMAF segments.
// It automatically builds an optimized encoding ladder and generates a DASH manifest (.mpd).
// Functional options can be provided to customize the encoding process.
func EncodeDash(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error) {
	return EncodeDashWithExecutor(ctx, job, executor.DefaultExecutor, opts...)
}

// EncodeDashWithExecutor is like EncodeDash but allows providing a custom CommandExecutor.
// This is primarily used for testing or advanced command execution scenarios.
func EncodeDashWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) (*executor.Usage, error) {
	if err := job.validate(); err != nil {
		return nil, err
	}

	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	effectiveInput, cleanupInput, err := prepareInputForEncoding(ctx, job.Input, exec, o)
	if err != nil {
		return nil, err
	}
	defer cleanupInput()

	effectiveJob := job
	effectiveJob.Input = effectiveInput

	info, profile, l, err := initializeWithExecutor(ctx, effectiveJob, exec, o)
	if err != nil {
		return nil, err
	}
	// 2. Encode
	return encoder.EncodeDASHCMAFWithExecutor(
		ctx,
		effectiveInput,
		job.OutputDir,
		info,
		profile,
		l,
		exec,
		buildProgressCallback(job.ProgressHandler, info.Duration),
		encoder.EncoderOptions{
			Threads:  o.threads,
			GPU:      o.gpu,
			LogLevel: o.logLevel,
		},
	)
}

func prepareInputForEncoding(
	ctx context.Context,
	inputPath string,
	exec executor.CommandExecutor,
	opts *options,
) (string, func(), error) {
	if !opts.normalizeOrientation {
		return inputPath, func() {}, nil
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "mosaic-normalized-*"+normalizedInputExt(inputPath))
	if err != nil {
		return "", nil, fmt.Errorf("create temp normalized input: %w", err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", nil, fmt.Errorf("close temp normalized input: %w", err)
	}

	cleanup := func() { _ = os.Remove(tmpPath) }
	if err := normalizeRotationWithExecutor(ctx, inputPath, tmpPath, exec); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("normalize input orientation: %w", err)
	}

	return tmpPath, cleanup, nil
}

func normalizedInputExt(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == "" || strings.Contains(ext, "?") {
		return ".mp4"
	}
	return ext
}
