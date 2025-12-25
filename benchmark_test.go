package mosaic

import (
	"context"
	"fmt"
	"testing"

	"github.com/farshidrezaei/mosaic/internal/executor"
	"github.com/farshidrezaei/mosaic/probe"
)

func BenchmarkInitialize(b *testing.B) {
	// Mock executor to avoid real command overhead
	mock := &executor.MockCommandExecutor{
		Responses: map[string]executor.MockResponse{
			"ffprobe": {
				Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
			},
		},
	}

	ctx := context.Background()

	profiles := []struct {
		name    string
		profile Profile
	}{
		{"VOD", ProfileVOD},
		{"Live", ProfileLive},
	}

	for _, p := range profiles {
		b.Run(fmt.Sprintf("Profile_%s", p.name), func(b *testing.B) {
			job := Job{
				Input:     "input.mp4",
				OutputDir: "out",
				Profile:   p.profile,
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _, _ = initializeWithExecutor(ctx, job, mock, defaultOptions())
			}
		})
	}
}

func BenchmarkProbe(b *testing.B) {
	mock := &executor.MockCommandExecutor{
		Responses: map[string]executor.MockResponse{
			"ffprobe": {
				Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
			},
		},
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = probe.InputWithExecutor(ctx, "input.mp4", mock)
	}
}
