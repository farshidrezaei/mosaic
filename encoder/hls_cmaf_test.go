package encoder

import (
	"strings"
	"testing"

	"github.com/farshidrezaei/mosaic/ladder"
	"github.com/farshidrezaei/mosaic/probe"
	"github.com/farshidrezaei/mosaic/watermark"
)

func TestBuildFilterGraph(t *testing.T) {
	info := probe.VideoInfo{Width: 1920, Height: 1080}
	tests := []struct {
		name       string
		expected   string
		watermark  *watermark.Config
		renditions []ladder.Rendition
	}{
		{
			name: "single rendition",
			renditions: []ladder.Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			expected: "[0:v]split=1[v0];[v0]scale=640:360,setsar=1,setdar=1920/1080[v0o]",
		},
		{
			name: "three renditions",
			renditions: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			expected: "[0:v]split=3[v0][v1][v2];[v0]scale=1920:1080,setsar=1,setdar=1920/1080[v0o];[v1]scale=1280:720,setsar=1,setdar=1920/1080[v1o];[v2]scale=640:360,setsar=1,setdar=1920/1080[v2o]",
		},
		{
			name: "with watermark",
			renditions: []ladder.Rendition{
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
			},
			watermark: &watermark.Config{
				Path:     "logo.png",
				Position: watermark.PositionTopRight,
				Opacity:  0.8,
			},
			expected: "overlay=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFilterGraph(tt.renditions, info, tt.watermark)
			if tt.watermark != nil {
				if !strings.Contains(result, tt.expected) {
					t.Errorf("filter graph missing watermark overlay:\nexpected substring: %s\ngot: %s", tt.expected, result)
				}
			} else if result != tt.expected {
				t.Errorf("filter graph mismatch:\nexpected: %s\ngot:      %s", tt.expected, result)
			}
		})
	}
}
