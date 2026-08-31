package mosaic

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
			tt.job.OutputDir = filepath.Join(t.TempDir(), "hls")
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

			_, err := EncodeHlsWithExecutor(context.Background(), tt.job, mock)

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
			tt.job.OutputDir = filepath.Join(t.TempDir(), "dash")
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

			_, err := EncodeDashWithExecutor(context.Background(), tt.job, mock)

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

func TestPrepareInputForEncoding(t *testing.T) {
	t.Run("normalization disabled", func(t *testing.T) {
		o := defaultOptions()
		o.normalizeOrientation = false
		mock := &orientationMockExecutor{}

		got, cleanup, err := prepareInputForEncoding(context.Background(), "input.mp4", mock, o)
		if err != nil {
			t.Fatalf("prepareInputForEncoding() err=%v", err)
		}
		defer cleanup()

		if got != "input.mp4" {
			t.Fatalf("got input %q want %q", got, "input.mp4")
		}
		if mock.ffmpegCalls != 0 {
			t.Fatalf("expected no ffmpeg calls, got %d", mock.ffmpegCalls)
		}
	})

	t.Run("normalization enabled", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := filepath.Join(dir, "in.mp4")
		if err := os.WriteFile(inputPath, []byte("src"), 0o644); err != nil {
			t.Fatalf("write input: %v", err)
		}

		o := defaultOptions()
		o.normalizeOrientation = true
		mock := &orientationMockExecutor{
			ffprobeOutputs: [][]byte{
				[]byte(`{"streams":[{"width":1920,"height":1080,"codec_name":"h264","side_data_list":[{"rotation":90}]}]}`),
				[]byte(`{"streams":[{"width":1080,"height":1920,"codec_name":"h264"}]}`),
			},
			createFFmpegOutput: true,
		}

		got, cleanup, err := prepareInputForEncoding(context.Background(), inputPath, mock, o)
		if err != nil {
			t.Fatalf("prepareInputForEncoding() err=%v", err)
		}
		if got == inputPath {
			t.Fatalf("expected temp normalized path, got original")
		}
		if _, err := os.Stat(got); err != nil {
			t.Fatalf("expected temp output to exist: %v", err)
		}
		cleanup()
		if _, err := os.Stat(got); !os.IsNotExist(err) {
			t.Fatalf("expected temp output cleanup, stat err=%v", err)
		}
	})
}

// sequentialMock handles sequential ffprobe calls
type sequentialMock struct {
	videoResponse  executor.MockResponse
	audioResponse  executor.MockResponse
	ffmpegResponse executor.MockResponse
	callCount      int
}

func (m *sequentialMock) Execute(ctx context.Context, name string, args ...string) ([]byte, *executor.Usage, error) {
	return m.ExecuteWithProgress(ctx, nil, name, args...)
}

func (m *sequentialMock) ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, *executor.Usage, error) {
	if progress != nil {
		close(progress)
	}
	m.callCount++
	if m.callCount == 1 {
		return m.videoResponse.Output, m.videoResponse.Usage, m.videoResponse.Err
	}
	if m.callCount == 2 {
		return m.audioResponse.Output, m.audioResponse.Usage, m.audioResponse.Err
	}
	if m.callCount == 3 {
		return m.ffmpegResponse.Output, m.ffmpegResponse.Usage, m.ffmpegResponse.Err
	}
	return nil, nil, fmt.Errorf("unexpected call to Execute: %s %v", name, args)
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

func (m *fullMock) Execute(ctx context.Context, name string, args ...string) ([]byte, *executor.Usage, error) {
	return m.ExecuteWithProgress(ctx, nil, name, args...)
}

func (m *fullMock) ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, *executor.Usage, error) {
	if name == "ffprobe" {
		if progress != nil {
			close(progress)
		}
		return m.probeVideoResponse.Output, m.probeVideoResponse.Usage, m.probeVideoResponse.Err
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
		return m.ffmpegResponse.Output, m.ffmpegResponse.Usage, m.ffmpegResponse.Err
	}

	if progress != nil {
		close(progress)
	}

	// ffprobe calls
	if m.callCount <= 2 {
		if m.callCount == 1 {
			// Video probe
			return m.probeVideoResponse.Output, m.probeVideoResponse.Usage, m.probeVideoResponse.Err
		}
		// Audio probe
		return m.probeAudioResponse.Output, m.probeAudioResponse.Usage, m.probeAudioResponse.Err
	}

	return nil, nil, errors.New("unexpected call")
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
		OutputDir: filepath.Join(t.TempDir(), "hls"),
		Profile:   ProfileVOD,
		ProgressHandler: func(info ProgressInfo) {
			progressUpdates = append(progressUpdates, info)
		},
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock)
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
	// Setup mock responses for probe and ffmpeg
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
		OutputDir: filepath.Join(t.TempDir(), "hls"),
		Profile:   ProfileVOD,
	}

	_, err := EncodeHls(context.Background(), job)

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
		OutputDir: filepath.Join(t.TempDir(), "dash"),
		Profile:   ProfileVOD,
	}

	_, err := EncodeDash(context.Background(), job)

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
		OutputDir: filepath.Join(t.TempDir(), "dash"),
		Profile:   ProfileVOD,
		ProgressHandler: func(info ProgressInfo) {
			progressCalled = true
		},
	}

	_, err := EncodeDashWithExecutor(context.Background(), job, mock)
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
	job := Job{Input: "test.mp4", OutputDir: filepath.Join(t.TempDir(), "hls"), Profile: ProfileVOD}
	_, err := EncodeHlsWithExecutor(context.Background(), job, mock)
	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestEncodeDashError(t *testing.T) {
	mock := &fullMock{
		probeVideoResponse: executor.MockResponse{Err: errors.New("probe failed")},
	}
	job := Job{Input: "test.mp4", OutputDir: filepath.Join(t.TempDir(), "dash"), Profile: ProfileVOD}
	_, err := EncodeDashWithExecutor(context.Background(), job, mock)
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
		OutputDir:       filepath.Join(t.TempDir(), "streaming"),
		Profile:         ProfileVOD,
		ProgressHandler: nil, // Explicitly nil
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock)
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor failed: %v", err)
	}

	_, err = EncodeDashWithExecutor(context.Background(), job, mock)
	if err != nil {
		t.Fatalf("EncodeDashWithExecutor failed: %v", err)
	}
}

func TestJobValidation(t *testing.T) {
	tests := []struct {
		name    string
		job     Job
		wantErr bool
	}{
		{
			name:    "empty input",
			job:     Job{Input: "", OutputDir: "/tmp/out"},
			wantErr: true,
		},
		{
			name:    "whitespace input",
			job:     Job{Input: "   ", OutputDir: "/tmp/out"},
			wantErr: true,
		},
		{
			name:    "empty output dir",
			job:     Job{Input: "test.mp4", OutputDir: ""},
			wantErr: true,
		},
		{
			name:    "whitespace output dir",
			job:     Job{Input: "test.mp4", OutputDir: "   "},
			wantErr: true,
		},
		{
			name:    "valid job",
			job:     Job{Input: "test.mp4", OutputDir: "/tmp/out"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test validate via EncodeHlsWithExecutor & EncodeDashWithExecutor
	invalidJob := Job{Input: "", OutputDir: ""}
	if _, err := EncodeHlsWithExecutor(context.Background(), invalidJob, &fullMock{}); err == nil {
		t.Error("expected EncodeHlsWithExecutor to fail validation, got nil")
	}
	if _, err := EncodeDashWithExecutor(context.Background(), invalidJob, &fullMock{}); err == nil {
		t.Error("expected EncodeDashWithExecutor to fail validation, got nil")
	}
}

func TestFunctionalOptions(t *testing.T) {
	o := defaultOptions()
	WithNormalizeOrientation()(o)
	if !o.normalizeOrientation {
		t.Error("expected normalizeOrientation to be true")
	}

	WithNormalizeOrientation(false)(o)
	if o.normalizeOrientation {
		t.Error("expected normalizeOrientation to be false")
	}

	WithNVENC()(o)
	if o.gpu != config.GPU_NVENC {
		t.Errorf("expected GPU_NVENC, got %v", o.gpu)
	}

	WithVideoToolbox()(o)
	if o.gpu != config.GPU_VIDEOTOOLBOX {
		t.Errorf("expected GPU_VIDEOTOOLBOX, got %v", o.gpu)
	}

	WithBFrames(2)(o)
	if o.bframes != 2 {
		t.Errorf("expected bframes 2, got %d", o.bframes)
	}

	WithScaleBitrateWithFPS()(o)
	if !o.scaleBitrateWithFPS {
		t.Error("expected scaleBitrateWithFPS to be true")
	}

	WithScaleBitrateWithFPS(false)(o)
	if o.scaleBitrateWithFPS {
		t.Error("expected scaleBitrateWithFPS to be false")
	}
}

func TestProgressPercentageCalculation(t *testing.T) {
	mock := &fullMock{
		probeVideoResponse: executor.MockResponse{
			Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1","duration":"10.000000"}]}`),
			Err:    nil,
		},
		probeAudioResponse: executor.MockResponse{Output: []byte("0"), Err: nil},
		ffmpegResponse:     executor.MockResponse{Output: []byte(""), Err: nil},
		progressData: []string{
			"frame=150\nout_time=00:00:05.000000\nprogress=continue\n",
		},
	}

	var capturedPercentage float64
	job := Job{
		Input:     "test.mp4",
		OutputDir: filepath.Join(t.TempDir(), "hls"),
		Profile:   ProfileVOD,
		ProgressHandler: func(info ProgressInfo) {
			capturedPercentage = info.Percentage
		},
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock)
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor failed: %v", err)
	}

	// 5.0 seconds out of 10.0 seconds total duration = 50.0%
	if capturedPercentage != 50.0 {
		t.Errorf("expected progress percentage 50.0, got %f", capturedPercentage)
	}
}

func TestNormalizedInputExt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to/video.mp4", ".mp4"},
		{"/path/to/video.MOV", ".mov"},
		{"/path/to/video.mkv", ".mkv"},
		{"/path/to/video.webm", ".webm"},
		{"/path/to/video.avi", ".avi"},
		{"/path/to/video.ts", ".ts"},
		{"https://example.com/video.mp4?token=123&exp=456", ".mp4"},
		{"https://example.com/video.m3u8?token=123", ".mp4"},
		{"https://api.tupic.com/assets/v1/assets/proxy?path=https%3A%2F%2Fstorage.tupic.com%2F2026%2F08%2F24%2Fprivate%2Fhls%2Fuploaded%2Findex.m3u8%3FExpires%3D1787699912%26Signature%3DE2RsxdM8ZI-C8Arl4ueUKSkCrPYqBfU2YXEywL~5n~~WQHL0QCZcqPmAts3OmD9p1YMRVeg2Ac67X8N-jN4vtMDLd-JFO-GiA6s2VgwS~VddQYVqhqubarV9Bp4ucHl5Jm5XBWyYT1krtmVeFYcwCrk7oIAhsSF~Fc-D11aeJTXf~SU-K4HtnoJZdSX7QoqqH2uKlWNE6dSCOTGU0TjGzTxhV2ksvHhCnqfI6xCOnd2jftI9aKJiLPW4Y6611~MhpS8PmUssLFB4VYV1qnFnP91eJV7tlQgnEORgRi02NTJvmtvFTn2UknHWFxQS~tcKBitJWzYu5va7PNiVFaqPUw__%26Key-Pair-Id%3DK1R1M45CWKN77T", ".mp4"},
		{"/path/to/file_without_ext", ".mp4"},
		{"", ".mp4"},
	}

	for _, tt := range tests {
		got := normalizedInputExt(tt.input)
		if got != tt.expected {
			t.Errorf("normalizedInputExt(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestWithThumbnails(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1","duration":"30.0"}],"format":{"duration":"30.0"}}`),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte("0"),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffmpeg", executor.MockResponse{
		Output: []byte(""),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffmpeg", executor.MockResponse{
		Output: []byte(""),
		Err:    nil,
	})

	tmpDir := t.TempDir()
	job := Job{
		Input:     "test.mp4",
		OutputDir: tmpDir,
		Profile:   ProfileVOD,
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock, WithThumbnails())
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor with thumbnails failed: %v", err)
	}

	vttFile := filepath.Join(tmpDir, "thumbnails.vtt")
	if _, err := os.Stat(vttFile); os.IsNotExist(err) {
		t.Errorf("expected thumbnails.vtt to be created, but it was not found")
	}
}

func TestWithIFrames(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1","duration":"30.0"}],"format":{"duration":"30.0"}}`),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte("0"),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffmpeg", executor.MockResponse{
		Output: []byte(""),
		Err:    nil,
	})

	tmpDir := t.TempDir()
	job := Job{
		Input:     "test.mp4",
		OutputDir: tmpDir,
		Profile:   ProfileVOD,
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock, WithIFrames())
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor with iframes failed: %v", err)
	}

	foundIframeFlag := false
	for _, call := range mock.CallLog {
		if call.Name == "ffmpeg" {
			for _, arg := range call.Args {
				if strings.Contains(arg, "iframes_only") {
					foundIframeFlag = true
					break
				}
			}
		}
	}

	if !foundIframeFlag {
		t.Errorf("expected iframes_only flag in ffmpeg args")
	}
}

func TestWithNormalizeAudio(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1","duration":"30.0"}],"format":{"duration":"30.0"}}`),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte("1"),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffmpeg", executor.MockResponse{
		Output: []byte(""),
		Err:    nil,
	})

	tmpDir := t.TempDir()
	job := Job{
		Input:     "test.mp4",
		OutputDir: tmpDir,
		Profile:   ProfileVOD,
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock, WithNormalizeAudio())
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor with audio normalization failed: %v", err)
	}

	foundLoudnorm := false
	for _, call := range mock.CallLog {
		if call.Name == "ffmpeg" {
			for _, arg := range call.Args {
				if strings.Contains(arg, "loudnorm") {
					foundLoudnorm = true
					break
				}
			}
		}
	}

	if !foundLoudnorm {
		t.Errorf("expected loudnorm filter in ffmpeg args")
	}
}

func TestWithSubtitles(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1","duration":"30.0"}],"format":{"duration":"30.0"}}`),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte("0"),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffmpeg", executor.MockResponse{
		Output: []byte(""),
		Err:    nil,
	})

	tmpDir := t.TempDir()
	subFile := filepath.Join(tmpDir, "sample.srt")
	_ = os.WriteFile(subFile, []byte("1\n00:00:01,000 --> 00:00:03,000\nTest\n"), 0o644)

	job := Job{
		Input:     "test.mp4",
		OutputDir: tmpDir,
		Profile:   ProfileVOD,
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock, WithSubtitles(SubtitleTrack{
		Path:     subFile,
		Language: "fa",
		Label:    "فارسی",
		Default:  true,
	}))
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor with subtitles failed: %v", err)
	}

	subM3U8 := filepath.Join(tmpDir, "sub_fa.m3u8")
	if _, err := os.Stat(subM3U8); os.IsNotExist(err) {
		t.Errorf("expected sub_fa.m3u8 to be created")
	}
}

func TestWithWatermark(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1","duration":"30.0"}],"format":{"duration":"30.0"}}`),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte("0"),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffmpeg", executor.MockResponse{
		Output: []byte(""),
		Err:    nil,
	})

	tmpDir := t.TempDir()
	job := Job{
		Input:     "test.mp4",
		OutputDir: tmpDir,
		Profile:   ProfileVOD,
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock, WithWatermark(WatermarkConfig{
		Path:     "logo.png",
		Position: PositionBottomRight,
	}))
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor with watermark failed: %v", err)
	}

	foundOverlay := false
	for _, call := range mock.CallLog {
		if call.Name == "ffmpeg" {
			for _, arg := range call.Args {
				if strings.Contains(arg, "overlay=") {
					foundOverlay = true
					break
				}
			}
		}
	}

	if !foundOverlay {
		t.Errorf("expected watermark overlay in ffmpeg args")
	}
}

func TestWithAES128Encryption(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1","duration":"30.0"}],"format":{"duration":"30.0"}}`),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte("0"),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffmpeg", executor.MockResponse{
		Output: []byte(""),
		Err:    nil,
	})

	tmpDir := t.TempDir()
	job := Job{
		Input:     "test.mp4",
		OutputDir: tmpDir,
		Profile:   ProfileVOD,
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock, WithAES128Encryption())
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor with encryption failed: %v", err)
	}

	foundKeyInfo := false
	for _, call := range mock.CallLog {
		if call.Name == "ffmpeg" {
			for _, arg := range call.Args {
				if strings.Contains(arg, "enc.keyinfo") {
					foundKeyInfo = true
					break
				}
			}
		}
	}

	if !foundKeyInfo {
		t.Errorf("expected -hls_key_info_file with enc.keyinfo in ffmpeg args")
	}

	keyFile := filepath.Join(tmpDir, "enc.key")
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Errorf("expected enc.key to be created")
	}
}

func TestWithCodecAndCRF(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1","duration":"30.0"}],"format":{"duration":"30.0"}}`),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte("0"),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffmpeg", executor.MockResponse{
		Output: []byte(""),
		Err:    nil,
	})

	tmpDir := t.TempDir()
	job := Job{
		Input:     "test.mp4",
		OutputDir: tmpDir,
		Profile:   ProfileVOD,
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock, WithAV1(), WithCRF(28))
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor with AV1 and CRF failed: %v", err)
	}

	foundAV1 := false
	foundCRF := false
	for _, call := range mock.CallLog {
		if call.Name == "ffmpeg" {
			for i, arg := range call.Args {
				if arg == "libsvtav1" {
					foundAV1 = true
				}
				if arg == "-crf:v:0" && i+1 < len(call.Args) && call.Args[i+1] == "28" {
					foundCRF = true
				}
			}
		}
	}

	if !foundAV1 {
		t.Errorf("expected libsvtav1 encoder in ffmpeg args")
	}
	if !foundCRF {
		t.Errorf("expected -crf:v:0 28 in ffmpeg args")
	}
}

func TestWithS3Upload(t *testing.T) {
	uploadedCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadedCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mock := executor.NewMockExecutor()
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1","duration":"30.0"}],"format":{"duration":"30.0"}}`),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffprobe", executor.MockResponse{
		Output: []byte("0"),
		Err:    nil,
	})
	mock.AddSequentialResponse("ffmpeg", executor.MockResponse{
		Output: []byte(""),
		Err:    nil,
	})

	tmpDir := t.TempDir()
	// create dummy segment to upload
	_ = os.WriteFile(filepath.Join(tmpDir, "master.m3u8"), []byte("#EXTM3U"), 0o644)

	job := Job{
		Input:     "test.mp4",
		OutputDir: tmpDir,
		Profile:   ProfileVOD,
	}

	_, err := EncodeHlsWithExecutor(context.Background(), job, mock, WithS3Upload(S3Config{
		Bucket:   "test-bucket",
		Endpoint: server.URL,
	}))
	if err != nil {
		t.Fatalf("EncodeHlsWithExecutor with S3 upload failed: %v", err)
	}

	if uploadedCount == 0 {
		t.Errorf("expected S3 upload requests to be made to mock server")
	}
}
