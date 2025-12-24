package mosaic

// Profile represents an encoding profile (e.g., VOD or Live).
type Profile string

const (
	// ProfileVOD is the profile for on-demand content with 5-second segments.
	ProfileVOD Profile = "vod"
	// ProfileLive is the profile for live content with 2-second segments.
	ProfileLive Profile = "live"
)

// Job defines the parameters for an encoding task.
type Job struct {
	// Input is the path or URL to the source video file.
	Input string
	// OutputDir is the directory where the generated segments and manifests will be saved.
	OutputDir string
	// Profile is the encoding profile to use (VOD or Live).
	Profile Profile
}
