package preview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDetectStreams(t *testing.T) {
	tmpDir := t.TempDir()

	hasHLS, hasDASH, hasThumb := DetectStreams(tmpDir)
	if hasHLS || hasDASH || hasThumb {
		t.Errorf("expected all false on empty dir, got hls=%t, dash=%t, thumb=%t", hasHLS, hasDASH, hasThumb)
	}

	// Create dummy files
	_ = os.WriteFile(filepath.Join(tmpDir, "master.m3u8"), []byte("#EXTM3U"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "manifest.mpd"), []byte("<MPD></MPD>"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "thumbnails.vtt"), []byte("WEBVTT"), 0o644)

	hasHLS, hasDASH, hasThumb = DetectStreams(tmpDir)
	if !hasHLS || !hasDASH || !hasThumb {
		t.Errorf("expected all true, got hls=%t, dash=%t, thumb=%t", hasHLS, hasDASH, hasThumb)
	}
}

func TestServer_Handler(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "master.m3u8"), []byte("#EXTM3U\n#EXT-X-VERSION:7"), 0o644)

	srv := NewServer(tmpDir, 0)
	handler := srv.Handler()

	// Test HTML root
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for root, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("missing CORS header")
	}
	if !strings.Contains(rec.Body.String(), "Mosaic Stream Preview") {
		t.Errorf("expected HTML title in body, got: %s", rec.Body.String())
	}

	// Test static file serving
	reqFile := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/master.m3u8", nil)
	recFile := httptest.NewRecorder()
	handler.ServeHTTP(recFile, reqFile)

	if recFile.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for file, got %d", recFile.Code)
	}
	if !strings.Contains(recFile.Body.String(), "#EXTM3U") {
		t.Errorf("expected file content, got: %s", recFile.Body.String())
	}

	// Test OPTIONS preflight
	reqOptions := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/master.m3u8", nil)
	recOptions := httptest.NewRecorder()
	handler.ServeHTTP(recOptions, reqOptions)

	if recOptions.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for OPTIONS, got %d", recOptions.Code)
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	srv := NewServer(tmpDir, 0) // Port 0 allows OS to pick free port

	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	if srv.URL() == "" {
		t.Errorf("expected non-empty URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("failed to shutdown server: %v", err)
	}
}
