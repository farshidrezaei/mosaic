package mosaic

// Profile represents an encoding profile that determines segment duration and latency settings.
type Profile string

const (
	// ProfileVOD is optimized for on-demand content with 5-second segments for better compression.
	ProfileVOD Profile = "vod"
	// ProfileLive is optimized for live streaming with 2-second segments and low-latency features.
	ProfileLive Profile = "live"
)

// ProgressInfo contains real-time information about the current encoding progress.
type ProgressInfo struct {
	// CurrentTime is the current timestamp in the video being processed (e.g., "00:01:23.45").
	CurrentTime string
	// Bitrate is the current encoding bitrate (e.g., "2500kbits/s").
	Bitrate string
	// Speed is the current encoding speed relative to real-time (e.g., "1.5x").
	Speed string
	// Percentage is the estimated completion percentage (0.0 to 100.0).
	Percentage float64
}

// ProgressHandler is a callback function that receives ProgressInfo updates during encoding.
type ProgressHandler func(ProgressInfo)

// Job defines the parameters and configuration for an adaptive bitrate encoding task.
type Job struct {
	// Input is the absolute path or public URL to the source video file.
	Input string
	// OutputDir is the directory where generated segments, playlists, and manifests will be stored.
	OutputDir string
	// ProgressHandler is an optional callback to monitor encoding progress in real-time.
	ProgressHandler ProgressHandler
	// Profile determines the segment duration and latency characteristics of the output.
	Profile Profile
}
