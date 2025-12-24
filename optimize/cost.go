package optimize

func capBitrate(height int, bitrate int) int {
	switch {
	case height >= 1080 && bitrate > 5000:
		return 5000
	case height >= 720 && bitrate > 3000:
		return 3000
	case bitrate > 1000:
		return 1000
	default:
		return bitrate
	}
}
