package encoder

import (
	"fmt"
	"math"
	"os"
	"strconv"
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

func appendVideoRotationMetadataReset(args []string, streamIndex int) []string {
	return append(args,
		fmt.Sprintf("-metadata:s:v:%d", streamIndex), "rotate=0",
	)
}

func ensureOutputDir(outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	return nil
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

// StreamProgress reads updates from progressChan, accumulates key-value pairs into blocks,
// and invokes handler whenever a progress block completes (at "progress=continue" or "progress=end").
func StreamProgress(progressChan <-chan string, handler func(map[string]string)) {
	if handler == nil {
		for range progressChan {
		}
		return
	}

	block := make(map[string]string)
	emitted := false
	for raw := range progressChan {
		lines := strings.Split(raw, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				// Retain last known valid values if FFmpeg emits "N/A" (e.g. at progress=end)
				if v == "N/A" && block[k] != "" && block[k] != "N/A" {
					continue
				}
				block[k] = v

				if k == "progress" {
					if v == "continue" || v == "end" {
						snapshot := make(map[string]string, len(block))
						for key, val := range block {
							snapshot[key] = val
						}
						handler(snapshot)
						emitted = true
					}
				}
			}
		}
	}

	if !emitted && len(block) > 0 {
		handler(block)
	}
}

// ParseOutTimeSeconds converts FFmpeg progress timestamps (out_time_us, out_time_ms, or out_time) to seconds.
func ParseOutTimeSeconds(m map[string]string) float64 {
	if usStr, ok := m["out_time_us"]; ok && usStr != "" {
		if us, err := strconv.ParseInt(strings.TrimSpace(usStr), 10, 64); err == nil && us >= 0 {
			return float64(us) / 1000000.0
		}
	}

	if msStr, ok := m["out_time_ms"]; ok && msStr != "" {
		if ms, err := strconv.ParseInt(strings.TrimSpace(msStr), 10, 64); err == nil && ms >= 0 {
			return float64(ms) / 1000.0
		}
	}

	if outTime, ok := m["out_time"]; ok && outTime != "" {
		parts := strings.Split(strings.TrimSpace(outTime), ":")
		if len(parts) == 3 {
			h, errH := strconv.ParseFloat(parts[0], 64)
			min, errM := strconv.ParseFloat(parts[1], 64)
			sec, errS := strconv.ParseFloat(parts[2], 64)
			if errH == nil && errM == nil && errS == nil {
				return h*3600.0 + min*60.0 + sec
			}
		}
	}

	return 0
}
