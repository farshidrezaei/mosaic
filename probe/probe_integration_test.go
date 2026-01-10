package probe

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/farshidrezaei/mosaic/internal/executor"
)

func TestInputWithExecutor(t *testing.T) {
	tests := []struct {
		responses map[string]executor.MockResponse
		name      string
		wantInfo  VideoInfo
		wantErr   bool
	}{
		{
			name: "1080p video with audio",
			responses: map[string]executor.MockResponse{
				"ffprobe": {
					Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30/1"}]}`),
					Err:    nil,
				},
			},
			wantInfo: VideoInfo{
				Width:    1920,
				Height:   1080,
				FPS:      30.0,
				HasAudio: true, // Second call returns non-empty
			},
			wantErr: false,
		},
		{
			name: "720p video without audio",
			responses: map[string]executor.MockResponse{
				"ffprobe": {
					Output: []byte(`{"streams":[{"width":1280,"height":720,"avg_frame_rate":"25/1"}]}`),
					Err:    nil,
				},
			},
			wantInfo: VideoInfo{
				Width:    1280,
				Height:   720,
				FPS:      25.0,
				HasAudio: false, // No audio stream returned
			},
			wantErr: false,
		},
		{
			name: "29.97 fps NTSC video",
			responses: map[string]executor.MockResponse{
				"ffprobe": {
					Output: []byte(`{"streams":[{"width":1920,"height":1080,"avg_frame_rate":"30000/1001"}]}`),
					Err:    nil,
				},
			},
			wantInfo: VideoInfo{
				Width:    1920,
				Height:   1080,
				FPS:      29.97002997002997,
				HasAudio: false,
			},
			wantErr: false,
		},
		{
			name: "ffprobe command fails",
			responses: map[string]executor.MockResponse{
				"ffprobe": {
					Output: nil,
					Err:    errors.New("ffprobe not found"),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid JSON response",
			responses: map[string]executor.MockResponse{
				"ffprobe": {
					Output: []byte(`invalid json`),
					Err:    nil,
				},
			},
			wantErr: true,
		},
		{
			name: "no video stream found",
			responses: map[string]executor.MockResponse{
				"ffprobe": {
					Output: []byte(`{"streams":[]}`),
					Err:    nil,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				// For error cases, use simple mock that fails
				mock := executor.NewMockExecutor()
				mock.Responses = tt.responses

				_, err := InputWithExecutor(context.Background(), "test.mp4", mock)

				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			// For success cases, use custom executor with sequential responses
			customMock := &customMockExecutor{
				videoResponse: tt.responses["ffprobe"],
				audioResponse: executor.MockResponse{
					Output: []byte("0"), // Audio stream exists by default
					Err:    nil,
				},
			}

			// Adjust audio response based on test case
			if !tt.wantInfo.HasAudio {
				customMock.audioResponse.Output = []byte("") // No audio
			}

			gotInfo, err := InputWithExecutor(context.Background(), "test.mp4", customMock)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotInfo.Width != tt.wantInfo.Width {
				t.Errorf("Width: got %d, want %d", gotInfo.Width, tt.wantInfo.Width)
			}
			if gotInfo.Height != tt.wantInfo.Height {
				t.Errorf("Height: got %d, want %d", gotInfo.Height, tt.wantInfo.Height)
			}
			if gotInfo.FPS != tt.wantInfo.FPS {
				t.Errorf("FPS: got %f, want %f", gotInfo.FPS, tt.wantInfo.FPS)
			}
			if gotInfo.HasAudio != tt.wantInfo.HasAudio {
				t.Errorf("HasAudio: got %v, want %v", gotInfo.HasAudio, tt.wantInfo.HasAudio)
			}
		})
	}
}

// customMockExecutor handles the two sequential calls (video probe, audio probe)
type customMockExecutor struct {
	videoResponse executor.MockResponse
	audioResponse executor.MockResponse
}

func (m *customMockExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, *executor.Usage, error) {
	return m.ExecuteWithProgress(ctx, nil, name, args...)
}

func (m *customMockExecutor) ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, *executor.Usage, error) {
	if progress != nil {
		close(progress)
	}
	for _, arg := range args {
		if arg == "v:0" {
			return m.videoResponse.Output, m.videoResponse.Usage, m.videoResponse.Err
		}
		if arg == "a:0" {
			return m.audioResponse.Output, m.audioResponse.Usage, m.audioResponse.Err
		}
	}
	return nil, nil, fmt.Errorf("unexpected call to Execute: %s %v", name, args)
}
