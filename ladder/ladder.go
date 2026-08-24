package ladder

import (
	"math"

	"github.com/farshidrezaei/mosaic/probe"
)

// Build generates an initial encoding ladder based on the source video's height.
// It creates standard quality rungs while preserving the source display aspect ratio.
// Optional bframes parameter configures the number of B-frames for non-baseline profiles (default 0).
func Build(info probe.VideoInfo, bframes ...int) []Rendition {
	var out []Rendition
	sourceWidth := info.DisplayWidth()
	sourceHeight := info.DisplayHeight()
	if sourceWidth <= 0 || sourceHeight <= 0 {
		return out
	}

	bf := 0
	if len(bframes) > 0 && bframes[0] >= 0 {
		bf = bframes[0]
	}

	makeRendition := func(height, maxRate, bufSize int, profile, level string) Rendition {
		height = evenDimension(height)
		width := scaledWidth(sourceWidth, sourceHeight, height)
		rBFrames := bf
		if profile == "baseline" {
			rBFrames = 0
		}
		return Rendition{
			Width:   width,
			Height:  height,
			MaxRate: maxRate,
			BufSize: bufSize,
			Profile: profile,
			Level:   level,
			BFrames: rBFrames,
		}
	}

	if sourceHeight >= 1080 {
		out = append(out, makeRendition(1080, 5200, 10400, "main", "4.0"))
	}
	if sourceHeight >= 720 {
		out = append(out, makeRendition(720, 3000, 6000, "main", "3.1"))
	}
	if sourceHeight >= 360 {
		out = append(out, makeRendition(360, 1000, 2000, "baseline", "3.0"))
	}

	if sourceHeight < 360 {
		out = append(out, makeRendition(sourceHeight, 1000, 2000, "baseline", "3.0"))
	}

	return out
}

func scaledWidth(sourceWidth, sourceHeight, targetHeight int) int {
	if sourceHeight <= 0 {
		return evenDimension(sourceWidth)
	}
	width := int(math.Round(float64(targetHeight) * float64(sourceWidth) / float64(sourceHeight)))
	return evenDimension(width)
}

func evenDimension(n int) int {
	if n < 2 {
		return 2
	}
	if n%2 != 0 {
		n--
	}
	return n
}
