package optimize

import "github.com/farshidrezaei/mosaic/ladder"

func Apply(in []ladder.Rendition) []ladder.Rendition {
	var out []ladder.Rendition

	for _, r := range in {
		r.MaxRate = capBitrate(r.Height, r.MaxRate)
		r.BufSize = r.MaxRate * 2
		out = append(out, r)
	}

	return trim(out)
}

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
