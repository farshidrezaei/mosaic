package mosaic

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestIntegrationFFmpeg runs a real encoding job using FFmpeg.
// It is gated by the FFMPEG_INTEGRATION environment variable.
// To run: FFMPEG_INTEGRATION=1 go test -v integration_test.go
func TestIntegrationFFmpeg(t *testing.T) {
	if os.Getenv("FFMPEG_INTEGRATION") == "" {
		t.Skip("Skipping integration test (FFMPEG_INTEGRATION not set)")
	}

	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "mosaic-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}(tmpDir)

	// We need a real input file. For now, we'll assume there's a sample.mp4 in testdata
	// or we could generate one using FFmpeg.
	inputPath := filepath.Join(tmpDir, "input.mp4")

	// Generate a 1-second silent video for testing
	// ffmpeg -f lavfi -i color=c=blue:s=320x240:d=1 -f lavfi -i anullsrc=r=44100:cl=mono -t 1 -c:v libx264 -c:a aac input.mp4
	err = runCommand("ffmpeg", "-f", "lavfi", "-i", "color=c=blue:s=320x240:d=1", "-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", "1", "-c:v", "libx264", "-c:a", "aac", inputPath)
	if err != nil {
		t.Fatalf("Failed to generate test video: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")
	_ = os.MkdirAll(outputDir, 0755)

	job := Job{
		Input:     inputPath,
		OutputDir: outputDir,
		Profile:   ProfileVOD,
		ProgressHandler: func(info ProgressInfo) {
			t.Logf("Progress: %.1f%%, Time: %s", info.Percentage, info.CurrentTime)
		},
	}

	t.Run("HLS Encoding", func(t *testing.T) {
		_, err := EncodeHls(context.Background(), job, WithLogLevel("error"))
		if err != nil {
			t.Errorf("HLS Encoding failed: %v", err)
		}

		// Verify output files
		if _, err := os.Stat(filepath.Join(outputDir, "master.m3u8")); os.IsNotExist(err) {
			t.Error("master.m3u8 not found")
		}
	})

	t.Run("DASH Encoding", func(t *testing.T) {
		dashOutputDir := filepath.Join(tmpDir, "output_dash")
		_ = os.MkdirAll(dashOutputDir, 0755)
		job.OutputDir = dashOutputDir

		// Note: EncodeDash currently returns only an error, not a string like EncodeHls.
		// The change to `_, err :=` is made to align with the user's request,
		// assuming a future change to EncodeDash's signature or a desire for consistency.
		// The original `err := EncodeDash(...)` was already correctly handling the single error return.
		_, err := EncodeDash(context.Background(), job, WithLogLevel("error"))
		if err != nil {
			t.Errorf("DASH Encoding failed: %v", err)
		}

		// Verify output files
		if _, err := os.Stat(filepath.Join(dashOutputDir, "manifest.mpd")); os.IsNotExist(err) {
			t.Error("manifest.mpd not found")
		}
	})
}

func runCommand(name string, args ...string) error {
	// Simple wrapper for os/exec to avoid circular dependency on internal/executor
	// which might be overkill for this helper.
	cmd := exec.CommandContext(context.Background(), name, args...)
	return cmd.Run()
}
