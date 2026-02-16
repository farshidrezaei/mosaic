package mosaic

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/farshidrezaei/mosaic/internal/executor"
)

type orientationProbeResponse struct {
	Streams []orientationProbeStream `json:"streams"`
}

type orientationProbeStream struct {
	CodecName    string                     `json:"codec_name"`
	Tags         map[string]string          `json:"tags"`
	SideDataList []orientationProbeSideData `json:"side_data_list"`
	Width        int                        `json:"width"`
	Height       int                        `json:"height"`
}

type orientationProbeSideData struct {
	Rotation interface{} `json:"rotation"`
}

type orientationMetadata struct {
	CodecName string
	Width     int
	Height    int
	Rotation  int
}

// NormalizeVideoOrientation normalizes source orientation by physically rotating
// frames for 90/180/270 metadata-based rotations and clearing rotate metadata.
func NormalizeVideoOrientation(ctx context.Context, inputPath, outputPath string) error {
	return normalizeRotationWithExecutor(ctx, inputPath, outputPath, executor.DefaultExecutor)
}

func normalizeRotationWithExecutor(
	ctx context.Context,
	inputPath, outputPath string,
	exec executor.CommandExecutor,
) error {
	if strings.TrimSpace(inputPath) == "" {
		return fmt.Errorf("input path is required")
	}
	if strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("output path is required")
	}

	meta, err := probeOrientationMetadata(ctx, inputPath, exec)
	if err != nil {
		return err
	}

	filter, shouldRotate := rotationFilter(meta.Rotation)
	if !shouldRotate {
		tmpOutput, cleanup, prepErr := prepareTempOutput(outputPath)
		if prepErr != nil {
			return prepErr
		}
		defer cleanup()

		args := buildRemuxFFmpegArgs(inputPath, tmpOutput)
		if _, _, execErr := exec.Execute(ctx, "ffmpeg", args...); execErr != nil {
			return fmt.Errorf("normalize orientation: ffmpeg remux failed: %w", execErr)
		}
		if renameErr := os.Rename(tmpOutput, outputPath); renameErr != nil {
			return fmt.Errorf("finalize remux output: %w", renameErr)
		}
		return nil
	}

	tmpOutput, cleanup, err := prepareTempOutput(outputPath)
	if err != nil {
		return err
	}
	defer cleanup()

	enc := preferredVideoEncoder(meta.CodecName)
	args := buildRotateFFmpegArgs(inputPath, tmpOutput, enc, filter)
	_, _, err = exec.Execute(ctx, "ffmpeg", args...)
	if err != nil && enc != "libx264" {
		args = buildRotateFFmpegArgs(inputPath, tmpOutput, "libx264", filter)
		_, _, err = exec.Execute(ctx, "ffmpeg", args...)
	}
	if err != nil {
		return fmt.Errorf("normalize orientation: ffmpeg failed: %w", err)
	}

	outMeta, err := probeOrientationMetadata(ctx, tmpOutput, exec)
	if err != nil {
		return fmt.Errorf("verify normalized output: %w", err)
	}
	if outMeta.Rotation != 0 {
		return fmt.Errorf("verify normalized output: rotate metadata still present (%d)", outMeta.Rotation)
	}

	if err := os.Rename(tmpOutput, outputPath); err != nil {
		return fmt.Errorf("finalize normalized output: %w", err)
	}
	return nil
}

func probeOrientationMetadata(
	ctx context.Context,
	inputPath string,
	exec executor.CommandExecutor,
) (orientationMetadata, error) {
	args := []string{
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,codec_name:stream_tags=rotate:stream_side_data=rotation",
		"-of", "json",
		inputPath,
	}
	out, _, err := exec.Execute(ctx, "ffprobe", args...)
	if err != nil {
		return orientationMetadata{}, fmt.Errorf("ffprobe orientation probe failed: %w", err)
	}
	return parseOrientationProbeOutput(out)
}

func parseOrientationProbeOutput(data []byte) (orientationMetadata, error) {
	var resp orientationProbeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return orientationMetadata{}, fmt.Errorf("parse ffprobe json: %w", err)
	}
	if len(resp.Streams) == 0 {
		return orientationMetadata{}, fmt.Errorf("no video stream found")
	}

	s := resp.Streams[0]
	return orientationMetadata{
		Width:     s.Width,
		Height:    s.Height,
		CodecName: s.CodecName,
		Rotation:  detectOrientationRotation(s),
	}, nil
}

func detectOrientationRotation(stream orientationProbeStream) int {
	for _, sideData := range stream.SideDataList {
		if r, ok := parseRotationValue(sideData.Rotation); ok {
			return normalizeRotationDegrees(r)
		}
	}
	if stream.Tags != nil {
		if raw, ok := stream.Tags["rotate"]; ok {
			if r, ok := parseRotationValue(raw); ok {
				return normalizeRotationDegrees(r)
			}
		}
	}
	return 0
}

func parseRotationValue(v interface{}) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(math.Round(x)), true
	case string:
		trimmed := strings.TrimSpace(x)
		if trimmed == "" {
			return 0, false
		}
		if i, err := strconv.Atoi(trimmed); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return int(math.Round(f)), true
		}
		return 0, false
	default:
		return 0, false
	}
}

func normalizeRotationDegrees(rotation int) int {
	rotation %= 360
	if rotation < 0 {
		rotation += 360
	}
	return rotation
}

func rotationFilter(rotation int) (string, bool) {
	switch normalizeRotationDegrees(rotation) {
	case 90:
		return "transpose=1", true
	case 180:
		return "transpose=1,transpose=1", true
	case 270:
		return "transpose=2", true
	default:
		return "", false
	}
}

func preferredVideoEncoder(codecName string) string {
	switch strings.ToLower(strings.TrimSpace(codecName)) {
	case "h264":
		return "libx264"
	case "hevc":
		return "libx265"
	case "vp8":
		return "libvpx"
	case "vp9":
		return "libvpx-vp9"
	case "av1":
		return "libaom-av1"
	case "mpeg4", "mjpeg", "prores", "dnxhd":
		return strings.ToLower(strings.TrimSpace(codecName))
	default:
		return "libx264"
	}
}

func buildRotateFFmpegArgs(inputPath, outputPath, encoderName, filter string) []string {
	return []string{
		"-y",
		"-v", "error",
		"-noautorotate",
		"-i", inputPath,
		"-map", "0:v:0",
		"-map", "0:a?",
		"-vf", filter,
		"-c:v", encoderName,
		"-c:a", "copy",
		"-metadata:s:v:0", "rotate=0",
		outputPath,
	}
}

func buildRemuxFFmpegArgs(inputPath, outputPath string) []string {
	return []string{
		"-y",
		"-v", "error",
		"-i", inputPath,
		"-c", "copy",
		"-metadata:s:v:0", "rotate=0",
		outputPath,
	}
}

func prepareTempOutput(outputPath string) (string, func(), error) {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, fmt.Errorf("create output dir: %w", err)
	}

	ext := filepath.Ext(outputPath)
	pattern := ".mosaic-orientation-*"
	if ext != "" {
		pattern += ext
	}

	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", nil, fmt.Errorf("create temp output: %w", err)
	}
	tmpPath := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", nil, fmt.Errorf("close temp output: %w", err)
	}
	return tmpPath, func() { _ = os.Remove(tmpPath) }, nil
}
