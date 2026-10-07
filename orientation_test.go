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
		wantAudio    string
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
			name:      "audio codec from first audio stream",
			json:      `{"streams":[{"codec_type":"audio","codec_name":"vorbis"},{"codec_type":"video","width":1280,"height":720,"codec_name":"vp8"}]}`,
			wantCodec: "vp8",
			wantAudio: "vorbis",
		},
		{
			name:    "no stream",
			json:    `{"streams":[]}`,
			wantErr: true,
		},
		{
			name:    "audio only",
			json:    `{"streams":[{"codec_type":"audio","codec_name":"aac"}]}`,
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
			if got.AudioCodecName != tt.wantAudio {
				t.Fatalf("AudioCodecName=%q want %q", got.AudioCodecName, tt.wantAudio)
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
		{name: "90", rotation: 90, wantFilter: "transpose=2", wantOK: true},
		{name: "180", rotation: 180, wantFilter: "transpose=1,transpose=1", wantOK: true},
		{name: "270", rotation: 270, wantFilter: "transpose=1", wantOK: true},
		{name: "-90", rotation: -90, wantFilter: "transpose=1", wantOK: true},
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
	assertArgValue(t, mock.lastFFmpegArgs, "-c:v", "copy")
	assertArgValue(t, mock.lastFFmpegArgs, "-c:a", "copy")
	assertContainsArg(t, mock.lastFFmpegArgs, "0:v:0")
	assertContainsArg(t, mock.lastFFmpegArgs, "0:a:0?")
	assertContainsArg(t, mock.lastFFmpegArgs, "-dn")
	assertContainsArg(t, mock.lastFFmpegArgs, "-sn")
	assertContainsArg(t, mock.lastFFmpegArgs, "-map_chapters")
	assertContainsArg(t, mock.lastFFmpegArgs, "-1")
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
	assertContainsArg(t, mock.lastFFmpegArgs, "0:v:0")
	assertContainsArg(t, mock.lastFFmpegArgs, "0:a:0?")
	assertContainsArg(t, mock.lastFFmpegArgs, "-dn")
	assertContainsArg(t, mock.lastFFmpegArgs, "-sn")
	assertContainsArg(t, mock.lastFFmpegArgs, "-map_chapters")
	assertContainsArg(t, mock.lastFFmpegArgs, "-1")
	assertContainsArg(t, mock.lastFFmpegArgs, "transpose=2")
	assertContainsArg(t, mock.lastFFmpegArgs, "rotate=0")
}

func TestNormalizeRotationWithExecutor_RotationNegative90RunsFFmpeg(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "in.mp4")
	outputPath := filepath.Join(dir, "out.mp4")
	if err := os.WriteFile(inputPath, []byte("input"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	mock := &orientationMockExecutor{
		ffprobeOutputs: [][]byte{
			[]byte(`{"streams":[{"width":1920,"height":1080,"codec_name":"h264","side_data_list":[{"rotation":-90}]}]}`),
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

func TestOutputAcceptsCodec(t *testing.T) {
	tests := []struct {
		codecs map[string]bool
		output string
		codec  string
		want   bool
	}{
		{output: "out.mp4", codec: "h264", codecs: mp4CopyVideoCodecs, want: true},
		{output: "out.mp4", codec: "HEVC", codecs: mp4CopyVideoCodecs, want: true},
		{output: "out.mp4", codec: "vp8", codecs: mp4CopyVideoCodecs, want: false},
		{output: "out.mp4", codec: "theora", codecs: mp4CopyVideoCodecs, want: false},
		{output: "out.m4v", codec: "msmpeg4v3", codecs: mp4CopyVideoCodecs, want: false},
		{output: "out.mp4", codec: "aac", codecs: mp4CopyAudioCodecs, want: true},
		{output: "out.mp4", codec: "vorbis", codecs: mp4CopyAudioCodecs, want: false},
		{output: "out.mp4", codec: "wmav2", codecs: mp4CopyAudioCodecs, want: false},
		{output: "out.mp4", codec: "", codecs: mp4CopyAudioCodecs, want: true},
		{output: "out.webm", codec: "vp8", codecs: mp4CopyVideoCodecs, want: true},
		{output: "out.mkv", codec: "vorbis", codecs: mp4CopyAudioCodecs, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.output+"/"+tt.codec, func(t *testing.T) {
			if got := outputAcceptsCodec(tt.output, tt.codec, tt.codecs); got != tt.want {
				t.Fatalf("outputAcceptsCodec(%q, %q)=%v want %v", tt.output, tt.codec, got, tt.want)
			}
		})
	}
}

func TestNormalizeRotationWithExecutor_NoRotationCodecs(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		probe     string
		wantVideo string
		wantAudio string
	}{
		{
			name:      "h264 aac copies",
			output:    "out.mp4",
			probe:     `{"streams":[{"codec_type":"video","codec_name":"h264"},{"codec_type":"audio","codec_name":"aac"}]}`,
			wantVideo: "copy",
			wantAudio: "copy",
		},
		{
			name:      "vp8 vorbis transcodes",
			output:    "out.mp4",
			probe:     `{"streams":[{"codec_type":"video","codec_name":"vp8"},{"codec_type":"audio","codec_name":"vorbis"}]}`,
			wantVideo: "libx264",
			wantAudio: "aac",
		},
		{
			name:      "theora vorbis transcodes",
			output:    "out.mp4",
			probe:     `{"streams":[{"codec_type":"video","codec_name":"theora"},{"codec_type":"audio","codec_name":"vorbis"}]}`,
			wantVideo: "libx264",
			wantAudio: "aac",
		},
		{
			name:      "msmpeg4v3 wmav2 transcodes",
			output:    "out.mp4",
			probe:     `{"streams":[{"codec_type":"video","codec_name":"msmpeg4v3"},{"codec_type":"audio","codec_name":"wmav2"}]}`,
			wantVideo: "libx264",
			wantAudio: "aac",
		},
		{
			name:      "h264 vorbis transcodes audio only",
			output:    "out.mp4",
			probe:     `{"streams":[{"codec_type":"video","codec_name":"h264"},{"codec_type":"audio","codec_name":"vorbis"}]}`,
			wantVideo: "copy",
			wantAudio: "aac",
		},
		{
			name:      "vp8 without audio",
			output:    "out.mp4",
			probe:     `{"streams":[{"codec_type":"video","codec_name":"vp8"}]}`,
			wantVideo: "libx264",
			wantAudio: "copy",
		},
		{
			name:      "webm output keeps copy",
			output:    "out.webm",
			probe:     `{"streams":[{"codec_type":"video","codec_name":"vp8"},{"codec_type":"audio","codec_name":"vorbis"}]}`,
			wantVideo: "copy",
			wantAudio: "copy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(t.TempDir(), tt.output)
			mock := &orientationMockExecutor{
				ffprobeOutputs:     [][]byte{[]byte(tt.probe)},
				createFFmpegOutput: true,
			}

			if err := normalizeRotationWithExecutor(context.Background(), "in", outputPath, mock); err != nil {
				t.Fatalf("normalizeRotationWithExecutor() err=%v", err)
			}
			if mock.ffmpegCalls != 1 {
				t.Fatalf("expected 1 ffmpeg call, got %d", mock.ffmpegCalls)
			}
			assertArgValue(t, mock.lastFFmpegArgs, "-c:v", tt.wantVideo)
			assertArgValue(t, mock.lastFFmpegArgs, "-c:a", tt.wantAudio)
			if tt.wantVideo == "copy" {
				assertNotContainsArg(t, mock.lastFFmpegArgs, "-pix_fmt")
			} else {
				assertArgValue(t, mock.lastFFmpegArgs, "-pix_fmt", "yuv420p")
			}
		})
	}
}

func TestNormalizeRotationWithExecutor_RotationCodecs(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		video     string
		audio     string
		wantVideo string
		wantAudio string
	}{
		{name: "h264 aac", output: "out.mp4", video: "h264", audio: "aac", wantVideo: "libx264", wantAudio: "copy"},
		{name: "vp9 opus", output: "out.mp4", video: "vp9", audio: "opus", wantVideo: "libvpx-vp9", wantAudio: "copy"},
		{name: "vp8 vorbis", output: "out.mp4", video: "vp8", audio: "vorbis", wantVideo: "libx264", wantAudio: "aac"},
		{name: "theora vorbis", output: "out.mp4", video: "theora", audio: "vorbis", wantVideo: "libx264", wantAudio: "aac"},
		{name: "msmpeg4v3 wmav2", output: "out.mp4", video: "msmpeg4v3", audio: "wmav2", wantVideo: "libx264", wantAudio: "aac"},
		{name: "vp8 vorbis webm", output: "out.webm", video: "vp8", audio: "vorbis", wantVideo: "libvpx", wantAudio: "copy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(t.TempDir(), tt.output)
			mock := &orientationMockExecutor{
				ffprobeOutputs: [][]byte{
					[]byte(`{"streams":[{"codec_type":"video","width":1920,"height":1080,"codec_name":"` + tt.video + `","side_data_list":[{"rotation":90}]},{"codec_type":"audio","codec_name":"` + tt.audio + `"}]}`),
					[]byte(`{"streams":[{"codec_type":"video","width":1080,"height":1920,"codec_name":"h264"}]}`),
				},
				createFFmpegOutput: true,
			}

			if err := normalizeRotationWithExecutor(context.Background(), "in", outputPath, mock); err != nil {
				t.Fatalf("normalizeRotationWithExecutor() err=%v", err)
			}
			if mock.ffmpegCalls != 1 {
				t.Fatalf("expected 1 ffmpeg call, got %d", mock.ffmpegCalls)
			}
			assertArgValue(t, mock.lastFFmpegArgs, "-c:v", tt.wantVideo)
			assertArgValue(t, mock.lastFFmpegArgs, "-c:a", tt.wantAudio)
			assertContainsArg(t, mock.lastFFmpegArgs, "transpose=2")
		})
	}
}

func TestNormalizeRotationWithExecutor_RotationFallbackKeepsAudioCodec(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "out.mp4")
	mock := &orientationMockExecutor{
		ffprobeOutputs: [][]byte{
			[]byte(`{"streams":[{"codec_type":"video","codec_name":"av1","side_data_list":[{"rotation":90}]},{"codec_type":"audio","codec_name":"vorbis"}]}`),
			[]byte(`{"streams":[{"codec_type":"video","codec_name":"h264"}]}`),
		},
		ffmpegErrors:       []error{errors.New("unknown encoder")},
		createFFmpegOutput: true,
	}

	if err := normalizeRotationWithExecutor(context.Background(), "in", outputPath, mock); err != nil {
		t.Fatalf("normalizeRotationWithExecutor() err=%v", err)
	}
	if mock.ffmpegCalls != 2 {
		t.Fatalf("expected 2 ffmpeg calls, got %d", mock.ffmpegCalls)
	}
	assertArgValue(t, mock.lastFFmpegArgs, "-c:v", "libx264")
	assertArgValue(t, mock.lastFFmpegArgs, "-c:a", "aac")
}

func assertArgValue(t *testing.T, args []string, flag, want string) {
	t.Helper()
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			if args[i+1] != want {
				t.Fatalf("args %v: %s=%q want %q", args, flag, args[i+1], want)
			}
			return
		}
	}
	t.Fatalf("args %v do not contain %q", args, flag)
}

func assertNotContainsArg(t *testing.T, args []string, unwanted string) {
	t.Helper()
	for _, a := range args {
		if a == unwanted {
			t.Fatalf("args %v contain %q", args, unwanted)
		}
	}
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
