package encoder

import (
	"reflect"
	"testing"
)

func TestParseProgress(t *testing.T) {
	tests := []struct {
		expected map[string]string
		name     string
		input    string
	}{
		{
			name:  "standard progress output",
			input: "frame=100\nfps=30.0\nbitrate=1000.0kbits/s\nout_time=00:00:10.000000\nspeed=1.5x\nprogress=continue\n",
			expected: map[string]string{
				"frame":    "100",
				"fps":      "30.0",
				"bitrate":  "1000.0kbits/s",
				"out_time": "00:00:10.000000",
				"speed":    "1.5x",
				"progress": "continue",
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "malformed input",
			input: "invalid line\nkey=value\n",
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			name:  "input with spaces",
			input: "key = value \n",
			expected: map[string]string{
				"key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseProgress(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseProgress() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestCalcGOP(t *testing.T) {
	tests := []struct {
		fps        float64
		segmentSec int
		want       int
	}{
		{fps: 30.0, segmentSec: 2, want: 60},
		{fps: 23.976, segmentSec: 2, want: 48}, // 47.952 -> 48
		{fps: 30.0, segmentSec: 5, want: 150},
		{fps: 10.0, segmentSec: 2, want: 24}, // 20 -> 24 (min)
		{fps: 25.0, segmentSec: 1, want: 26}, // 25 -> 26 (even)
	}

	for _, tt := range tests {
		got := calcGOP(tt.fps, tt.segmentSec)
		if got != tt.want {
			t.Errorf("calcGOP(%v, %v) = %v, want %v", tt.fps, tt.segmentSec, got, tt.want)
		}
	}
}

func TestCalcGOPProperties(t *testing.T) {
	// Property: GOP should always be even and >= 24
	for fps := 1.0; fps <= 120.0; fps += 0.5 {
		for seg := 1; seg <= 10; seg++ {
			got := calcGOP(fps, seg)
			if got%2 != 0 {
				t.Errorf("calcGOP(%v, %v) = %v, want even", fps, seg, got)
			}
			if got < 24 {
				t.Errorf("calcGOP(%v, %v) = %v, want >= 24", fps, seg, got)
			}
		}
	}
}
