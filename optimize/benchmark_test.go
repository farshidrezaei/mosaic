package optimize

import (
	"testing"

	"github.com/farshidrezaei/mosaic/ladder"
)

func BenchmarkApply(b *testing.B) {
	l := []ladder.Rendition{
		{Width: 1920, Height: 1080, MaxRate: 6000, BufSize: 12000},
		{Width: 1280, Height: 720, MaxRate: 4000, BufSize: 8000},
		{Width: 640, Height: 360, MaxRate: 1500, BufSize: 3000},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Apply(l)
	}
}
