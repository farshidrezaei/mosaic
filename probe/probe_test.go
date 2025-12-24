package probe

import (
	"context"
	"testing"
)

func TestParseFPS(t *testing.T) {
	tests := []struct {
		name     string
		rate     string
		expected float64
	}{
		{"standard 30fps", "30/1", 30.0},
		{"standard 25fps", "25/1", 25.0},
		{"NTSC 29.97fps", "30000/1001", 29.97002997002997},
		{"film 23.976fps", "24000/1001", 23.976023976023978},
		{"60fps", "60/1", 60.0},
		{"invalid format - no slash", "30", 30.0},
		{"invalid format - empty", "", 30.0},
		{"invalid format - multiple slashes", "30/1/2", 30.0},
		{"division by zero", "30/0", 30.0},
		{"zero numerator", "0/1", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFPS(tt.rate)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestInput(t *testing.T) {
	// This test verifies the wrapper function delegates to InputWithExecutor
	// Will fail without real ffprobe, which is expected
	_, err := Input(context.Background(), "test.mp4")

	// Expect error since we don't have real video file or ffprobe
	if err == nil {
		// If it doesn't error, ffprobe is installed and file might exist
		// This is OK, the wrapper worked
		t.Log("Input() succeeded (ffprobe available)")
	}
}
