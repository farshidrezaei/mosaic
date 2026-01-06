package ladder

// Rendition represents a single video quality level in the encoding ladder.
type Rendition struct {
	// Profile is the H.264 profile (e.g., "main", "baseline").
	Profile string
	// Level is the H.264 level (e.g., "4.0", "3.1").
	Level string
	// Width is the horizontal resolution of the rendition.
	Width int
	// Height is the vertical resolution of the rendition.
	Height int
	// MaxRate is the maximum bitrate in kbps.
	MaxRate int
	// BufSize is the VBV buffer size in kbps.
	BufSize int
	// BFrames number of B-frames (Bidirectional frames) between I/P frames.
	BFrames int
}
