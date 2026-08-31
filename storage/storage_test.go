package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestContentTypeForFile(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"master.m3u8", "application/vnd.apple.mpegurl"},
		{"manifest.mpd", "application/dash+xml"},
		{"seg_0_0.m4s", "video/iso.segment"},
		{"seg_0_0.ts", "video/mp2t"},
		{"thumbnails.vtt", "text/vtt"},
		{"thumbnails_0.jpg", "image/jpeg"},
		{"enc.key", "application/octet-stream"},
		{"other.bin", "application/octet-stream"},
	}

	for _, tt := range tests {
		ct := ContentTypeForFile(tt.filename)
		if ct != tt.expected {
			t.Errorf("ContentTypeForFile(%s) = %s, expected %s", tt.filename, ct, tt.expected)
		}
	}
}

func TestUploadDirectory(t *testing.T) {
	var mu sync.Mutex
	uploadedFiles := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		mu.Lock()
		uploadedFiles[r.URL.Path] = r.Header.Get("Content-Type")
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "master.m3u8"), []byte("#EXTM3U"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "seg_0_0.m4s"), []byte("video-data"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "thumbnails.vtt"), []byte("WEBVTT"), 0o644)

	cfg := S3Config{
		Endpoint:        server.URL,
		Bucket:          "mybucket",
		Region:          "us-east-1",
		AccessKey:       "test-access-key",
		SecretKey:       "test-secret-key",
		KeyPrefix:       "videos/123",
		ConcurrentFiles: 2,
	}

	err := UploadDirectory(context.Background(), tmpDir, cfg, server.Client())
	if err != nil {
		t.Fatalf("UploadDirectory failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	expectedPaths := []string{
		"/mybucket/videos/123/master.m3u8",
		"/mybucket/videos/123/seg_0_0.m4s",
		"/mybucket/videos/123/thumbnails.vtt",
	}

	for _, p := range expectedPaths {
		if _, ok := uploadedFiles[p]; !ok {
			t.Errorf("expected uploaded file at %s, but got %v", p, uploadedFiles)
		}
	}
}
