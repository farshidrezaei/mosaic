package mediapack

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/farshidrezaei/mosaic/internal/executor"
)

func TestInitializeWithExecutor(t *testing.T) {
	tests := []struct {
		name      string
		job       Job
		responses map[string]executor.MockResponse
		wantErr   bool
	}{
		{
			name: "VOD profile success",
			job: Job{
				Input:     "test.mp4",
				OutputDir: "/output",
				Profile:   ProfileVOD,
			},
			responses: map[string]executor.MockResponse{
				"ffprobe": {
					Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
					Err:    nil,
				},
			},
			wantErr: false,
		},
		{
			name: "LIVE profile  success",
			job: Job{
				Input:     "test.mp4",
				OutputDir: "/output",
				Profile:   ProfileLive,
			},
			responses: map[string]executor.MockResponse{
				"ffprobe": {
					Output: []byte(`{"streams":[{"width":1280,"height":720,"avg_frame_rate":"25/1"}]}`),
					Err:    nil,
				},
			},
			wantErr: false,
		},
		{
			name: "probe error",
			job: Job{
				Input:     "bad.mp4",
				OutputDir: "/output",
				Profile:   ProfileVOD,
			},
			responses: map[string]executor.MockResponse{
				"ffprobe": {
					Output: nil,
					Err:    errors.New("file not found"),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &sequentialMock{
				videoResponse:  tt.responses["ffprobe"],
				audioResponse:  executor.MockResponse{Output: []byte("0"), Err: nil},
				ffmpegResponse: executor.MockResponse{Output: []byte(""), Err: nil},
			}

			_, _, renditions, err := initializeWithExecutor(context.Background(), tt.job, mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("initializeWithExecutor() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify we got renditions
				if len(renditions) == 0 {
					t.Error("expected renditions but got none")
				}
			}
		})
	}
}

func TestEncodeHlsWithExecutor(t *testing.T) {
	tests := []struct {
		name    string
		job     Job
		wantErr bool
	}{
		{
			name: "successful HLS encoding",
			job: Job{
				Input:     "test.mp4",
				OutputDir: "/output/hls",
				Profile:   ProfileVOD,
			},
			wantErr: false,
		},
		{
			name: "probe fails",
			job: Job{
				Input:     "bad.mp4",
				OutputDir: "/output/hls",
				Profile:   ProfileVOD,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fullMock{
				probeVideoResponse: executor.MockResponse{
					Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
					Err:    nil,
				},
				probeAudioResponse: executor.MockResponse{Output: []byte("0"), Err: nil},
				ffmpegResponse:     executor.MockResponse{Output: []byte(""), Err: nil},
			}

			if tt.wantErr {
				mock.probeVideoResponse.Err = errors.New("file not found")
			}

			err := EncodeHlsWithExecutor(context.Background(), tt.job, mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeHlsWithExecutor() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify ffmpeg was called for successful cases
			if !tt.wantErr && mock.ffmpegCallCount == 0 {
				t.Error("expected ffmpeg to be called but was not")
			}
		})
	}
}

func TestEncodeDashWithExecutor(t *testing.T) {
	tests := []struct {
		name    string
		job     Job
		wantErr bool
	}{
		{
			name: "successful DASH encoding",
			job: Job{
				Input:     "test.mp4",
				OutputDir: "/output/dash",
				Profile:   ProfileVOD,
			},
			wantErr: false,
		},
		{
			name: "ffmpeg fails",
			job: Job{
				Input:     "test.mp4",
				OutputDir: "/output/dash",
				Profile:   ProfileVOD,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &fullMock{
				probeVideoResponse: executor.MockResponse{
					Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
					Err:    nil,
				},
				probeAudioResponse: executor.MockResponse{Output: []byte("0"), Err: nil},
				ffmpegResponse:     executor.MockResponse{Output: []byte(""), Err: nil},
			}

			if tt.wantErr {
				mock.ffmpegResponse.Err = errors.New("encoding failed")
			}

			err := EncodeDashWithExecutor(context.Background(), tt.job, mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeDashWithExecutor() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify ffmpeg was called for successful cases
			if !tt.wantErr && mock.ffmpegCallCount == 0 {
				t.Error("expected ffmpeg to be called but was not")
			}
		})
	}
}

// sequentialMock handles sequential ffprobe calls
type sequentialMock struct {
	videoResponse  executor.MockResponse
	audioResponse  executor.MockResponse
	ffmpegResponse executor.MockResponse
	callCount      int
	responses      []executor.MockResponse // Added to make the provided snippet compile
}

func (m *sequentialMock) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	m.callCount++
	if m.callCount == 1 {
		return m.videoResponse.Output, m.videoResponse.Err
	}
	if m.callCount == 2 {
		return m.audioResponse.Output, m.audioResponse.Err
	}
	if m.callCount == 3 {
		return m.ffmpegResponse.Output, m.ffmpegResponse.Err
	}
	return nil, fmt.Errorf("unexpected call to Execute: %s %v", name, args)
}

// fullMock handles all commands (ffprobe x2, ffmpeg)
type fullMock struct {
	probeVideoResponse executor.MockResponse
	probeAudioResponse executor.MockResponse
	ffmpegResponse     executor.MockResponse
	callCount          int
	ffmpegCallCount    int
}

func (m *fullMock) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	if name == "ffprobe" {
		return m.probeVideoResponse.Output, m.probeVideoResponse.Err
	}

	m.callCount++

	if name == "ffmpeg" {
		m.ffmpegCallCount++
		return m.ffmpegResponse.Output, m.ffmpegResponse.Err
	}

	// ffprobe calls
	if m.callCount <= 2 {
		if m.callCount == 1 {
			// Video probe
			return m.probeVideoResponse.Output, m.probeVideoResponse.Err
		}
		// Audio probe
		return m.probeAudioResponse.Output, m.probeAudioResponse.Err
	}

	return nil, errors.New("unexpected call")
}

func TestEncodeHls(t *testing.T) {
	// Save original executor and restore after test
	origExec := executor.DefaultExecutor
	defer func() { executor.DefaultExecutor = origExec }()

	// Use a mock executor
	mock := executor.NewMockExecutor()
	// Setup mock responses for probe and ffmpeg
	mock.Responses["ffprobe"] = executor.MockResponse{
		Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
		Err:    nil,
	}
	// We need two ffprobe responses (video and audio)
	// But the mock implementation in executor package might be simple map lookup
	// Let's check how NewMockExecutor works. It uses a map.
	// If we need sequential responses for same command, the simple map mock might not suffice
	// unless we use the sequential mock defined in this file.
	// But DefaultExecutor is of type executor.CommandExecutor.
	// The sequentialMock in this file implements that interface.

	seqMock := &sequentialMock{
		videoResponse: executor.MockResponse{
			Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
			Err:    nil,
		},
		audioResponse: executor.MockResponse{
			Output: []byte("0"),
			Err:    nil,
		},
		ffmpegResponse: executor.MockResponse{
			Output: []byte(""),
			Err:    nil,
		},
	}
	executor.DefaultExecutor = seqMock

	// This test verifies the wrapper function exists and delegates correctly
	job := Job{
		Input:     "test.mp4",
		OutputDir: "/output",
		Profile:   ProfileVOD,
	}

	err := EncodeHls(context.Background(), job)

	if err != nil {
		t.Errorf("EncodeHls() error = %v", err)
	}
}

func TestEncodeDash(t *testing.T) {
	// Save original executor and restore after test
	origExec := executor.DefaultExecutor
	defer func() { executor.DefaultExecutor = origExec }()

	seqMock := &sequentialMock{
		videoResponse: executor.MockResponse{
			Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
			Err:    nil,
		},
		audioResponse: executor.MockResponse{
			Output: []byte("0"),
			Err:    nil,
		},
		ffmpegResponse: executor.MockResponse{
			Output: []byte(""),
			Err:    nil,
		},
	}
	executor.DefaultExecutor = seqMock

	// This test verifies the wrapper function exists and delegates correctly
	job := Job{
		Input:     "test.mp4",
		OutputDir: "/output",
		Profile:   ProfileVOD,
	}

	err := EncodeDash(context.Background(), job)

	if err != nil {
		t.Errorf("EncodeDash() error = %v", err)
	}
}

func TestInitialize(t *testing.T) {
	// Save original executor and restore after test
	origExec := executor.DefaultExecutor
	defer func() { executor.DefaultExecutor = origExec }()

	seqMock := &sequentialMock{
		videoResponse: executor.MockResponse{
			Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
			Err:    nil,
		},
		audioResponse: executor.MockResponse{
			Output: []byte("0"),
			Err:    nil,
		},
		ffmpegResponse: executor.MockResponse{
			Output: []byte(""),
			Err:    nil,
		},
	}
	executor.DefaultExecutor = seqMock

	// This test verifies the wrapper function exists
	job := Job{
		Input:     "test.mp4",
		OutputDir: "/output",
		Profile:   ProfileVOD,
	}

	_, _, _, err := initialize(context.Background(), job)
	if err != nil {
		t.Errorf("initialize() error = %v", err)
	}
}
