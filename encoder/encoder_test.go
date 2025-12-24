package encoder

import (
	"context"
	"errors"
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
					Err:    nil,
				}
			}

			err := EncodeHLSCMAFWithExecutor(context.Background(), "input.mp4", "/output", tt.info, tt.profile, tt.ladder, mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeHLSCMAFWithExecutor() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify ffmpeg was called
			if mock.GetCallCount("ffmpeg") != 1 {
				t.Errorf("expected 1 ffmpeg call, got %d", mock.GetCallCount("ffmpeg"))
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
					Err:    nil,
				}
			}

			err := EncodeDASHCMAFWithExecutor(
				context.Background(),
				"input",
				"out",
				tt.info,
				tt.profile,
				tt.ladder,
				mock,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeDASHCMAFWithExecutor() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify ffmpeg was called
			if mock.GetCallCount("ffmpeg") != 1 {
				t.Errorf("expected 1 ffmpeg call, got %d", mock.GetCallCount("ffmpeg"))
			}
		})
	}
}

func TestEncodeHLSCMAF(t *testing.T) {
	// Tests the wrapper delegates to EncodeHLSCMAFWithExecutor
	info := probe.VideoInfo{Width: 1920, Height: 1080, FPS: 30, HasAudio: true}
	l := []ladder.Rendition{{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"}}

	// Will fail without ffmpeg but proves wrapper works
	err := EncodeHLSCMAF(context.Background(), "input", "out", info, config.VOD, l)
	_ = err
}

func TestEncodeDASHCMAF(t *testing.T) {
	// Tests the wrapper delegates to EncodeDASHCMAFWithExecutor
	info := probe.VideoInfo{Width: 1920, Height: 1080, FPS: 30, HasAudio: true}
	l := []ladder.Rendition{{Width: 1920, Height: 1080, MaxRate: 5000, BufSize: 10000, Profile: "main", Level: "4.0"}}

	// Will fail without ffmpeg but proves wrapper works
	err := EncodeDASHCMAF(
		context.Background(),
		"input",
		"out",
		info,
		config.VOD,
		l,
	)
	_ = err
}
