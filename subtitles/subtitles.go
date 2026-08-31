// Package subtitles provides WebVTT and SRT subtitle ingestion, conversion, and playlist injection
// for HLS master playlists and MPEG-DASH manifests.
package subtitles

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Track defines a subtitle track to be packaged alongside HLS or DASH streams.
type Track struct {
	Path     string // Input subtitle file path (.vtt or .srt)
	Language string // ISO language code (e.g. "en", "fa")
	Label    string // Display label in video player (e.g. "English", "فارسی")
	Default  bool   // Whether this track is the default active subtitle
	Forced   bool   // Whether this is a forced subtitle track
}

// ConvertSRTToVTT converts an SRT subtitle string into a compliant WebVTT string.
func ConvertSRTToVTT(srtContent string) string {
	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")

	// Regex for SRT timestamp line: 00:01:20,000 --> 00:01:23,456
	timestampRegex := regexp.MustCompile(`(\d{2}:\d{2}:\d{2}),(\d{3})\s*-->\s*(\d{2}:\d{2}:\d{2}),(\d{3})`)

	scanner := bufio.NewScanner(strings.NewReader(srtContent))
	for scanner.Scan() {
		line := scanner.Text()
		// Replace comma with dot in timestamps
		if timestampRegex.MatchString(line) {
			line = timestampRegex.ReplaceAllString(line, "$1.$2 --> $3.$4")
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// GenerateHLSPlaylist generates a single-file WebVTT HLS playlist for a subtitle track.
func GenerateHLSPlaylist(vttFilename string, duration float64) string {
	targetDuration := int(duration) + 1
	if targetDuration < 1 {
		targetDuration = 10
	}

	return fmt.Sprintf(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:%d
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:%.3f,
%s
#EXT-X-ENDLIST
`, targetDuration, duration, vttFilename)
}

// InjectHLSSubtitles updates an HLS master playlist content to reference subtitle media groups.
func InjectHLSSubtitles(masterContent string, tracks []Track) string {
	if len(tracks) == 0 {
		return masterContent
	}

	var subMediaLines strings.Builder
	for i, track := range tracks {
		lang := track.Language
		if lang == "" {
			lang = "und"
		}
		label := track.Label
		if label == "" {
			label = fmt.Sprintf("Subtitle %d", i+1)
		}

		defaultStr := "NO"
		autoSelectStr := "NO"
		if track.Default {
			defaultStr = "YES"
			autoSelectStr = "YES"
		}

		forcedStr := "NO"
		if track.Forced {
			forcedStr = "YES"
		}

		playlistName := fmt.Sprintf("sub_%s.m3u8", lang)
		_, _ = fmt.Fprintf(&subMediaLines,
			"#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\",NAME=\"%s\",DEFAULT=%s,AUTOSELECT=%s,FORCED=%s,LANGUAGE=\"%s\",URI=\"%s\"\n",
			label, defaultStr, autoSelectStr, forcedStr, lang, playlistName,
		)
	}

	// Insert subtitle media lines after the header (#EXTM3U)
	lines := strings.Split(masterContent, "\n")
	var result strings.Builder
	injected := false

	for _, line := range lines {
		result.WriteString(line)
		result.WriteString("\n")

		if !injected && (strings.HasPrefix(line, "#EXT-X-VERSION") || strings.HasPrefix(line, "#EXTM3U")) {
			result.WriteString(subMediaLines.String())
			injected = true
		}
	}

	// Add SUBTITLES="subs" attribute to all #EXT-X-STREAM-INF lines if not already present
	finalContent := result.String()
	streamInfRegex := regexp.MustCompile(`(#EXT-X-STREAM-INF:[^\n]+)`)
	finalContent = streamInfRegex.ReplaceAllStringFunc(finalContent, func(match string) string {
		if !strings.Contains(match, "SUBTITLES=") {
			return match + ",SUBTITLES=\"subs\""
		}
		return match
	})

	return finalContent
}

// InjectDASHSubtitles injects Subtitle AdaptationSets into a DASH manifest (.mpd).
func InjectDASHSubtitles(manifestContent string, tracks []Track) string {
	if len(tracks) == 0 {
		return manifestContent
	}

	var subAdaptationSets strings.Builder
	for i, track := range tracks {
		lang := track.Language
		if lang == "" {
			lang = "und"
		}
		label := track.Label
		if label == "" {
			label = fmt.Sprintf("Subtitle %d", i+1)
		}

		vttFilename := fmt.Sprintf("sub_%s.vtt", lang)
		setID := 100 + i

		_, _ = fmt.Fprintf(&subAdaptationSets, `    <AdaptationSet id="%d" contentType="text" mimeType="text/vtt" lang="%s">
      <Role schemeIdUri="urn:mpeg:dash:role:2011" value="subtitle"/>
      <Label>%s</Label>
      <Representation id="sub_%s" bandwidth="256">
        <BaseURL>%s</BaseURL>
      </Representation>
    </AdaptationSet>
`, setID, lang, label, lang, vttFilename)
	}

	// Inject right before </Period>
	if strings.Contains(manifestContent, "</Period>") {
		return strings.Replace(manifestContent, "</Period>", subAdaptationSets.String()+"  </Period>", 1)
	}

	return manifestContent
}

// ProcessTracks copies/converts subtitle files into outputDir and generates corresponding HLS playlists.
func ProcessTracks(_ context.Context, tracks []Track, outDir string, duration float64) error {
	if len(tracks) == 0 {
		return nil
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create subtitles directory: %w", err)
	}

	for _, track := range tracks {
		if track.Path == "" {
			continue
		}

		content, err := os.ReadFile(track.Path)
		if err != nil {
			return fmt.Errorf("read subtitle file %s: %w", track.Path, err)
		}

		vttText := string(content)
		if strings.HasSuffix(strings.ToLower(track.Path), ".srt") || !strings.HasPrefix(strings.TrimSpace(vttText), "WEBVTT") {
			vttText = ConvertSRTToVTT(vttText)
		}

		lang := track.Language
		if lang == "" {
			lang = "und"
		}

		vttFilename := fmt.Sprintf("sub_%s.vtt", lang)
		destVTTPath := filepath.Join(outDir, vttFilename)
		if err := os.WriteFile(destVTTPath, []byte(vttText), 0o644); err != nil {
			return fmt.Errorf("write subtitle vtt %s: %w", destVTTPath, err)
		}

		// Generate HLS subtitle playlist
		m3u8Text := GenerateHLSPlaylist(vttFilename, duration)
		destM3U8Path := filepath.Join(outDir, fmt.Sprintf("sub_%s.m3u8", lang))
		if err := os.WriteFile(destM3U8Path, []byte(m3u8Text), 0o644); err != nil {
			return fmt.Errorf("write subtitle playlist %s: %w", destM3U8Path, err)
		}
	}

	// Inject into master.m3u8 if present
	masterPath := filepath.Join(outDir, "master.m3u8")
	if masterBytes, err := os.ReadFile(masterPath); err == nil {
		newMaster := InjectHLSSubtitles(string(masterBytes), tracks)
		_ = os.WriteFile(masterPath, []byte(newMaster), 0o644)
	}

	// Inject into manifest.mpd if present
	mpdPath := filepath.Join(outDir, "manifest.mpd")
	if mpdBytes, err := os.ReadFile(mpdPath); err == nil {
		newMPD := InjectDASHSubtitles(string(mpdBytes), tracks)
		_ = os.WriteFile(mpdPath, []byte(newMPD), 0o644)
	}

	return nil
}
