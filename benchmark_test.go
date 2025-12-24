package mediapack

import (
	"context"
	"testing"

	"github.com/farshidrezaei/mosaic/internal/executor"
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

	job := Job{
		Input:     "input.mp4",
		OutputDir: "out",
		Profile:   ProfileVOD,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = initializeWithExecutor(ctx, job, mock)
	}
}
