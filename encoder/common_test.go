package encoder

import "testing"

func TestCalcGOP(t *testing.T) {
	tests := []struct {
		name       string
		fps        float64
		segmentSec int
		expected   int
	}{
		{"30fps 5sec segment", 30.0, 5, 150},
		{"24fps 5sec segment", 24.0, 5, 120},
		{"29.97fps 5sec segment", 29.97, 5, 150},
		{"25fps 2sec segment", 25.0, 2, 50},
		{"60fps 5sec segment", 60.0, 5, 300},
		{"23.976fps 5sec segment", 23.976, 5, 120},
		{"low fps forces minimum", 4.0, 5, 24}, // 20 rounds to 24
		{"odd GOP becomes even", 5.0, 5, 26},   // 25 -> 26
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcGOP(tt.fps, tt.segmentSec)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestBuildVarStreamMap(t *testing.T) {
	tests := []struct {
		name     string
		variants int
		hasAudio bool
		expected string
	}{
		{"1 variant with audio", 1, true, "v:0,a:0"},
		{"1 variant no audio", 1, false, "v:0"},
		{"3 variants with audio", 3, true, "v:0,a:0 v:1,a:1 v:2,a:2"},
		{"3 variants no audio", 3, false, "v:0 v:1 v:2"},
		{"2 variants with audio", 2, true, "v:0,a:0 v:1,a:1"},
		{"2 variants no audio", 2, false, "v:0 v:1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildVarStreamMap(tt.variants, tt.hasAudio)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
