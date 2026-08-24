package optimize

import "math"

func capBitrate(height int, bitrate int, fps float64, scaleWithFPS bool) int {
	factor := 1.0
	if scaleWithFPS && fps > 30 {
		factor = math.Min(1.5, fps/30.0)
	}

	cap1080 := int(math.Round(5000 * factor))
	cap720 := int(math.Round(3000 * factor))
	capLow := int(math.Round(1000 * factor))

	switch {
	case height >= 1080 && bitrate > cap1080:
		return cap1080
	case height >= 720 && bitrate > cap720:
		return cap720
	case bitrate > capLow:
		return capLow
	default:
		return bitrate
	}
}
