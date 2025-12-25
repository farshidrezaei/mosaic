package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/farshidrezaei/mosaic/internal/executor"
)

// VideoInfo contains technical metadata about a video file extracted via ffprobe.
type VideoInfo struct {
	// Width is the horizontal resolution in pixels.
	Width int
	// Height is the vertical resolution in pixels.
	Height int
	// FPS is the average frame rate of the video (e.g., 23.976, 30.0, 60.0).
	FPS float64
	// HasAudio is true if the video file contains at least one audio stream.
	HasAudio bool
}

// Input returns technical metadata for the given video file or URL.
// It uses the default command executor to run ffprobe.
func Input(ctx context.Context, input string) (VideoInfo, error) {
	return InputWithExecutor(ctx, input, executor.DefaultExecutor)
}

// InputWithExecutor is like Input but allows providing a custom CommandExecutor.
func InputWithExecutor(ctx context.Context, input string, exec executor.CommandExecutor) (VideoInfo, error) {
	// Probe video stream
	args := []string{
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,avg_frame_rate",
		"-of", "json",
		input,
	}
	out, err := exec.Execute(ctx, "ffprobe", args...)
	if err != nil {
		return VideoInfo{}, err
	}

	var data struct {
		Streams []struct {
			FPS    string `json:"avg_frame_rate"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"streams"`
	}

	err = json.Unmarshal(out, &data)
	if err != nil {
		return VideoInfo{}, err
	}
	if len(data.Streams) == 0 {
		return VideoInfo{}, fmt.Errorf("no video stream found")
	}

	info := VideoInfo{
		Width:  data.Streams[0].Width,
		Height: data.Streams[0].Height,
		FPS:    parseFPS(data.Streams[0].FPS),
	}

	// audio check
	aout, err := exec.Execute(
		ctx,
		"ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		input,
	)
	_ = err // Ignore audio probe errors
	info.HasAudio = strings.TrimSpace(string(aout)) != ""

	return info, nil
}

func parseFPS(rate string) float64 {
	parts := strings.Split(rate, "/")
	if len(parts) != 2 {
		return 30
	}
	n, _ := strconv.ParseFloat(parts[0], 64)
	d, _ := strconv.ParseFloat(parts[1], 64)
	if d == 0 {
		return 30
	}
	return n / d
}
