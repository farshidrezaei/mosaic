// Package thumbnail provides video storyboard thumbnail sprite sheet and WebVTT cue generation.
//
// It extracts frames at regular intervals, tiles them into single or multi-page sprite sheets
// using FFmpeg, and generates standard WebVTT cue files compatible with modern HTML5 video players
// (such as Video.js, Plyr, and Shaka Player) for timeline scrubber previews.
package thumbnail

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/farshidrezaei/mosaic/internal/executor"
)

// DefaultConfig returns the recommended default thumbnail generation configuration.
var DefaultConfig = Config{
	SpriteFilename:  "thumbnails_%d.jpg",
	VTTFilename:     "thumbnails.vtt",
	IntervalSeconds: 5,
	TileWidth:       160,
	TileHeight:      90,
	Columns:         10,
	Rows:            10,
	Quality:         3,
}

// Config specifies thumbnail sprite generation parameters.
type Config struct {
	// SpriteFilename is the output filename pattern for the sprite image (e.g. "thumbnails_%d.jpg").
	SpriteFilename string
	// VTTFilename is the output filename for the WebVTT metadata file (default "thumbnails.vtt").
	VTTFilename string
	// IntervalSeconds is the interval in seconds between captured thumbnail frames (default 5).
	IntervalSeconds int
	// TileWidth is the width of each thumbnail frame in pixels (default 160).
	TileWidth int
	// TileHeight is the height of each thumbnail frame in pixels (default 90).
	TileHeight int
	// Columns is the number of thumbnail columns per sprite sheet (default 10).
	Columns int
	// Rows is the number of thumbnail rows per sprite sheet (default 10).
	Rows int
	// Quality is the JPEG quality factor passed to FFmpeg -q:v (1=highest, 31=lowest, default 3).
	Quality int
}

// Normalize validates and fills default values for unspecified configuration fields.
func (c *Config) Normalize() {
	if c.IntervalSeconds <= 0 {
		c.IntervalSeconds = 5
	}
	if c.TileWidth <= 0 {
		c.TileWidth = 160
	}
	if c.TileHeight <= 0 {
		c.TileHeight = 90
	}
	if c.Columns <= 0 {
		c.Columns = 10
	}
	if c.Rows <= 0 {
		c.Rows = 10
	}
	if c.Quality <= 0 {
		c.Quality = 3
	}
	if c.SpriteFilename == "" {
		c.SpriteFilename = "thumbnails_%d.jpg"
	}
	if c.VTTFilename == "" {
		c.VTTFilename = "thumbnails.vtt"
	}
}

// BuildFFmpegArgs constructs the FFmpeg argument list for thumbnail sprite generation.
func BuildFFmpegArgs(input string, outputPattern string, cfg Config) []string {
	cfg.Normalize()
	vf := fmt.Sprintf("fps=1/%d,scale=%d:%d,tile=%dx%d",
		cfg.IntervalSeconds, cfg.TileWidth, cfg.TileHeight, cfg.Columns, cfg.Rows)

	return []string{
		"-y",
		"-i", input,
		"-vf", vf,
		"-start_number", "0",
		"-q:v", strconv.Itoa(cfg.Quality),
		outputPattern,
	}
}

// FormatVTTTime formats seconds into WebVTT timestamp format (HH:MM:SS.mmm).
func FormatVTTTime(seconds float64) string {
	totalMs := int64(seconds * 1000)
	if totalMs < 0 {
		totalMs = 0
	}
	hours := totalMs / 3600000
	totalMs %= 3600000
	mins := totalMs / 60000
	totalMs %= 60000
	secs := totalMs / 1000
	ms := totalMs % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, mins, secs, ms)
}

// GenerateVTT creates the WebVTT cue content matching the tiled sprite sheets.
func GenerateVTT(duration float64, cfg Config) string {
	cfg.Normalize()
	if duration <= 0 {
		return "WEBVTT\n"
	}

	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")

	tilesPerSheet := cfg.Columns * cfg.Rows
	interval := float64(cfg.IntervalSeconds)
	totalFrames := int(duration / interval)
	if duration > 0 && float64(totalFrames)*interval < duration {
		totalFrames++
	}

	for i := 0; i < totalFrames; i++ {
		startTime := float64(i) * interval
		endTime := startTime + interval
		if endTime > duration {
			endTime = duration
		}

		sheetIndex := i / tilesPerSheet
		frameInSheet := i % tilesPerSheet

		col := frameInSheet % cfg.Columns
		row := frameInSheet / cfg.Columns

		x := col * cfg.TileWidth
		y := row * cfg.TileHeight

		var spriteName string
		if strings.Contains(cfg.SpriteFilename, "%") {
			spriteName = fmt.Sprintf(cfg.SpriteFilename, sheetIndex)
		} else {
			spriteName = cfg.SpriteFilename
		}

		// WebVTT cue line
		_, _ = fmt.Fprintf(&sb, "%s --> %s\n", FormatVTTTime(startTime), FormatVTTTime(endTime))
		_, _ = fmt.Fprintf(&sb, "%s#xywh=%d,%d,%d,%d\n\n", spriteName, x, y, cfg.TileWidth, cfg.TileHeight)
	}

	return sb.String()
}

// Generate creates thumbnail sprites and the corresponding WebVTT file using DefaultExecutor.
func Generate(ctx context.Context, input string, outputDir string, duration float64, cfg Config) error {
	return GenerateWithExecutor(ctx, input, outputDir, duration, cfg, executor.DefaultExecutor)
}

// GenerateWithExecutor creates thumbnail sprites and WebVTT using the specified CommandExecutor.
func GenerateWithExecutor(
	ctx context.Context,
	input string,
	outputDir string,
	duration float64,
	cfg Config,
	exec executor.CommandExecutor,
) error {
	cfg.Normalize()

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create thumbnail output directory: %w", err)
	}

	spriteOutputPath := filepath.Join(outputDir, cfg.SpriteFilename)
	args := BuildFFmpegArgs(input, spriteOutputPath, cfg)

	if _, _, err := exec.Execute(ctx, "ffmpeg", args...); err != nil {
		return fmt.Errorf("ffmpeg thumbnail generation failed: %w", err)
	}

	vttContent := GenerateVTT(duration, cfg)
	vttPath := filepath.Join(outputDir, cfg.VTTFilename)

	if err := os.WriteFile(vttPath, []byte(vttContent), 0o644); err != nil {
		return fmt.Errorf("write thumbnails.vtt: %w", err)
	}

	return nil
}
