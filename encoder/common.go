package encoder

import (
	"fmt"
	"math"
	"strings"
)

func calcGOP(fps float64, segmentSec int) int {
	gop := int(math.Round(fps * float64(segmentSec)))

	// x264 prefers even GOP
	if gop%2 != 0 {
		gop++
	}

	if gop < 24 {
		gop = 24
	}

	return gop
}

// buildVarStreamMap generates the var_stream_map string for FFmpeg's HLS muxer.
func buildVarStreamMap(variants int, hasAudio bool) string {
	var parts []string

	for i := 0; i < variants; i++ {
		if hasAudio {
			parts = append(parts, fmt.Sprintf("v:%d,a:%d", i, i))
		} else {
			parts = append(parts, fmt.Sprintf("v:%d", i))
		}
	}

	return strings.Join(parts, " ")
}
