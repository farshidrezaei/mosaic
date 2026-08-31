package subtitles

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertSRTToVTT(t *testing.T) {
	srt := `1
00:00:01,500 --> 00:00:04,200
Hello World!

2
00:00:05,000 --> 00:00:08,000
Second subtitle line.`

	vtt := ConvertSRTToVTT(srt)
	if !strings.HasPrefix(vtt, "WEBVTT\n\n") {
		t.Errorf("expected WEBVTT header, got:\n%s", vtt)
	}

	if !strings.Contains(vtt, "00:00:01.500 --> 00:00:04.200") {
		t.Errorf("timestamp was not converted properly: %s", vtt)
	}
}

func TestGenerateHLSPlaylist(t *testing.T) {
	playlist := GenerateHLSPlaylist("sub_en.vtt", 30.5)
	if !strings.Contains(playlist, "#EXTINF:30.500,") {
		t.Errorf("expected target duration in playlist:\n%s", playlist)
	}
	if !strings.Contains(playlist, "sub_en.vtt") {
		t.Errorf("expected sub_en.vtt in playlist:\n%s", playlist)
	}
}

func TestInjectHLSSubtitles(t *testing.T) {
	master := `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080
stream_0.m3u8`

	tracks := []Track{
		{Language: "en", Label: "English", Default: true},
		{Language: "fa", Label: "فارسی", Default: false},
	}

	injected := InjectHLSSubtitles(master, tracks)

	if !strings.Contains(injected, `#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subs",NAME="English",DEFAULT=YES`) {
		t.Errorf("missing English subtitle tag:\n%s", injected)
	}
	if !strings.Contains(injected, `#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subs",NAME="فارسی",DEFAULT=NO`) {
		t.Errorf("missing Persian subtitle tag:\n%s", injected)
	}
	if !strings.Contains(injected, `SUBTITLES="subs"`) {
		t.Errorf("missing SUBTITLES attribute on stream inf:\n%s", injected)
	}
}

func TestInjectDASHSubtitles(t *testing.T) {
	mpd := `<?xml version="1.0"?>
<MPD>
  <Period>
    <AdaptationSet id="0" contentType="video"></AdaptationSet>
  </Period>
</MPD>`

	tracks := []Track{
		{Language: "en", Label: "English", Default: true},
	}

	injected := InjectDASHSubtitles(mpd, tracks)
	if !strings.Contains(injected, `contentType="text" mimeType="text/vtt" lang="en"`) {
		t.Errorf("missing subtitle AdaptationSet in MPD:\n%s", injected)
	}
	if !strings.Contains(injected, `<BaseURL>sub_en.vtt</BaseURL>`) {
		t.Errorf("missing BaseURL in MPD:\n%s", injected)
	}
}

func TestProcessTracks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test SRT file
	srtPath := filepath.Join(tmpDir, "input.srt")
	srtContent := "1\n00:00:01,000 --> 00:00:03,000\nTesting Subtitle\n"
	if err := os.WriteFile(srtPath, []byte(srtContent), 0o644); err != nil {
		t.Fatalf("failed to write srt file: %v", err)
	}

	// Create mock master.m3u8 and manifest.mpd
	masterPath := filepath.Join(tmpDir, "master.m3u8")
	_ = os.WriteFile(masterPath, []byte("#EXTM3U\n#EXT-X-VERSION:7\n#EXT-X-STREAM-INF:BANDWIDTH=1000\nstream_0.m3u8\n"), 0o644)
	mpdPath := filepath.Join(tmpDir, "manifest.mpd")
	_ = os.WriteFile(mpdPath, []byte("<MPD><Period></Period></MPD>"), 0o644)

	tracks := []Track{
		{Path: srtPath, Language: "fa", Label: "فارسی", Default: true},
	}

	err := ProcessTracks(context.Background(), tracks, tmpDir, 10.0)
	if err != nil {
		t.Fatalf("ProcessTracks failed: %v", err)
	}

	vttPath := filepath.Join(tmpDir, "sub_fa.vtt")
	if _, err := os.Stat(vttPath); os.IsNotExist(err) {
		t.Errorf("expected sub_fa.vtt to be generated")
	}

	vttPlaylist := filepath.Join(tmpDir, "sub_fa.m3u8")
	if _, err := os.Stat(vttPlaylist); os.IsNotExist(err) {
		t.Errorf("expected sub_fa.m3u8 to be generated")
	}
}
