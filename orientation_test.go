package mosaic

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/farshidrezaei/mosaic/internal/executor"
)

func TestParseOrientationProbeOutput(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		wantCodec    string
		wantRotation int
		wantErr      bool
	}{
		{
			name:         "side data rotation has priority",
			json:         `{"streams":[{"width":1920,"height":1080,"codec_name":"h264","tags":{"rotate":"270"},"side_data_list":[{"rotation":90}]}]}`,
			wantRotation: 90,
			wantCodec:    "h264",
		},
		{
			name:         "tag rotation fallback",
			json:         `{"streams":[{"width":1920,"height":1080,"codec_name":"hevc","tags":{"rotate":"-90"}}]}`,
			wantRotation: 270,
			wantCodec:    "hevc",
		},
		{
			name:         "broken metadata",
			json:         `{"streams":[{"width":1920,"height":1080,"codec_name":"h264","tags":{"rotate":"bad"},"side_data_list":[{"rotation":"bad"}]}]}`,
			wantRotation: 0,
			wantCodec:    "h264",
		},
		{
			name:    "no stream",
			json:    `{"streams":[]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOrientationProbeOutput([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseOrientationProbeOutput() err=%v wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Rotation != tt.wantRotation {
				t.Fatalf("Rotation=%d want %d", got.Rotation, tt.wantRotation)
			}
			if got.CodecName != tt.wantCodec {
				t.Fatalf("CodecName=%q want %q", got.CodecName, tt.wantCodec)
			}
		})
	}
}

func TestRotationFilter(t *testing.T) {
	tests := []struct {
		name       string
		wantFilter string
		rotation   int
		wantOK     bool
	}{
		{name: "90", rotation: 90, wantFilter: "transpose=1", wantOK: true},
		{name: "180", rotation: 180, wantFilter: "transpose=1,transpose=1", wantOK: true},
		{name: "270", rotation: 270, wantFilter: "transpose=2", wantOK: true},
		{name: "0", rotation: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilter, gotOK := rotationFilter(tt.rotation)
			if gotOK != tt.wantOK {
				t.Fatalf("ok=%v want %v", gotOK, tt.wantOK)
			}
			if gotFilter != tt.wantFilter {
				t.Fatalf("filter=%q want %q", gotFilter, tt.wantFilter)
			}
		})
	}
}

func TestNormalizeRotationWithExecutor_NoRotationRemuxes(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "in.mp4")
	outputPath := filepath.Join(dir, "out.mp4")
	if err := os.WriteFile(inputPath, []byte("plain"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	mock := &orientationMockExecutor{
		ffprobeOutputs: [][]byte{
			[]byte(`{"streams":[{"width":1280,"height":720,"codec_name":"h264"}]}`),
		},
		createFFmpegOutput: true,
	}

	if err := normalizeRotationWithExecutor(context.Background(), inputPath, outputPath, mock); err != nil {
		t.Fatalf("normalizeRotationWithExecutor() err=%v", err)
	}
	if mock.ffmpegCalls != 1 {
		t.Fatalf("expected 1 ffmpeg call, got %d", mock.ffmpegCalls)
	}
	assertContainsArg(t, mock.lastFFmpegArgs, "-c")
	assertContainsArg(t, mock.lastFFmpegArgs, "copy")
	assertContainsArg(t, mock.lastFFmpegArgs, "rotate=0")
}

func TestNormalizeRotationWithExecutor_NoRotationURLInput(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "out.mp4")

	mock := &orientationMockExecutor{
		ffprobeOutputs: [][]byte{
			[]byte(`{"streams":[{"width":1280,"height":720,"codec_name":"h264"}]}`),
		},
		createFFmpegOutput: true,
	}

	if err := normalizeRotationWithExecutor(context.Background(), "https://example.com/video.mp4", outputPath, mock); err != nil {
		t.Fatalf("normalizeRotationWithExecutor() err=%v", err)
	}
	if mock.ffmpegCalls != 1 {
		t.Fatalf("expected 1 ffmpeg call, got %d", mock.ffmpegCalls)
	}
}

func TestNormalizeRotationWithExecutor_Rotation90RunsFFmpeg(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "in.mp4")
	outputPath := filepath.Join(dir, "out.mp4")
	if err := os.WriteFile(inputPath, []byte("input"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	mock := &orientationMockExecutor{
		ffprobeOutputs: [][]byte{
			[]byte(`{"streams":[{"width":1920,"height":1080,"codec_name":"h264","side_data_list":[{"rotation":90}]}]}`),
			[]byte(`{"streams":[{"width":1080,"height":1920,"codec_name":"h264"}]}`),
		},
		createFFmpegOutput: true,
	}

	if err := normalizeRotationWithExecutor(context.Background(), inputPath, outputPath, mock); err != nil {
		t.Fatalf("normalizeRotationWithExecutor() err=%v", err)
	}
	if mock.ffmpegCalls != 1 {
		t.Fatalf("expected 1 ffmpeg call, got %d", mock.ffmpegCalls)
	}
	assertContainsArg(t, mock.lastFFmpegArgs, "-noautorotate")
	assertContainsArg(t, mock.lastFFmpegArgs, "transpose=1")
	assertContainsArg(t, mock.lastFFmpegArgs, "rotate=0")
}

func assertContainsArg(t *testing.T, args []string, want string) {
	t.Helper()
	for _, a := range args {
		if a == want {
			return
		}
	}
	t.Fatalf("args %v do not contain %q", args, want)
}

type orientationMockExecutor struct {
	ffprobeErr     error
	ffprobeOutputs [][]byte
	ffmpegErrors   []error
	lastFFmpegArgs []string

	createFFmpegOutput bool
	ffprobeCalls       int
	ffmpegCalls        int
}

func (m *orientationMockExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, *executor.Usage, error) {
	return m.ExecuteWithProgress(ctx, nil, name, args...)
}

func (m *orientationMockExecutor) ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, *executor.Usage, error) {
	if progress != nil {
		close(progress)
	}
	switch name {
	case "ffprobe":
		m.ffprobeCalls++
		if m.ffprobeErr != nil {
			return nil, nil, m.ffprobeErr
		}
		if m.ffprobeCalls <= len(m.ffprobeOutputs) {
			return m.ffprobeOutputs[m.ffprobeCalls-1], nil, nil
		}
		return []byte(`{"streams":[{"width":1920,"height":1080,"codec_name":"h264"}]}`), nil, nil
	case "ffmpeg":
		m.ffmpegCalls++
		m.lastFFmpegArgs = append([]string(nil), args...)
		if m.createFFmpegOutput && len(args) > 0 {
			_ = os.WriteFile(args[len(args)-1], []byte("normalized"), 0o644)
		}
		if len(m.ffmpegErrors) >= m.ffmpegCalls {
			return nil, nil, m.ffmpegErrors[m.ffmpegCalls-1]
		}
		return nil, nil, nil
	default:
		return nil, nil, errors.New("unexpected command: " + name + " " + strings.Join(args, " "))
	}
}
