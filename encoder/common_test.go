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

func TestParseOutTimeSeconds(t *testing.T) {
	tests := []struct {
		input    map[string]string
		name     string
		expected float64
	}{
		{
			name:     "from out_time_us",
			input:    map[string]string{"out_time_us": "12345678"},
			expected: 12.345678,
		},
		{
			name:     "from out_time_ms",
			input:    map[string]string{"out_time_ms": "12345"},
			expected: 12.345,
		},
		{
			name:     "from out_time string",
			input:    map[string]string{"out_time": "01:02:03.500000"},
			expected: 3723.5, // 1*3600 + 2*60 + 3.5
		},
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: 0.0,
		},
		{
			name:     "malformed out_time",
			input:    map[string]string{"out_time": "invalid"},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseOutTimeSeconds(tt.input)
			if got != tt.expected {
				t.Errorf("ParseOutTimeSeconds() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStreamProgress(t *testing.T) {
	ch := make(chan string, 10)
	ch <- "frame=10"
	ch <- "fps=30"
	ch <- "out_time=00:00:01.000000"
	ch <- "speed=1.5x"
	ch <- "progress=continue"
	ch <- "frame=20"
	ch <- "out_time=00:00:02.000000"
	ch <- "progress=end"
	close(ch)

	var snapshots []map[string]string
	StreamProgress(ch, func(m map[string]string) {
		snapshots = append(snapshots, m)
	})

	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}

	// Snapshot 1
	if snapshots[0]["frame"] != "10" || snapshots[0]["out_time"] != "00:00:01.000000" || snapshots[0]["speed"] != "1.5x" {
		t.Errorf("unexpected snapshot 0: %+v", snapshots[0])
	}

	// Snapshot 2
	if snapshots[1]["frame"] != "20" || snapshots[1]["out_time"] != "00:00:02.000000" || snapshots[1]["progress"] != "end" {
		t.Errorf("unexpected snapshot 1: %+v", snapshots[1])
	}
}
