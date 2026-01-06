package ladder

import "github.com/farshidrezaei/mosaic/probe"

// Build generates an initial encoding ladder based on the source video's height.
// It creates a set of standard renditions (1080p, 720p, 360p) that are suitable
// for adaptive bitrate streaming.
func Build(info probe.VideoInfo) []Rendition {
	var out []Rendition

	if info.Height >= 1080 {
		out = append(out, Rendition{Width: 1920, Height: 1080, MaxRate: 5200, BufSize: 10400, Profile: "main", Level: "4.0", BFrames: 0})
	}
	if info.Height >= 720 {
		out = append(out, Rendition{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1", BFrames: 0})
	}
	if info.Height >= 360 {
		out = append(out, Rendition{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0", BFrames: 0})
	}

	if info.Height < 360 {
		out = append(out, Rendition{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0", BFrames: 0})
	}

	return out
}
