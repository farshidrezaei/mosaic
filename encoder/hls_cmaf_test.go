package encoder

import (
	"testing"

	"github.com/farshidrezaei/mosaic/ladder"
)

func TestBuildFilterGraph(t *testing.T) {
	tests := []struct {
		name       string
		renditions []ladder.Rendition
		expected   string
	}{
		{
			name: "single rendition",
			renditions: []ladder.Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			expected: "[0:v]split=1[v0];[v0]scale=640:360:force_original_aspect_ratio=decrease,pad=640:360:(ow-iw)/2:(oh-ih)/2,setsar=1[v0o]",
		},
		{
			name: "three renditions",
			renditions: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			expected: "[0:v]split=3[v0][v1][v2];[v0]scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2,setsar=1[v0o];[v1]scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2,setsar=1[v1o];[v2]scale=640:360:force_original_aspect_ratio=decrease,pad=640:360:(ow-iw)/2:(oh-ih)/2,setsar=1[v2o]",
		},
		{
			name: "two renditions",
			renditions: []ladder.Rendition{
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			expected: "[0:v]split=2[v0][v1];[v0]scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2,setsar=1[v0o];[v1]scale=640:360:force_original_aspect_ratio=decrease,pad=640:360:(ow-iw)/2:(oh-ih)/2,setsar=1[v1o]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFilterGraph(tt.renditions)
			if result != tt.expected {
				t.Errorf("filter graph mismatch:\nexpected: %s\ngot:      %s", tt.expected, result)
			}
		})
	}
}
