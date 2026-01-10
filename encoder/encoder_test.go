package encoder

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/farshidrezaei/mosaic/config"
	"github.com/farshidrezaei/mosaic/internal/executor"
	"github.com/farshidrezaei/mosaic/ladder"
	"github.com/farshidrezaei/mosaic/probe"
)

func TestEncodeHLSCMAFWithExecutor(t *testing.T) {
	tests := []struct {
		name    string
		ladder  []ladder.Rendition
		info    probe.VideoInfo
		profile config.Profile
		wantErr bool
	}{
		{
			name: "successful encode with audio",
			info: probe.VideoInfo{
				Width:    1920,
				Height:   1080,
				FPS:      30.0,
				HasAudio: true,
			},
			profile: config.VOD,
			ladder: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
			},
			wantErr: false,
		},
		{
			name: "successful encode without audio",
			info: probe.VideoInfo{
				Width:    1280,
				Height:   720,
				FPS:      25.0,
				HasAudio: false,
			},
			profile: config.LIVE,
			ladder: []ladder.Rendition{
				{Width: 1280, Height: 720, MaxRate: 3000, BufSize: 6000, Profile: "main", Level: "3.1"},
			},
			wantErr: false,
		},
		{
			name: "ffmpeg failure",
			info: probe.VideoInfo{
				Width:    640,
				Height:   360,
				FPS:      30.0,
				HasAudio: false,
			},
			profile: config.VOD,
			ladder: []ladder.Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			wantErr: true,
		},
		{
			name:    "GPU and threads",
			info:    probe.VideoInfo{Width: 1920, Height: 1080, FPS: 30, HasAudio: true},
			profile: config.VOD,
			ladder: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := executor.NewMockExecutor()
			if tt.wantErr {
				mock.Responses["ffmpeg"] = executor.MockResponse{
					Output: nil,
					Err:    errors.New("ffmpeg failed"),
				}
			} else {
				mock.Responses["ffmpeg"] = executor.MockResponse{
					Output: []byte(""),
					Usage:  &executor.Usage{UserTime: 1.0},
					Err:    nil,
				}
			}

			opts := EncoderOptions{LogLevel: "warning"}
			if tt.name == "GPU and threads" {
				opts.GPU = config.GPU_NVENC
				opts.Threads = 4
			}

			var progressCalled bool
			progressHandler := func(m map[string]string) {
				progressCalled = true
			}

			usage, err := EncodeHLSCMAFWithExecutor(context.Background(), "input.mp4", "/output", tt.info, tt.profile, tt.ladder, mock, progressHandler, opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeHLSCMAFWithExecutor() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if usage == nil {
					t.Error("expected usage stats, got nil")
				}
			}

			_ = progressCalled

			// Verify ffmpeg was called
			if mock.GetCallCount("ffmpeg") != 1 {
				t.Errorf("expected 1 ffmpeg call, got %d", mock.GetCallCount("ffmpeg"))
			}

			if tt.name == "GPU and threads" {
				args := mock.CallLog[0].Args
				foundGPU := false
				foundThreads := false
				for i, arg := range args {
					if arg == "-c:v:0" && args[i+1] == "h264_nvenc" {
						foundGPU = true
					}
					if arg == "-threads" && args[i+1] == "4" {
						foundThreads = true
					}
				}
				if !foundGPU {
					t.Error("expected h264_nvenc codec")
				}
				if !foundThreads {
					t.Error("expected -threads 4")
				}
			}
		})
	}
}

func TestEncodeDASHCMAFWithExecutor(t *testing.T) {
	tests := []struct {
		name    string
		ladder  []ladder.Rendition
		info    probe.VideoInfo
		profile config.Profile
		wantErr bool
	}{
		{
			name: "successful encode with audio",
			info: probe.VideoInfo{
				Width:    1920,
				Height:   1080,
				FPS:      30.0,
				HasAudio: true,
			},
			profile: config.VOD,
			ladder: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
			},
			wantErr: false,
		},
		{
			name: "ffmpeg failure",
			info: probe.VideoInfo{
				Width:    640,
				Height:   360,
				FPS:      30.0,
				HasAudio: false,
			},
			profile: config.VOD,
			ladder: []ladder.Rendition{
				{Width: 640, Height: 360, MaxRate: 1000, BufSize: 2000, Profile: "baseline", Level: "3.0"},
			},
			wantErr: true,
		},
		{
			name:    "GPU and threads",
			info:    probe.VideoInfo{Width: 1920, Height: 1080, FPS: 30, HasAudio: true},
			profile: config.VOD,
			ladder: []ladder.Rendition{
				{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := executor.NewMockExecutor()
			if tt.wantErr {
				mock.Responses["ffmpeg"] = executor.MockResponse{
					Output: nil,
					Err:    errors.New("ffmpeg failed"),
				}
			} else {
				mock.Responses["ffmpeg"] = executor.MockResponse{
					Output: []byte(""),
					Usage:  &executor.Usage{UserTime: 1.0},
				}
			}

			opts := EncoderOptions{LogLevel: "warning"}
			if tt.name == "GPU and threads" {
				opts.GPU = config.GPU_NVENC
				opts.Threads = 4
			}

			var progressCalled bool
			progressHandler := func(m map[string]string) {
				progressCalled = true
			}

			usage, err := EncodeDASHCMAFWithExecutor(
				context.Background(),
				"input",
				"out",
				tt.info,
				tt.profile,
				tt.ladder,
				mock,
				progressHandler,
				opts,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeDASHCMAFWithExecutor() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if usage == nil {
					t.Error("expected usage stats, got nil")
				}
			}

			_ = progressCalled

			// Verify ffmpeg was called
			if mock.GetCallCount("ffmpeg") != 1 {
				t.Errorf("expected 1 ffmpeg call, got %d", mock.GetCallCount("ffmpeg"))
			}

			if tt.name == "GPU and threads" {
				args := mock.CallLog[0].Args
				foundGPU := false
				foundThreads := false
				for i, arg := range args {
					if arg == "-c:v:0" && args[i+1] == "h264_nvenc" {
						foundGPU = true
					}
					if arg == "-threads" && args[i+1] == "4" {
						foundThreads = true
					}
				}
				if !foundGPU {
					t.Error("expected h264_nvenc codec")
				}
				if !foundThreads {
					t.Error("expected -threads 4")
				}
			}
		})
	}
}

func TestEncodeHLSCMAF(t *testing.T) {
	// This test uses the real executor, so we skip it in short mode or if ffmpeg is not available.
	// For unit testing purposes, we can't easily mock the default executor without changing global state.
	// So we'll just skip this or rely on integration tests.
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	orig := executor.DefaultExecutor
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{Output: []byte(""), Err: nil}
	executor.DefaultExecutor = mock
	defer func() { executor.DefaultExecutor = orig }()

	info := probe.VideoInfo{Width: 1920, Height: 1080, FPS: 30, HasAudio: true}
	l := []ladder.Rendition{{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"}}

	_, err := EncodeHLSCMAF(context.Background(), "input", "out", info, config.VOD, l)
	if err != nil {
		t.Errorf("EncodeHLSCMAF failed: %v", err)
	}
}

func TestEncodeDASHCMAF(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	orig := executor.DefaultExecutor
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{Output: []byte(""), Err: nil}
	executor.DefaultExecutor = mock
	defer func() { executor.DefaultExecutor = orig }()

	info := probe.VideoInfo{Width: 1920, Height: 1080, FPS: 30, HasAudio: true}
	l := []ladder.Rendition{{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"}}

	_, err := EncodeDASHCMAF(context.Background(), "input", "out", info, config.VOD, l)
	if err != nil {
		t.Errorf("EncodeDASHCMAF failed: %v", err)
	}
}

func TestEncodeHLSCMAFWithExecutor_Progress(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{
		Output:       []byte(""),
		ProgressData: []string{"frame=100\nout_time=00:00:01.000000"},
		Usage:        &executor.Usage{UserTime: 1.0},
	}

	info := probe.VideoInfo{FPS: 30}
	profile := config.VOD
	l := []ladder.Rendition{{Width: 1920, Height: 1080}}

	var progressCalled bool
	usage, err := EncodeHLSCMAFWithExecutor(context.Background(), "in.mp4", "out", info, profile, l, mock, func(m map[string]string) {
		progressCalled = true
	}, EncoderOptions{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !progressCalled {
		t.Error("expected progress handler to be called")
	}
	if usage == nil {
		t.Fatal("expected usage stats, got nil")
	}
}

func TestEncodeDASHCMAFWithExecutor_Progress(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{
		Output:       []byte(""),
		ProgressData: []string{"frame=100\nout_time=00:00:01.000000"},
		Usage:        &executor.Usage{UserTime: 1.0},
	}

	info := probe.VideoInfo{FPS: 30}
	profile := config.VOD
	l := []ladder.Rendition{{Width: 1920, Height: 1080}}

	var progressCalled bool
	usage, err := EncodeDASHCMAFWithExecutor(context.Background(), "in.mp4", "out", info, profile, l, mock, func(m map[string]string) {
		progressCalled = true
	}, EncoderOptions{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !progressCalled {
		t.Error("expected progress handler to be called")
	}
	if usage == nil {
		t.Fatal("expected usage stats, got nil")
	}
}

func TestEncodeHLSCMAFWithExecutor_Error(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{
		Err: fmt.Errorf("ffmpeg failed"),
	}

	info := probe.VideoInfo{FPS: 30}
	profile := config.VOD
	l := []ladder.Rendition{{Width: 1920, Height: 1080}}

	_, err := EncodeHLSCMAFWithExecutor(context.Background(), "in.mp4", "out", info, profile, l, mock, nil, EncoderOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEncodeDASHCMAFWithExecutor_Error(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{
		Err: fmt.Errorf("ffmpeg failed"),
	}

	info := probe.VideoInfo{FPS: 30}
	profile := config.VOD
	l := []ladder.Rendition{{Width: 1920, Height: 1080}}

	_, err := EncodeDASHCMAFWithExecutor(context.Background(), "in.mp4", "out", info, profile, l, mock, nil, EncoderOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEncodeHLSCMAFWithExecutor_Progress_Error(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{
		Err: fmt.Errorf("ffmpeg failed"),
	}

	info := probe.VideoInfo{FPS: 30}
	profile := config.VOD
	l := []ladder.Rendition{{Width: 1920, Height: 1080}}

	_, err := EncodeHLSCMAFWithExecutor(context.Background(), "in.mp4", "out", info, profile, l, mock, func(m map[string]string) {}, EncoderOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEncodeDASHCMAFWithExecutor_Progress_Error(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{
		Err: fmt.Errorf("ffmpeg failed"),
	}

	info := probe.VideoInfo{FPS: 30}
	profile := config.VOD
	l := []ladder.Rendition{{Width: 1920, Height: 1080}}

	_, err := EncodeDASHCMAFWithExecutor(context.Background(), "in.mp4", "out", info, profile, l, mock, func(m map[string]string) {}, EncoderOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEncodeLowLatency(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{Output: []byte(""), Err: nil}

	info := probe.VideoInfo{Width: 1920, Height: 1080, FPS: 30, HasAudio: true}
	profile := config.LIVE // Low latency
	ladder := []ladder.Rendition{{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"}}

	t.Run("HLS Low Latency", func(t *testing.T) {
		_, err := EncodeHLSCMAFWithExecutor(context.Background(), "in", "out", info, profile, ladder, mock, nil, EncoderOptions{})
		if err != nil {
			t.Errorf("EncodeHLSCMAFWithExecutor failed: %v", err)
		}
	})

	t.Run("DASH Low Latency", func(t *testing.T) {
		_, err := EncodeDASHCMAFWithExecutor(context.Background(), "in", "out", info, profile, ladder, mock, nil, EncoderOptions{})
		if err != nil {
			t.Errorf("EncodeDASHCMAFWithExecutor failed: %v", err)
		}
	})
}

func TestEncodeEncoderError(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Responses["ffmpeg"] = executor.MockResponse{Err: errors.New("ffmpeg failed")}

	info := probe.VideoInfo{Width: 1920, Height: 1080, FPS: 30, HasAudio: true}
	l := []ladder.Rendition{{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"}}

	t.Run("HLS Error with Progress", func(t *testing.T) {
		_, err := EncodeHLSCMAFWithExecutor(context.Background(), "in", "out", info, config.VOD, l, mock, func(m map[string]string) {}, EncoderOptions{})
		if err == nil {
			t.Error("expected error but got none")
		}
	})

	t.Run("DASH Error with Progress", func(t *testing.T) {
		_, err := EncodeDASHCMAFWithExecutor(context.Background(), "in", "out", info, config.VOD, l, mock, func(m map[string]string) {}, EncoderOptions{})
		if err == nil {
			t.Error("expected error but got none")
		}
	})

	t.Run("HLS Error without Progress", func(t *testing.T) {
		_, err := EncodeHLSCMAFWithExecutor(context.Background(), "in", "out", info, config.VOD, l, mock, nil, EncoderOptions{})
		if err == nil {
			t.Error("expected error but got none")
		}
	})

	t.Run("DASH Error without Progress", func(t *testing.T) {
		_, err := EncodeDASHCMAFWithExecutor(context.Background(), "in", "out", info, config.VOD, l, mock, nil, EncoderOptions{})
		if err == nil {
			t.Error("expected error but got none")
		}
	})
}
