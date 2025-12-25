package mosaic

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/farshidrezaei/mosaic/config"
	"github.com/farshidrezaei/mosaic/internal/executor"
)

func TestInitializeWithExecutor(t *testing.T) {
	tests := []struct {
		responses map[string]executor.MockResponse
		job       Job
		name      string
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

			_, _, renditions, err := initializeWithExecutor(context.Background(), tt.job, mock, defaultOptions())

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
}

func (m *sequentialMock) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	return m.ExecuteWithProgress(ctx, nil, name, args...)
}

func (m *sequentialMock) ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, error) {
	if progress != nil {
		close(progress)
	}
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
	progressData       []string
	callCount          int
	ffmpegCallCount    int
}

func (m *fullMock) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	return m.ExecuteWithProgress(ctx, nil, name, args...)
}

func (m *fullMock) ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, error) {
	if name == "ffprobe" {
		if progress != nil {
			close(progress)
		}
		return m.probeVideoResponse.Output, m.probeVideoResponse.Err
	}

	m.callCount++

	if name == "ffmpeg" {
		m.ffmpegCallCount++
		if progress != nil {
			for _, p := range m.progressData {
				progress <- p
			}
			close(progress)
		}
		return m.ffmpegResponse.Output, m.ffmpegResponse.Err
	}

	if progress != nil {
		close(progress)
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

func TestProgressReporting(t *testing.T) {
	mock := &fullMock{
		probeVideoResponse: executor.MockResponse{
			Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
			Err:    nil,
		},
		probeAudioResponse: executor.MockResponse{Output: []byte("0"), Err: nil},
		ffmpegResponse:     executor.MockResponse{Output: []byte(""), Err: nil},
		progressData: []string{
			"frame=100\nfps=30.0\nstream_0_0_q=28.0\nbitrate=1000.0kbits/s\ntotal_size=1000000\nout_time_us=10000000\nout_time_ms=10000\nout_time=00:00:10.000000\ndup_frames=0\ndrop_frames=0\nspeed=1.5x\nprogress=continue\n",
			"frame=200\nfps=30.0\nstream_0_0_q=28.0\nbitrate=1200.0kbits/s\ntotal_size=2000000\nout_time_us=20000000\nout_time_ms=20000\nout_time=00:00:20.000000\ndup_frames=0\ndrop_frames=0\nspeed=1.6x\nprogress=end\n",
		},
	}

	var progressUpdates []ProgressInfo
	job := Job{
		Input:     "test.mp4",
		OutputDir: "/output",
		Profile:   ProfileVOD,
		ProgressHandler: func(info ProgressInfo) {
			progressUpdates = append(progressUpdates, info)
		},
	}

	err := EncodeHlsWithExecutor(context.Background(), job, mock)
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor failed: %v", err)
	}

	if len(progressUpdates) != 2 {
		t.Errorf("expected 2 progress updates, got %d", len(progressUpdates))
	}

	if progressUpdates[0].CurrentTime != "00:00:10.000000" {
		t.Errorf("expected time 00:00:10.000000, got %s", progressUpdates[0].CurrentTime)
	}
	if progressUpdates[0].Speed != "1.5x" {
		t.Errorf("expected speed 1.5x, got %s", progressUpdates[0].Speed)
	}
	if progressUpdates[1].Bitrate != "1200.0kbits/s" {
		t.Errorf("expected bitrate 1200.0kbits/s, got %s", progressUpdates[1].Bitrate)
	}
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

func TestOptions(t *testing.T) {
	o := defaultOptions()

	WithThreads(8)(o)
	if o.threads != 8 {
		t.Errorf("expected 8 threads, got %d", o.threads)
	}

	WithGPU()(o)
	if o.gpu != config.GPU_NVENC {
		t.Errorf("expected GPU_NVENC, got %s", o.gpu)
	}

	WithVAAPI()(o)
	if o.gpu != config.GPU_VAAPI {
		t.Errorf("expected GPU_VAAPI, got %s", o.gpu)
	}

	WithLogLevel("debug")(o)
	if o.logLevel != "debug" {
		t.Errorf("expected loglevel debug, got %s", o.logLevel)
	}

	logger := slog.Default()
	WithLogger(logger)(o)
	if o.logger != logger {
		t.Error("expected custom logger")
	}
}

func TestProgressReportingDash(t *testing.T) {
	mock := &fullMock{
		probeVideoResponse: executor.MockResponse{
			Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
			Err:    nil,
		},
		probeAudioResponse: executor.MockResponse{Output: []byte("0"), Err: nil},
		ffmpegResponse:     executor.MockResponse{Output: []byte(""), Err: nil},
		progressData: []string{
			"frame=100\nout_time=00:00:10.000000\nprogress=continue\n",
		},
	}

	var progressCalled bool
	job := Job{
		Input:     "test.mp4",
		OutputDir: "/output",
		Profile:   ProfileVOD,
		ProgressHandler: func(info ProgressInfo) {
			progressCalled = true
		},
	}

	err := EncodeDashWithExecutor(context.Background(), job, mock)
	if err != nil {
		t.Fatalf("EncodeDashWithExecutor failed: %v", err)
	}

	if !progressCalled {
		t.Error("expected progress handler to be called")
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

	_, _, _, err := initialize(context.Background(), job, defaultOptions())
	if err != nil {
		t.Errorf("initialize() error = %v", err)
	}
	seqMock2 := &sequentialMock{
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
	executor.DefaultExecutor = seqMock2
	job.Profile = ProfileLive
	_, profile, _, err := initialize(context.Background(), job, defaultOptions())
	if err != nil {
		t.Errorf("initialize() error = %v", err)
	}
	if !profile.LowLatency {
		t.Error("expected low latency to be true")
	}
}
func TestEncodeHlsError(t *testing.T) {
	mock := &fullMock{
		probeVideoResponse: executor.MockResponse{Err: errors.New("probe failed")},
	}
	job := Job{Input: "test.mp4", OutputDir: "/out", Profile: ProfileVOD}
	err := EncodeHlsWithExecutor(context.Background(), job, mock)
	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestEncodeDashError(t *testing.T) {
	mock := &fullMock{
		probeVideoResponse: executor.MockResponse{Err: errors.New("probe failed")},
	}
	job := Job{Input: "test.mp4", OutputDir: "/out", Profile: ProfileVOD}
	err := EncodeDashWithExecutor(context.Background(), job, mock)
	if err == nil {
		t.Error("expected error but got none")
	}
}
func TestNilProgressHandler(t *testing.T) {
	mock := &fullMock{
		probeVideoResponse: executor.MockResponse{
			Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
			Err:    nil,
		},
		probeAudioResponse: executor.MockResponse{Output: []byte("0"), Err: nil},
		ffmpegResponse:     executor.MockResponse{Output: []byte(""), Err: nil},
		progressData: []string{
			"frame=100\nout_time=00:00:10.000000\nprogress=continue\n",
		},
	}

	job := Job{
		Input:           "test.mp4",
		OutputDir:       "/output",
		Profile:         ProfileVOD,
		ProgressHandler: nil, // Explicitly nil
	}

	err := EncodeHlsWithExecutor(context.Background(), job, mock)
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor failed: %v", err)
	}

	err = EncodeDashWithExecutor(context.Background(), job, mock)
	if err != nil {
		t.Fatalf("EncodeDashWithExecutor failed: %v", err)
	}
}
