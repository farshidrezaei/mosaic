package encoder

import (
	"fmt"
	"math"
	"strings"
)

// calcGOP calculates the Group of Pictures (GOP) size based on FPS and segment duration.
// It ensures the GOP is even (preferred by x264) and at least 24 frames.
func calcGOP(fps float64, segmentSec int) int {
	gop := int(math.Round(fps * float64(segmentSec)))

	// x264 and many hardware encoders prefer even GOP sizes for better alignment.
	if gop%2 != 0 {
		gop++
	}

	// Minimum GOP size to ensure stability in very low FPS scenarios.
	if gop < 24 {
		gop = 24
	}

	return gop
}

// buildVarStreamMap generates the var_stream_map string for FFmpeg's HLS muxer.
// It maps video and audio streams to variant groups (e.g., "v:0,a:0 v:1,a:1").
func buildVarStreamMap(variants int, hasAudio bool) string {
	var parts []string

	for i := 0; i < variants; i++ {
		if hasAudio {
			// Map video stream i and audio stream i to the same variant.
			parts = append(parts, fmt.Sprintf("v:%d,a:%d", i, i))
		} else {
			parts = append(parts, fmt.Sprintf("v:%d", i))
		}
	}

	return strings.Join(parts, " ")
}

// ParseProgress parses FFmpeg's machine-readable progress output (from -progress pipe:1).
// It returns a map of keys and values (e.g., "frame" -> "100", "out_time" -> "00:00:10.000000").
func ParseProgress(raw string) map[string]string {
	lines := strings.Split(raw, "\n")
	progress := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			// Trim whitespace to handle potential variations in FFmpeg output.
			progress[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return progress
}
