package ladder

import (
	"testing"

	"github.com/farshidrezaei/mosaic/probe"
)

func BenchmarkBuild(b *testing.B) {
	info := probe.VideoInfo{
		Width:  1920,
		Height: 1080,
		FPS:    30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Build(info)
	}
}
