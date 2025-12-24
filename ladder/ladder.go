package ladder

import "github.com/farshidrezaei/mosaic/probe"

// Build generates an initial encoding ladder based on the source video information.
// It includes renditions for 1080p, 720p, and 360p if the source height is sufficient.
func Build(info probe.VideoInfo) []Rendition {
	var out []Rendition

	if info.Height >= 1080 {
		out = append(out, Rendition{1920, 1080, 5200, 10400, "main", "4.0"})
	}
	if info.Height >= 720 {
		out = append(out, Rendition{1280, 720, 3000, 6000, "main", "3.1"})
	}
	if info.Height >= 360 {
		out = append(out, Rendition{640, 360, 1000, 2000, "baseline", "3.0"})
	}

	return out
}
