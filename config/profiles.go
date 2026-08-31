package config

// Profile defines the technical configuration for an encoding profile,
// including segment duration, low-latency settings, and B-frame count.
type Profile struct {
	// SegmentDuration is the duration of each media segment in seconds.
	SegmentDuration int
	// LowLatency indicates if low-latency features (like chunked transfer) should be enabled.
	LowLatency bool
	// BFrames specifies the number of B-frames between reference frames (defaults to 0).
	BFrames int
}

// VideoCodec represents a video compression format.
type VideoCodec string

const (
	// CodecH264 uses H.264 / AVC codec (default).
	CodecH264 VideoCodec = "h264"
	// CodecHEVC uses H.265 / HEVC codec.
	CodecHEVC VideoCodec = "hevc"
	// CodecAV1 uses AV1 next-generation open codec.
	CodecAV1 VideoCodec = "av1"
)

// GPUType represents a specific hardware acceleration backend supported by FFmpeg.
type GPUType string

const (
	// GPU_NVENC uses NVIDIA's NVENC hardware encoder for H.264/HEVC/AV1.
	GPU_NVENC GPUType = "nvenc"
	// GPU_VAAPI uses the Video Acceleration API (Intel/AMD) for hardware encoding.
	GPU_VAAPI GPUType = "vaapi"
	// GPU_VIDEOTOOLBOX uses Apple's VideoToolbox framework for hardware encoding on macOS.
	GPU_VIDEOTOOLBOX GPUType = "videotoolbox"
)

// VOD is the default configuration for Video-On-Demand content.
var VOD = Profile{
	SegmentDuration: 5,
	LowLatency:      false,
	BFrames:         0,
}

// LIVE is the default configuration for low-latency Live streaming content.
var LIVE = Profile{
	SegmentDuration: 2,
	LowLatency:      true,
	BFrames:         0,
}
