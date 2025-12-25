package mosaic

import (
	"context"

	"log/slog"

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
	logger   *slog.Logger
	gpu      config.GPUType
	logLevel string
	threads  int
}

func defaultOptions() *options {
	return &options{
		threads:  0, // auto
		gpu:      "",
		logLevel: "warning",
		logger:   slog.Default(),
	}
}

// WithThreads sets the number of CPU threads to use for encoding.
// Set to 0 (default) to let FFmpeg auto-detect the optimal number of threads.
func WithThreads(n int) Option {
	return func(o *options) {
		o.threads = n
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

	// build ladder
	l := ladder.Build(info)

	// cost optimizer
	l = optimize.Apply(l)

	// profile
	var profile config.Profile
	switch job.Profile {
	case ProfileLive:
		profile = config.LIVE
	default:
		profile = config.VOD
	}

	opts.logger.Info("encoding variants", "count", len(l))

	return info, profile, l, err

}

// EncodeHls encodes the given job into HLS format with CMAF segments.
// It automatically builds an optimized encoding ladder and generates a master playlist.
// Functional options can be provided to customize the encoding process.
func EncodeHls(ctx context.Context, job Job, opts ...Option) error {
	return EncodeHlsWithExecutor(ctx, job, executor.DefaultExecutor, opts...)
}

// EncodeHlsWithExecutor is like EncodeHls but allows providing a custom CommandExecutor.
// This is primarily used for testing or advanced command execution scenarios.
func EncodeHlsWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) error {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	info, profile, l, err := initializeWithExecutor(ctx, job, exec, o)
	if err != nil {
		return err
	}
	// 2. Encode
	return encoder.EncodeHLSCMAFWithExecutor(
		ctx,
		job.Input,
		job.OutputDir,
		info,
		profile,
		l,
		exec,
		func(m map[string]string) {
			if job.ProgressHandler != nil {
				job.ProgressHandler(ProgressInfo{
					Percentage:  0,
					CurrentTime: m["out_time"],
					Bitrate:     m["bitrate"],
					Speed:       m["speed"],
				})
			}
		},
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
func EncodeDash(ctx context.Context, job Job, opts ...Option) error {
	return EncodeDashWithExecutor(ctx, job, executor.DefaultExecutor, opts...)
}

// EncodeDashWithExecutor is like EncodeDash but allows providing a custom CommandExecutor.
// This is primarily used for testing or advanced command execution scenarios.
func EncodeDashWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor, opts ...Option) error {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	info, profile, l, err := initializeWithExecutor(ctx, job, exec, o)
	if err != nil {
		return err
	}
	// 2. Encode
	return encoder.EncodeDASHCMAFWithExecutor(
		ctx,
		job.Input,
		job.OutputDir,
		info,
		profile,
		l,
		exec,
		func(m map[string]string) {
			if job.ProgressHandler != nil {
				job.ProgressHandler(ProgressInfo{
					Percentage:  0,
					CurrentTime: m["out_time"],
					Bitrate:     m["bitrate"],
					Speed:       m["speed"],
				})
			}
		},
		encoder.EncoderOptions{
			Threads:  o.threads,
			GPU:      o.gpu,
			LogLevel: o.logLevel,
		},
	)
}
