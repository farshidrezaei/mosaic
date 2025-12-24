package mosaic

import (
	"context"
	"fmt"

	"github.com/farshidrezaei/mosaic/config"
	"github.com/farshidrezaei/mosaic/encoder"
	"github.com/farshidrezaei/mosaic/internal/executor"
	"github.com/farshidrezaei/mosaic/ladder"
	"github.com/farshidrezaei/mosaic/optimize"
	"github.com/farshidrezaei/mosaic/probe"
)

func initialize(ctx context.Context, job Job) (probe.VideoInfo, config.Profile, []ladder.Rendition, error) {
	return initializeWithExecutor(ctx, job, executor.DefaultExecutor)
}

func initializeWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor) (probe.VideoInfo, config.Profile, []ladder.Rendition, error) {
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

	fmt.Printf("encoding %d variants\n", len(l))

	return info, profile, l, err

}

// EncodeHls encodes the given job into HLS format with CMAF segments.
// It uses the default command executor.
func EncodeHls(ctx context.Context, job Job) error {
	return EncodeHlsWithExecutor(ctx, job, executor.DefaultExecutor)
}

// EncodeHlsWithExecutor encodes the given job into HLS format using the provided executor.
func EncodeHlsWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor) error {
	info, profile, l, err := initializeWithExecutor(ctx, job, exec)
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
	)
}

// EncodeDash encodes the given job into DASH format with CMAF segments.
// It uses the default command executor.
func EncodeDash(ctx context.Context, job Job) error {
	return EncodeDashWithExecutor(ctx, job, executor.DefaultExecutor)
}

// EncodeDashWithExecutor encodes the given job into DASH format using the provided executor.
func EncodeDashWithExecutor(ctx context.Context, job Job, exec executor.CommandExecutor) error {
	info, profile, l, err := initializeWithExecutor(ctx, job, exec)
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
	)
}
