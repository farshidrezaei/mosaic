package optimize

import (
	"testing"

	"github.com/farshidrezaei/mosaic/ladder"
)

func TestCapBitrate(t *testing.T) {
	tests := []struct {
		name     string
		height   int
		bitrate  int
		expected int
	}{
		{"1080p - cap at 5000", 1080, 6000, 5000},
		{"1080p - under cap triggers 720 rule", 1080, 4500, 3000},     // cascades to case height >= 720
		{"1080p - at 5000 not > triggers 720 rule", 1080, 5000, 3000}, // 5000 is not > 5000, cascades
		{"720p - cap at 3000", 720, 4000, 3000},
		{"720p - under cap triggers 1000 rule", 720, 2500, 1000},     // cascades to case bitrate > 1000
		{"720p - at 3000 not > triggers 1000 rule", 720, 3000, 1000}, // 3000 is not > 3000, cascades
		{"720p - under 1000", 720, 800, 800},
		{"360p - cap at 1000", 360, 1500, 1000},
		{"360p - under cap", 360, 800, 800},
		{"360p - at 1000 passes through", 360, 1000, 1000}, // 1000 is not > 1000, returns as-is
		{"240p - cap at 1000", 240, 1200, 1000},
		{"240p - under cap", 240, 500, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := capBitrate(tt.height, tt.bitrate, 0, false)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCapBitrateWithFPS(t *testing.T) {
	// For 60fps with scaling enabled (factor = 60/30 = 2.0 capped at 1.5):
	// 1080p cap becomes 5000 * 1.5 = 7500
	// 720p cap becomes 3000 * 1.5 = 4500
	// Low cap becomes 1000 * 1.5 = 1500

	if got := capBitrate(1080, 8000, 60.0, true); got != 7500 {
		t.Errorf("expected 1080p@60fps cap 7500, got %d", got)
	}
	if got := capBitrate(720, 5000, 60.0, true); got != 4500 {
		t.Errorf("expected 720p@60fps cap 4500, got %d", got)
	}
	if got := capBitrate(360, 2000, 60.0, true); got != 1500 {
		t.Errorf("expected 360p@60fps cap 1500, got %d", got)
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name     string
		input    []ladder.Rendition
		expected []ladder.Rendition
	}{
		{
			name: "standard 1080p ladder - caps based on cascading rules",
			input: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5200, BufSize: 10400, Profile: "main", Level: "4.0"},
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			expected: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},  // 5200 > 5000 at 1080p
				{Width: 1280, Height: 720, MaxRate: 1000, BufSize: 2000, Profile: "main", Level: "3.1"},    // 3000 at 720p triggers > 1000 case
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"}, // 1000 at 360p triggers > 1000 case
			},
		},
		{
			name: "excessive bitrates - all capped",
			input: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 8000, BufSize: 16000, Profile: "main", Level: "4.0"},
				{Width: 1280, Height: 720, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1500, BufSize: 3000, Profile: "baseline", Level: "3.0"},
			},
			expected: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "close resolutions - trim middle",
			input: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
				{Width: 1600, Height: 900, MaxRate: 4000, BufSize: 8000, Profile: "main", Level: "3.2"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			expected: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "4.0"}, // 5000 at 1080p but < 5000 threshold triggers 720 rule
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "single rendition - no trimming",
			input: []ladder.Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			expected: []ladder.Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name:     "empty input",
			input:    []ladder.Rendition{},
			expected: []ladder.Rendition{},
		},
		{
			name: "very small resolution - cap at 1000",
			input: []ladder.Rendition{
				{Width: 160, Height: 90, MaxRate: 2000, BufSize: 4000, Profile: "baseline", Level: "1.0"},
			},
			expected: []ladder.Rendition{
				{Width: 160, Height: 90, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "1.0"},
			},
		},
		{
			name: "very high bitrate at low resolution",
			input: []ladder.Rendition{
				{Width: 640, Height: 360, MaxRate: 10000, BufSize: 20000, Profile: "main", Level: "3.1"},
			},
			expected: []ladder.Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "main", Level: "3.1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Apply(tt.input)

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

func TestTrim(t *testing.T) {
	tests := []struct {
		name     string
		input    []ladder.Rendition
		expected []ladder.Rendition
	}{
		{
			name: "keeps renditions with ratio >= 0.7",
			input: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},    // 720/1080 = 0.666 < 0.7, keep
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"}, // 360/720 = 0.5 < 0.7, keep
			},
			expected: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "trims very close resolutions",
			input: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
				{Width: 1600, Height: 900, MaxRate: 4000, BufSize: 8000, Profile: "main", Level: "3.2"},    // 900/1080 = 0.833 >= 0.7, skip
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"}, // 360/1080 = 0.333 < 0.7, keep
			},
			expected: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
		},
		{
			name: "single rendition - no changes",
			input: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
			},
			expected: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
			},
		},
		{
			name:     "empty list",
			input:    []ladder.Rendition{},
			expected: []ladder.Rendition{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trim(tt.input)

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

func TestApplyWithFPS(t *testing.T) {
	in := []ladder.Rendition{
		{Width: 1920, Height: 1080, MaxRate: 8000, BufSize: 16000, Profile: "main", Level: "4.0"},
	}

	out := Apply(in, WithFPS(60.0))
	if len(out) != 1 {
		t.Fatalf("expected 1 rendition, got %d", len(out))
	}
	if out[0].MaxRate != 7500 {
		t.Errorf("expected 1080p@60fps MaxRate 7500, got %d", out[0].MaxRate)
	}
	if out[0].BufSize != 15000 {
		t.Errorf("expected BufSize 15000, got %d", out[0].BufSize)
	}
}
