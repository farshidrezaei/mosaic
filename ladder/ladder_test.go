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
			name: "240p source - no renditions",
			info: probe.VideoInfo{
				Width:  426,
				Height: 240,
				FPS:    30.0,
			},
			expected: []Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
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
			name: "portrait source - portrait renditions",
			info: probe.VideoInfo{
				Width:  720,
				Height: 1280,
				FPS:    30.0,
			},
			expected: []Rendition{
				{Width: 1080, Height: 1920, MaxRate: 5200, BufSize: 10400, Profile: "main", Level: "4.0"},
				{Width: 720, Height: 1280, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 360, Height: 640, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "rotated portrait metadata - portrait renditions",
			info: probe.VideoInfo{
				Width:    1920,
				Height:   1080,
				FPS:      30.0,
				Rotation: 90,
			},
			expected: []Rendition{
				{Width: 1080, Height: 1920, MaxRate: 5200, BufSize: 10400, Profile: "main", Level: "4.0"},
				{Width: 720, Height: 1280, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 360, Height: 640, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
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
