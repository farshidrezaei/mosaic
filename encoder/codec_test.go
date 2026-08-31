package encoder

import (
	"testing"

	"github.com/farshidrezaei/mosaic/config"
)

func TestResolveVideoEncoder(t *testing.T) {
	tests := []struct {
		codec    config.VideoCodec
		gpu      config.GPUType
		expected string
	}{
		{codec: config.CodecH264, gpu: "", expected: "libx264"},
		{codec: config.CodecH264, gpu: config.GPU_NVENC, expected: "h264_nvenc"},
		{codec: config.CodecH264, gpu: config.GPU_VAAPI, expected: "h264_vaapi"},
		{codec: config.CodecH264, gpu: config.GPU_VIDEOTOOLBOX, expected: "h264_videotoolbox"},

		{codec: config.CodecHEVC, gpu: "", expected: "libx265"},
		{codec: config.CodecHEVC, gpu: config.GPU_NVENC, expected: "hevc_nvenc"},
		{codec: config.CodecHEVC, gpu: config.GPU_VAAPI, expected: "hevc_vaapi"},
		{codec: config.CodecHEVC, gpu: config.GPU_VIDEOTOOLBOX, expected: "hevc_videotoolbox"},

		{codec: config.CodecAV1, gpu: "", expected: "libsvtav1"},
		{codec: config.CodecAV1, gpu: config.GPU_NVENC, expected: "av1_nvenc"},
		{codec: config.CodecAV1, gpu: config.GPU_VAAPI, expected: "av1_vaapi"},
	}

	for _, tt := range tests {
		result := ResolveVideoEncoder(tt.codec, tt.gpu)
		if result != tt.expected {
			t.Errorf("ResolveVideoEncoder(%s, %s) = %s, expected %s", tt.codec, tt.gpu, result, tt.expected)
		}
	}
}
