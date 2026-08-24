package ladder

import (
	"testing"

	"github.com/farshidrezaei/mosaic/probe"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name     string
		expected []Rendition
		info     probe.VideoInfo
	}{
		{
			name: "1080p source - all renditions",
			info: probe.VideoInfo{
				Width:  1920,
				Height: 1080,
				FPS:    30.0,
			},
			expected: []Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5200, BufSize: 10400, Profile: "main", Level: "4.0"},
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "4K source - all renditions",
			info: probe.VideoInfo{
				Width:  3840,
				Height: 2160,
				FPS:    60.0,
			},
			expected: []Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5200, BufSize: 10400, Profile: "main", Level: "4.0"},
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "720p source - 720p and 360p",
			info: probe.VideoInfo{
				Width:  1280,
				Height: 720,
				FPS:    25.0,
			},
			expected: []Rendition{
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "360p source - 360p only",
			info: probe.VideoInfo{
				Width:  640,
				Height: 360,
				FPS:    30.0,
			},
			expected: []Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "240p source - source-sized rendition without upscaling",
			info: probe.VideoInfo{
				Width:  426,
				Height: 240,
				FPS:    30.0,
			},
			expected: []Rendition{
				{Width: 426, Height: 240, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "540p source - 360p only",
			info: probe.VideoInfo{
				Width:  960,
				Height: 540,
				FPS:    30.0,
			},
			expected: []Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "portrait source - aspect-preserving portrait renditions",
			info: probe.VideoInfo{
				Width:  720,
				Height: 1280,
				FPS:    30.0,
			},
			expected: []Rendition{
				{Width: 608, Height: 1080, MaxRate: 5200, BufSize: 10400, Profile: "main", Level: "4.0"},
				{Width: 404, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 202, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "rotated portrait metadata - aspect-preserving portrait renditions",
			info: probe.VideoInfo{
				Width:    1920,
				Height:   1080,
				FPS:      30.0,
				Rotation: 90,
			},
			expected: []Rendition{
				{Width: 608, Height: 1080, MaxRate: 5200, BufSize: 10400, Profile: "main", Level: "4.0"},
				{Width: 404, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 202, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "square source - square renditions",
			info: probe.VideoInfo{
				Width:  1000,
				Height: 1000,
				FPS:    30.0,
			},
			expected: []Rendition{
				{Width: 720, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 360, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "non-standard landscape source - aspect-preserving rendition",
			info: probe.VideoInfo{
				Width:  1280,
				Height: 718,
				FPS:    25.0,
			},
			expected: []Rendition{
				{Width: 642, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Build(tt.info)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d renditions, got %d", len(tt.expected), len(result))
			}

			for i, r := range result {
				if r != tt.expected[i] {
					t.Errorf("rendition %d mismatch:\nexpected: %+v\ngot:      %+v", i, tt.expected[i], r)
				}
			}
		})
	}
}

func TestBuildWithBFrames(t *testing.T) {
	info := probe.VideoInfo{Width: 1920, Height: 1080, FPS: 30.0}
	rungs := Build(info, 3)

	if len(rungs) != 3 {
		t.Fatalf("expected 3 rungs, got %d", len(rungs))
	}

	// Main profile rungs should receive BFrames = 3
	if rungs[0].BFrames != 3 || rungs[1].BFrames != 3 {
		t.Errorf("expected 1080p and 720p to have 3 BFrames, got %d and %d", rungs[0].BFrames, rungs[1].BFrames)
	}

	// Baseline profile rung must always have BFrames = 0
	if rungs[2].BFrames != 0 {
		t.Errorf("expected 360p baseline to have 0 BFrames, got %d", rungs[2].BFrames)
	}
}
