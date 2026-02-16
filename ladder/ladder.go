package ladder

import "github.com/farshidrezaei/mosaic/probe"

// Build generates an initial encoding ladder based on the source video's height.
// It creates a set of standard renditions (1080p, 720p, 360p) that are suitable
// for adaptive bitrate streaming.
func Build(info probe.VideoInfo) []Rendition {
	var out []Rendition
	portrait := info.IsPortrait()
	sourceHeight := info.DisplayHeight()

	makeRendition := func(width, height, maxRate, bufSize int, profile, level string) Rendition {
		if portrait {
			width, height = height, width
		}
		return Rendition{
			Width:   width,
			Height:  height,
			MaxRate: maxRate,
			BufSize: bufSize,
			Profile: profile,
			Level:   level,
			BFrames: 0,
		}
	}

	if sourceHeight >= 1080 {
		out = append(out, makeRendition(1920, 1080, 5200, 10400, "main", "4.0"))
	}
	if sourceHeight >= 720 {
		out = append(out, makeRendition(1280, 720, 3000, 6000, "main", "3.1"))
	}
	if sourceHeight >= 360 {
		out = append(out, makeRendition(640, 360, 1000, 2000, "baseline", "3.0"))
	}

	if sourceHeight < 360 {
		out = append(out, makeRendition(640, 360, 1000, 2000, "baseline", "3.0"))
	}

	return out
}
