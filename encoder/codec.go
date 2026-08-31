package encoder

import "github.com/farshidrezaei/mosaic/config"

// ResolveVideoEncoder selects the appropriate FFmpeg video encoder string based on codec and GPU backend.
func ResolveVideoEncoder(codec config.VideoCodec, gpu config.GPUType) string {
	switch codec {
	case config.CodecHEVC:
		switch gpu {
		case config.GPU_NVENC:
			return "hevc_nvenc"
		case config.GPU_VAAPI:
			return "hevc_vaapi"
		case config.GPU_VIDEOTOOLBOX:
			return "hevc_videotoolbox"
		default:
			return "libx265"
		}
	case config.CodecAV1:
		switch gpu {
		case config.GPU_NVENC:
			return "av1_nvenc"
		case config.GPU_VAAPI:
			return "av1_vaapi"
		default:
			return "libsvtav1"
		}
	case config.CodecH264:
		fallthrough
	default:
		switch gpu {
		case config.GPU_NVENC:
			return "h264_nvenc"
		case config.GPU_VAAPI:
			return "h264_vaapi"
		case config.GPU_VIDEOTOOLBOX:
			return "h264_videotoolbox"
		default:
			return "libx264"
		}
	}
}
