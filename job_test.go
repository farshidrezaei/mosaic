package mosaic

import "testing"

func TestProfileConstants(t *testing.T) {
	if ProfileVOD != "vod" {
		t.Errorf("expected ProfileVOD to be 'vod', got '%s'", ProfileVOD)
	}
	if ProfileLive != "live" {
		t.Errorf("expected ProfileLive to be 'live', got '%s'", ProfileLive)
	}
}

func TestJobStruct(t *testing.T) {
	job := Job{
		Input:     "/path/to/input.mp4",
		OutputDir: "/output",
		Profile:   ProfileVOD,
	}

	if job.Input != "/path/to/input.mp4" {
		t.Errorf("expected Input to be '/path/to/input.mp4', got '%s'", job.Input)
	}
	if job.OutputDir != "/output" {
		t.Errorf("expected OutputDir to be '/output', got '%s'", job.OutputDir)
	}
	if job.Profile != ProfileVOD {
		t.Errorf("expected Profile to be ProfileVOD, got '%s'", job.Profile)
	}
}
