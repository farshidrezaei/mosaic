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
	CodecType    string                     `json:"codec_type"`
	Tags         map[string]string          `json:"tags"`
	SideDataList []orientationProbeSideData `json:"side_data_list"`
	Width        int                        `json:"width"`
	Height       int                        `json:"height"`
}

type orientationProbeSideData struct {
	Rotation interface{} `json:"rotation"`
}

type orientationMetadata struct {
	CodecName      string
	AudioCodecName string
	Width          int
	Height         int
	Rotation       int
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

		videoCodec, audioCodec := "copy", "copy"
		if !outputAcceptsCodec(outputPath, meta.CodecName, mp4CopyVideoCodecs) {
			videoCodec = "libx264"
		}
		if !outputAcceptsCodec(outputPath, meta.AudioCodecName, mp4CopyAudioCodecs) {
			audioCodec = "aac"
		}
		args := buildRemuxFFmpegArgs(inputPath, tmpOutput, videoCodec, audioCodec)
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

	enc := "libx264"
	if outputAcceptsCodec(outputPath, meta.CodecName, mp4CopyVideoCodecs) {
		enc = preferredVideoEncoder(meta.CodecName)
	}
	audioCodec := "copy"
	if !outputAcceptsCodec(outputPath, meta.AudioCodecName, mp4CopyAudioCodecs) {
		audioCodec = "aac"
	}
	args := buildRotateFFmpegArgs(inputPath, tmpOutput, enc, filter, audioCodec)
	_, _, err = exec.Execute(ctx, "ffmpeg", args...)
	if err != nil && enc != "libx264" {
		args = buildRotateFFmpegArgs(inputPath, tmpOutput, "libx264", filter, audioCodec)
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
		"-show_entries", "stream=codec_type,width,height,codec_name:stream_tags=rotate:stream_side_data=rotation",
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
	var video *orientationProbeStream
	audioCodec := ""
	for i := range resp.Streams {
		s := &resp.Streams[i]
		switch s.CodecType {
		case "video", "":
			if video == nil {
				video = s
			}
		case "audio":
			if audioCodec == "" {
				audioCodec = s.CodecName
			}
		}
	}
	if video == nil {
		return orientationMetadata{}, fmt.Errorf("no video stream found")
	}

	return orientationMetadata{
		Width:          video.Width,
		Height:         video.Height,
		CodecName:      video.CodecName,
		AudioCodecName: audioCodec,
		Rotation:       detectOrientationRotation(*video),
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
		return "transpose=2", true
	case 180:
		return "transpose=1,transpose=1", true
	case 270:
		return "transpose=1", true
	default:
		return "", false
	}
}

// mp4CopyVideoCodecs and mp4CopyAudioCodecs list the codecs the MP4 muxer
// accepts with stream copy. Other codecs (vp8, theora, msmpeg4, wmv, vorbis, ...)
// are transcoded to H.264/AAC when the normalized output is an MP4 file.
var (
	mp4CopyVideoCodecs = map[string]bool{
		"h264": true, "hevc": true, "av1": true, "vp9": true,
		"mpeg4": true, "mpeg1video": true, "mpeg2video": true, "mjpeg": true,
	}
	mp4CopyAudioCodecs = map[string]bool{
		"aac": true, "mp3": true, "ac3": true, "eac3": true, "alac": true, "opus": true,
	}
)

// outputAcceptsCodec reports whether codecName can be stream-copied into
// outputPath. Only MP4 outputs are checked; other containers mirror the input
// container, so their codecs are kept as before.
func outputAcceptsCodec(outputPath, codecName string, mp4Codecs map[string]bool) bool {
	codec := strings.ToLower(strings.TrimSpace(codecName))
	if codec == "" {
		return true
	}
	switch strings.ToLower(filepath.Ext(outputPath)) {
	case ".mp4", ".m4v":
		return mp4Codecs[codec]
	default:
		return true
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

func buildRotateFFmpegArgs(inputPath, outputPath, encoderName, filter, audioCodec string) []string {
	return []string{
		"-y",
		"-v", "error",
		"-noautorotate",
		"-display_rotation:v:0", "0",
		"-i", inputPath,
		"-map", "0:v:0",
		"-map", "0:a:0?",
		"-dn",
		"-sn",
		"-map_chapters", "-1",
		"-vf", filter,
		"-c:v", encoderName,
		"-c:a", audioCodec,
		"-metadata:s:v:0", "rotate=0",
		outputPath,
	}
}

// buildRemuxFFmpegArgs stream-copies when videoCodec/audioCodec are "copy" and
// transcodes otherwise.
func buildRemuxFFmpegArgs(inputPath, outputPath, videoCodec, audioCodec string) []string {
	args := []string{
		"-y",
		"-v", "error",
		"-display_rotation:v:0", "0",
		"-i", inputPath,
		"-map", "0:v:0",
		"-map", "0:a:0?",
		"-dn",
		"-sn",
		"-map_chapters", "-1",
		"-c:v", videoCodec,
	}
	if videoCodec != "copy" {
		args = append(args, "-pix_fmt", "yuv420p")
	}
	return append(args,
		"-c:a", audioCodec,
		"-metadata:s:v:0", "rotate=0",
		outputPath,
	)
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
