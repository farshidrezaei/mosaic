package thumbnail

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/farshidrezaei/mosaic/internal/executor"
)

func TestNormalize(t *testing.T) {
	cfg := Config{}
	cfg.Normalize()

	if cfg.IntervalSeconds != 5 {
		t.Errorf("expected IntervalSeconds 5, got %d", cfg.IntervalSeconds)
	}
	if cfg.TileWidth != 160 {
		t.Errorf("expected TileWidth 160, got %d", cfg.TileWidth)
	}
	if cfg.TileHeight != 90 {
		t.Errorf("expected TileHeight 90, got %d", cfg.TileHeight)
	}
	if cfg.Columns != 10 {
		t.Errorf("expected Columns 10, got %d", cfg.Columns)
	}
	if cfg.Rows != 10 {
		t.Errorf("expected Rows 10, got %d", cfg.Rows)
	}
	if cfg.Quality != 3 {
		t.Errorf("expected Quality 3, got %d", cfg.Quality)
	}
	if cfg.SpriteFilename != "thumbnails_%d.jpg" {
		t.Errorf("expected SpriteFilename thumbnails_%%d.jpg, got %s", cfg.SpriteFilename)
	}
	if cfg.VTTFilename != "thumbnails.vtt" {
		t.Errorf("expected VTTFilename thumbnails.vtt, got %s", cfg.VTTFilename)
	}
}

func TestBuildFFmpegArgs(t *testing.T) {
	cfg := Config{
		IntervalSeconds: 10,
		TileWidth:       200,
		TileHeight:      112,
		Columns:         5,
		Rows:            5,
		Quality:         2,
	}

	args := BuildFFmpegArgs("input.mp4", "/output/thumbnails_%d.jpg", cfg)

	expectedVF := "fps=1/10,scale=200:112,tile=5x5"
	foundVF := false
	for i, arg := range args {
		if arg == "-vf" && i+1 < len(args) && args[i+1] == expectedVF {
			foundVF = true
			break
		}
	}

	if !foundVF {
		t.Fatalf("expected -vf %q in args: %v", expectedVF, args)
	}

	if args[len(args)-1] != "/output/thumbnails_%d.jpg" {
		t.Fatalf("expected output path at end, got %s", args[len(args)-1])
	}
}

func TestFormatVTTTime(t *testing.T) {
	tests := []struct {
		expected string
		seconds  float64
	}{
		{"00:00:00.000", 0},
		{"00:00:05.000", 5},
		{"00:01:05.432", 65.432},
		{"01:01:01.123", 3661.123},
		{"00:00:00.000", -10},
	}

	for _, tt := range tests {
		got := FormatVTTTime(tt.seconds)
		if got != tt.expected {
			t.Errorf("FormatVTTTime(%v) = %s; want %s", tt.seconds, got, tt.expected)
		}
	}
}

func TestGenerateVTT(t *testing.T) {
	cfg := Config{
		IntervalSeconds: 5,
		TileWidth:       160,
		TileHeight:      90,
		Columns:         2,
		Rows:            2,
		SpriteFilename:  "thumb_%d.jpg",
	}

	// 2x2 = 4 tiles per sheet. Duration 22s -> 5 cues (0-5, 5-10, 10-15, 15-20, 20-22)
	// Frame 0: sheet 0, col 0, row 0 -> 0,0
	// Frame 1: sheet 0, col 1, row 0 -> 160,0
	// Frame 2: sheet 0, col 0, row 1 -> 0,90
	// Frame 3: sheet 0, col 1, row 1 -> 160,90
	// Frame 4: sheet 1, col 0, row 0 -> 0,0
	vtt := GenerateVTT(22, cfg)

	if !strings.HasPrefix(vtt, "WEBVTT\n\n") {
		t.Fatalf("expected WEBVTT header, got:\n%s", vtt)
	}

	expectedCues := []string{
		"00:00:00.000 --> 00:00:05.000\nthumb_0.jpg#xywh=0,0,160,90",
		"00:00:05.000 --> 00:00:10.000\nthumb_0.jpg#xywh=160,0,160,90",
		"00:00:10.000 --> 00:00:15.000\nthumb_0.jpg#xywh=0,90,160,90",
		"00:00:15.000 --> 00:00:20.000\nthumb_0.jpg#xywh=160,90,160,90",
		"00:00:20.000 --> 00:00:22.000\nthumb_1.jpg#xywh=0,0,160,90",
	}

	for _, cue := range expectedCues {
		if !strings.Contains(vtt, cue) {
			t.Errorf("missing expected cue:\n%s\nin vtt output:\n%s", cue, vtt)
		}
	}
}

func TestGenerateWithExecutor_Success(t *testing.T) {
	mockExec := executor.NewMockExecutor()
	mockExec.Responses["ffmpeg"] = executor.MockResponse{
		Output: []byte("frame=50 fps=0.0 q=3.0 size=N/A time=00:00:25.00"),
	}

	tmpDir := t.TempDir()
	cfg := Config{
		IntervalSeconds: 5,
		TileWidth:       160,
		TileHeight:      90,
		Columns:         5,
		Rows:            5,
	}

	err := GenerateWithExecutor(context.Background(), "input.mp4", tmpDir, 25.0, cfg, mockExec)
	if err != nil {
		t.Fatalf("GenerateWithExecutor failed: %v", err)
	}

	if mockExec.GetCallCount("ffmpeg") != 1 {
		t.Errorf("expected 1 ffmpeg call, got %d", mockExec.GetCallCount("ffmpeg"))
	}

	vttPath := filepath.Join(tmpDir, "thumbnails.vtt")
	data, err := os.ReadFile(vttPath)
	if err != nil {
		t.Fatalf("failed to read generated VTT file: %v", err)
	}

	if !strings.Contains(string(data), "WEBVTT") {
		t.Errorf("invalid VTT content: %s", string(data))
	}
}

func TestGenerateWithExecutor_FFmpegError(t *testing.T) {
	mockExec := executor.NewMockExecutor()
	mockExec.Responses["ffmpeg"] = executor.MockResponse{
		Err: &executor.CommandError{
			Command: "ffmpeg",
			Stderr:  "Invalid data found when processing input",
		},
	}

	tmpDir := t.TempDir()
	err := GenerateWithExecutor(context.Background(), "corrupted.mp4", tmpDir, 10.0, DefaultConfig, mockExec)
	if err == nil {
		t.Fatal("expected error from ffmpeg failure, got nil")
	}

	if !strings.Contains(err.Error(), "ffmpeg thumbnail generation failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}
