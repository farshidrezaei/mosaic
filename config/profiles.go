package config

// Profile defines the configuration for an encoding profile.
type Profile struct {
	// SegmentDuration is the duration of each media segment in seconds.
	SegmentDuration int
	// LowLatency indicates if low-latency features should be enabled.
	LowLatency bool
}

// VOD is the configuration for on-demand content.
var VOD = Profile{
	SegmentDuration: 5,
	LowLatency:      false,
}

// LIVE is the configuration for live streaming content.
var LIVE = Profile{
	SegmentDuration: 2,
	LowLatency:      true,
}
