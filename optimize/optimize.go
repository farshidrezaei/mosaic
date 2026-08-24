package optimize

import "github.com/farshidrezaei/mosaic/ladder"

// Option configures bitrate optimization behavior.
type Option func(*options)

type options struct {
	fps          float64
	scaleWithFPS bool
}

// WithFPS enables optional bitrate scaling based on video frame rate.
func WithFPS(fps float64) Option {
	return func(o *options) {
		o.fps = fps
		o.scaleWithFPS = true
	}
}

// Apply performs bitrate optimization and rendition trimming on the encoding ladder.
// It caps bitrates based on resolution and removes redundant renditions that are
// too close in resolution to each other.
func Apply(in []ladder.Rendition, opts ...Option) []ladder.Rendition {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	var out []ladder.Rendition

	for _, r := range in {
		r.MaxRate = capBitrate(r.Height, r.MaxRate, o.fps, o.scaleWithFPS)
		r.BufSize = r.MaxRate * 2
		out = append(out, r)
	}

	return trim(out)
}

// trim removes renditions that are too close in resolution to the previous one.
// It uses a 0.7 height ratio threshold to determine if a rendition is redundant.
func trim(in []ladder.Rendition) []ladder.Rendition {
	if len(in) <= 1 {
		return in
	}

	var res []ladder.Rendition
	res = append(res, in[0])

	for i := 1; i < len(in); i++ {
		prev := res[len(res)-1]
		curr := in[i]

		if float64(curr.Height)/float64(prev.Height) < 0.7 {
			res = append(res, curr)
		}
	}
	return res
}
