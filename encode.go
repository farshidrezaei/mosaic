// Package mosaic provides production-ready Adaptive Bitrate (ABR) video packaging for Go.
//
// It probes media streams with FFprobe, computes aspect-preserving ABR rendition ladders,
// applies bitrate optimizations, and generates standardized HLS (fMP4) and DASH CMAF streams
// using single-pass FFmpeg filter graphs.
//
// Key features include:
//   - Standardized HLS (fMP4) and MPEG-DASH CMAF packaging.
//   - Aspect ratio preservation for 16:9, 1:1 square, 9:16 portrait, and custom video formats.
//   - Automatic mobile video display matrix probing and orientation normalization.
//   - Real-time block-accumulated progress reporting with percentage calculation (0.0% to 100.0%).
//   - Hardware acceleration support for NVIDIA NVENC, Intel/AMD VAAPI, and Apple VideoToolbox.
//   - High framerate (>30 FPS) bitrate scaling and configurable B-frame tuning.
//   - Interface-driven testable architecture with 100% mock executor coverage.
//
// Documentation and guides: https://farshidrezaei.github.io/mosaic/
package mosaic

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/farshidrezaei/mosaic/config"
	"github.com/farshidrezaei/mosaic/encoder"
	"github.com/farshidrezaei/mosaic/encryption"
	"github.com/farshidrezaei/mosaic/internal/executor"
	"github.com/farshidrezaei/mosaic/ladder"
	"github.com/farshidrezaei/mosaic/optimize"
	"github.com/farshidrezaei/mosaic/probe"
	"github.com/farshidrezaei/mosaic/storage"
	"github.com/farshidrezaei/mosaic/subtitles"
	"github.com/farshidrezaei/mosaic/thumbnail"
	"github.com/farshidrezaei/mosaic/watermark"
)

// Usage represents resource consumption metrics captured during an encoding run.
type Usage = executor.Usage

// ThumbnailConfig specifies thumbnail sprite generation parameters.
type ThumbnailConfig = thumbnail.Config

// SubtitleTrack defines a subtitle track to be packaged alongside HLS or DASH streams.
type SubtitleTrack = subtitles.Track

// WatermarkConfig specifies watermark overlay parameters.
type WatermarkConfig = watermark.Config

// EncryptionConfig defines the HLS AES-128 encryption configuration.
type EncryptionConfig = encryption.Config

// S3Config defines the cloud storage connection and upload parameters.
type S3Config = storage.S3Config

// VideoCodec represents a video compression format.
type VideoCodec = config.VideoCodec

const (
	// CodecH264 uses H.264 / AVC codec (default).
	CodecH264 = config.CodecH264
	// CodecHEVC uses H.265 / HEVC codec.
	CodecHEVC = config.CodecHEVC
	// CodecAV1 uses AV1 next-generation open codec.
	CodecAV1 = config.CodecAV1
)

// WatermarkPosition defines the placement of a watermark overlay.
type WatermarkPosition = watermark.Position

const (
	PositionTopRight    = watermark.PositionTopRight
	PositionTopLeft     = watermark.PositionTopLeft
	PositionBottomRight = watermark.PositionBottomRight
	PositionBottomLeft  = watermark.PositionBottomLeft
	PositionCenter      = watermark.PositionCenter
)

// Option defines a functional option for configuring encoding jobs.
// It allows for flexible and extensible configuration of the encoding process.
type Option func(*options)

type options struct {
	subtitles            []subtitles.Track
	logger               *slog.Logger
	watermark            *watermark.Config
	encryption           *encryption.Config
	s3Config             *storage.S3Config
	gpu                  config.GPUType
	logLevel             string
	codec                config.VideoCodec
	thumbnailConfig      thumbnail.Config
	threads              int
	bframes              int
	crf                  int
	normalizeOrientation bool
	scaleBitrateWithFPS  bool
	enableThumbnails     bool
	enableIFrames        bool
	normalizeAudio       bool
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

// WithThumbnails enables automatic generation of storyboard thumbnail sprites and WebVTT cue file.
func WithThumbnails(cfg ...ThumbnailConfig) Option {
	return func(o *options) {
		o.enableThumbnails = true
		if len(cfg) > 0 {
			o.thumbnailConfig = cfg[0]
		} else {
			o.thumbnailConfig = thumbnail.DefaultConfig
		}
	}
}

// WithIFrames enables generation of I-frame-only trick-play playlists for HLS.
func WithIFrames(enabled ...bool) Option {
	return func(o *options) {
		if len(enabled) == 0 {
			o.enableIFrames = true
			return
		}
		o.enableIFrames = enabled[0]
	}
}

// WithNormalizeAudio enables EBU R128 loudness normalization for audio tracks.
func WithNormalizeAudio(enabled ...bool) Option {
	return func(o *options) {
		if len(enabled) == 0 {
			o.normalizeAudio = true
			return
		}
		o.normalizeAudio = enabled[0]
	}
}

// WithSubtitles configures subtitle tracks to be converted and injected into HLS and DASH manifests.
func WithSubtitles(tracks ...SubtitleTrack) Option {
	return func(o *options) {
		o.subtitles = append(o.subtitles, tracks...)
	}
}

// WithWatermark enables and configures dynamic logo or watermark overlay on all video renditions.
func WithWatermark(cfg WatermarkConfig) Option {
	return func(o *options) {
		c := cfg
		c.Normalize()
		o.watermark = &c
	}
}

// WithAES128Encryption enables HLS AES-128 segment encryption.
func WithAES128Encryption(cfg ...EncryptionConfig) Option {
	return func(o *options) {
		if len(cfg) > 0 {
			c := cfg[0]
			o.encryption = &c
		} else {
			o.encryption = &EncryptionConfig{}
		}
	}
}

// WithCodec sets the target video compression format (h264, hevc, av1).
func WithCodec(codec VideoCodec) Option {
	return func(o *options) {
		o.codec = codec
	}
}

// WithHEVC configures video encoding to use H.265 / HEVC.
func WithHEVC() Option {
	return WithCodec(CodecHEVC)
}

// WithAV1 configures video encoding to use AV1 (libsvtav1 / av1_nvenc / av1_vaapi).
func WithAV1() Option {
	return WithCodec(CodecAV1)
}

// WithCRF sets the Constant Rate Factor for capped-CRF content-aware bitrate optimization.
func WithCRF(crf int) Option {
	return func(o *options) {
		o.crf = crf
	}
}

// WithS3Upload configures automatic direct upload of generated stream assets to S3/MinIO/R2.
func WithS3Upload(cfg S3Config) Option {
	return func(o *options) {
		c := cfg
		c.Normalize()
		o.s3Config = &c
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
			if m["progress"] == "end" {
				percentage = 100.0
			} else {
				currentSec := encoder.ParseOutTimeSeconds(m)
				percentage = (currentSec / duration) * 100.0
				if percentage > 100.0 {
					percentage = 100.0
				} else if percentage < 0.0 {
					percentage = 0.0
				}
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

	var keyInfoFile string
	if o.encryption != nil {
		var keyErr error
		keyInfoFile, keyErr = encryption.SetupKeyInfo(job.OutputDir, *o.encryption)
		if keyErr != nil {
			return nil, fmt.Errorf("setup encryption keyinfo: %w", keyErr)
		}
	}

	// 2. Encode
	usage, err := encoder.EncodeHLSCMAFWithExecutor(
		ctx,
		effectiveInput,
		job.OutputDir,
		info,
		profile,
		l,
		exec,
		buildProgressCallback(job.ProgressHandler, info.Duration),
		encoder.EncoderOptions{
			Threads:        o.threads,
			GPU:            o.gpu,
			Codec:          o.codec,
			CRF:            o.crf,
			LogLevel:       o.logLevel,
			EnableIFrames:  o.enableIFrames,
			NormalizeAudio: o.normalizeAudio,
			Watermark:      o.watermark,
			KeyInfoFile:    keyInfoFile,
		},
	)
	if err != nil {
		return nil, err
	}

	if o.enableThumbnails {
		if err := thumbnail.GenerateWithExecutor(ctx, effectiveInput, job.OutputDir, info.Duration, o.thumbnailConfig, exec); err != nil {
			return usage, fmt.Errorf("generate thumbnails: %w", err)
		}
	}

	if len(o.subtitles) > 0 {
		if err := subtitles.ProcessTracks(ctx, o.subtitles, job.OutputDir, info.Duration); err != nil {
			return usage, fmt.Errorf("process subtitles: %w", err)
		}
	}

	if o.s3Config != nil {
		if err := storage.UploadDirectory(ctx, job.OutputDir, *o.s3Config, nil); err != nil {
			return usage, fmt.Errorf("upload to S3: %w", err)
		}
	}

	return usage, nil
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
	usage, err := encoder.EncodeDASHCMAFWithExecutor(
		ctx,
		effectiveInput,
		job.OutputDir,
		info,
		profile,
		l,
		exec,
		buildProgressCallback(job.ProgressHandler, info.Duration),
		encoder.EncoderOptions{
			Threads:        o.threads,
			GPU:            o.gpu,
			Codec:          o.codec,
			CRF:            o.crf,
			LogLevel:       o.logLevel,
			NormalizeAudio: o.normalizeAudio,
			Watermark:      o.watermark,
		},
	)
	if err != nil {
		return nil, err
	}

	if o.enableThumbnails {
		if err := thumbnail.GenerateWithExecutor(ctx, effectiveInput, job.OutputDir, info.Duration, o.thumbnailConfig, exec); err != nil {
			return usage, fmt.Errorf("generate thumbnails: %w", err)
		}
	}

	if len(o.subtitles) > 0 {
		if err := subtitles.ProcessTracks(ctx, o.subtitles, job.OutputDir, info.Duration); err != nil {
			return usage, fmt.Errorf("process subtitles: %w", err)
		}
	}

	if o.s3Config != nil {
		if err := storage.UploadDirectory(ctx, job.OutputDir, *o.s3Config, nil); err != nil {
			return usage, fmt.Errorf("upload to S3: %w", err)
		}
	}

	return usage, nil
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
	cleanPath := inputPath
	if u, err := url.Parse(inputPath); err == nil && u.Path != "" {
		cleanPath = u.Path
	}
	if idx := strings.IndexAny(cleanPath, "?#"); idx != -1 {
		cleanPath = cleanPath[:idx]
	}

	ext := strings.ToLower(filepath.Ext(cleanPath))
	switch ext {
	case ".mp4", ".mov", ".mkv", ".webm", ".avi", ".ts", ".m4v":
		return ext
	default:
		return ".mp4"
	}
}
