// Package preview provides a local HTTP development server with an integrated HTML5 player
// for testing and inspecting generated HLS and DASH CMAF streams, rendition switches, and thumbnail cues.
package preview

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Server represents the local preview HTTP server.
type Server struct {
	httpServer *http.Server
	listener   net.Listener
	Dir        string
	Port       int
}

// NewServer creates a new preview server instance serving the specified directory.
func NewServer(dir string, port int) *Server {
	if port <= 0 {
		port = 8080
	}
	return &Server{
		Dir:  dir,
		Port: port,
	}
}

// DetectStreams scans the served directory for HLS and DASH entrypoints and thumbnail files.
func DetectStreams(dir string) (hasHLS bool, hasDASH bool, hasThumbnails bool) {
	if _, err := os.Stat(filepath.Join(dir, "master.m3u8")); err == nil {
		hasHLS = true
	} else if _, err := os.Stat(filepath.Join(dir, "stream_0.m3u8")); err == nil {
		hasHLS = true
	}

	if _, err := os.Stat(filepath.Join(dir, "manifest.mpd")); err == nil {
		hasDASH = true
	}

	if _, err := os.Stat(filepath.Join(dir, "thumbnails.vtt")); err == nil {
		hasThumbnails = true
	}

	return
}

// Handler returns the http.Handler serving the preview UI and media files with CORS enabled.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(s.Dir))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			hasHLS, hasDASH, hasThumbnails := DetectStreams(s.Dir)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(renderHTML(hasHLS, hasDASH, hasThumbnails)))
			return
		}

		fileServer.ServeHTTP(w, r)
	})

	return mux
}

// Start starts listening on the configured port.
func (s *Server) Start(ctx ...context.Context) error {
	addr := fmt.Sprintf(":%d", s.Port)
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}
	lc := net.ListenConfig{}
	listener, err := lc.Listen(c, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	s.listener = listener
	s.Port = listener.Addr().(*net.TCPAddr).Port

	s.httpServer = &http.Server{
		Handler:      s.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		_ = s.httpServer.Serve(listener)
	}()

	return nil
}

// URL returns the base URL of the running preview server.
func (s *Server) URL() string {
	if s.Port == 0 {
		return ""
	}
	return fmt.Sprintf("http://localhost:%d", s.Port)
}

// Shutdown gracefully stops the preview server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Serve runs the preview server blocking until context cancellation.
func Serve(ctx context.Context, dir string, port int) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve directory path: %w", err)
	}

	srv := NewServer(absDir, port)
	if err := srv.Start(); err != nil {
		return err
	}

	hasHLS, hasDASH, hasThumbnails := DetectStreams(absDir)
	streamType := "Unknown"
	if hasHLS && hasDASH {
		streamType = "HLS + DASH"
	} else if hasHLS {
		streamType = "HLS (fMP4)"
	} else if hasDASH {
		streamType = "DASH CMAF"
	}

	fmt.Printf("\n🎬 Mosaic Stream Preview Server\n")
	fmt.Printf("   Directory:    %s\n", absDir)
	fmt.Printf("   Stream Type:  %s\n", streamType)
	fmt.Printf("   Thumbnails:   %t\n", hasThumbnails)
	fmt.Printf("   Local URL:    %s\n\n", srv.URL())
	fmt.Printf("Press Ctrl+C to stop the preview server.\n")

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func renderHTML(hasHLS, hasDASH, hasThumbnails bool) string {
	initialStream := "master.m3u8"
	initialType := "hls"
	if !hasHLS && hasDASH {
		initialStream = "manifest.mpd"
		initialType = "dash"
	}

	hlsActive := "active"
	dashActive := ""
	if !hasHLS && hasDASH {
		hlsActive = ""
		dashActive = "active"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Mosaic Stream Preview</title>
  <script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
  <script src="https://cdn.dashjs.org/latest/dash.all.min.js"></script>
  <style>
    :root {
      --bg: #0f172a;
      --card: #1e293b;
      --accent: #38bdf8;
      --text: #f8fafc;
      --muted: #94a3b8;
      --border: #334155;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background-color: var(--bg);
      color: var(--text);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
      padding: 24px;
      display: flex;
      flex-direction: column;
      align-items: center;
      min-height: 100vh;
    }
    .container {
      width: 100%%;
      max-width: 1000px;
      display: flex;
      flex-direction: column;
      gap: 20px;
    }
    header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding-bottom: 12px;
      border-bottom: 1px solid var(--border);
    }
    h1 {
      font-size: 1.4rem;
      font-weight: 700;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .badge {
      background: #0284c7;
      color: white;
      font-size: 0.75rem;
      padding: 3px 8px;
      border-radius: 9999px;
      font-weight: 600;
    }
    .tabs {
      display: flex;
      gap: 8px;
    }
    .tab-btn {
      background: var(--card);
      border: 1px solid var(--border);
      color: var(--muted);
      padding: 6px 14px;
      border-radius: 6px;
      cursor: pointer;
      font-weight: 600;
      transition: all 0.2s;
    }
    .tab-btn.active {
      background: var(--accent);
      color: #0f172a;
      border-color: var(--accent);
    }
    .video-wrapper {
      position: relative;
      background: #000;
      border-radius: 12px;
      overflow: hidden;
      box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
      border: 1px solid var(--border);
      width: 100%%;
      max-height: 70vh;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    video {
      width: 100%%;
      max-height: 70vh;
      display: block;
      object-fit: contain;
    }
    .grid {
      display: grid;
      grid-template-columns: 2fr 1fr;
      gap: 16px;
    }
    @media (max-width: 768px) {
      .grid { grid-template-columns: 1fr; }
    }
    .card {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 16px;
    }
    .card-title {
      font-size: 0.9rem;
      font-weight: 600;
      color: var(--muted);
      margin-bottom: 12px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .stat-row {
      display: flex;
      justify-content: space-between;
      padding: 6px 0;
      border-bottom: 1px solid rgba(255, 255, 255, 0.05);
      font-size: 0.9rem;
    }
    .stat-row:last-child { border-bottom: none; }
    .stat-label { color: var(--muted); }
    .stat-value { font-family: monospace; font-weight: 600; color: var(--accent); }
    .select-wrapper {
      margin-top: 10px;
    }
    select {
      width: 100%%;
      background: #0f172a;
      color: var(--text);
      border: 1px solid var(--border);
      padding: 8px 12px;
      border-radius: 6px;
      outline: none;
      font-size: 0.9rem;
      cursor: pointer;
    }
  </style>
</head>
<body>
  <div class="container">
    <header>
      <h1>Mosaic Preview <span class="badge">Live</span></h1>
      <div class="tabs">
        <button class="tab-btn %s" id="btn-hls" onclick="loadStream('master.m3u8', 'hls')">HLS</button>
        <button class="tab-btn %s" id="btn-dash" onclick="loadStream('manifest.mpd', 'dash')">DASH</button>
      </div>
    </header>

    <div class="video-wrapper">
      <video id="player" controls playsinline></video>
    </div>

    <div class="grid">
      <div class="card">
        <div class="card-title">Stream Statistics</div>
        <div class="stat-row">
          <span class="stat-label">Source URL</span>
          <span class="stat-value" id="stat-url">%s</span>
        </div>
        <div class="stat-row">
          <span class="stat-label">Active Resolution</span>
          <span class="stat-value" id="stat-res">-</span>
        </div>
        <div class="stat-row">
          <span class="stat-label">Current Bitrate</span>
          <span class="stat-value" id="stat-bitrate">-</span>
        </div>
        <div class="stat-row">
          <span class="stat-label">Buffer Health</span>
          <span class="stat-value" id="stat-buffer">0.0s</span>
        </div>
      </div>

      <div class="card">
        <div class="card-title">Quality Selector</div>
        <div class="select-wrapper">
          <select id="quality-selector" onchange="changeQuality(this.value)">
            <option value="-1">Auto (ABR Engine)</option>
          </select>
        </div>
        <div style="margin-top: 14px; font-size: 0.8rem; color: var(--muted);">
          Thumbnails (WebVTT): <strong>%t</strong>
        </div>
      </div>
    </div>
  </div>

  <script>
    const video = document.getElementById('player');
    const statRes = document.getElementById('stat-res');
    const statBitrate = document.getElementById('stat-bitrate');
    const statBuffer = document.getElementById('stat-buffer');
    const statUrl = document.getElementById('stat-url');
    const qualitySelector = document.getElementById('quality-selector');

    let hlsInstance = null;
    let dashInstance = null;

    function resetQualityOptions() {
      qualitySelector.innerHTML = '<option value="-1">Auto (ABR Engine)</option>';
    }

    function loadStream(url, type) {
      statUrl.textContent = url;
      document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
      const btn = type === 'hls' ? document.getElementById('btn-hls') : document.getElementById('btn-dash');
      if (btn) btn.classList.add('active');

      if (hlsInstance) { hlsInstance.destroy(); hlsInstance = null; }
      if (dashInstance) { dashInstance.reset(); dashInstance = null; }
      resetQualityOptions();

      if (type === 'hls') {
        if (Hls.isSupported()) {
          hlsInstance = new Hls();
          hlsInstance.loadSource(url);
          hlsInstance.attachMedia(video);
          hlsInstance.on(Hls.Events.MANIFEST_PARSED, (e, data) => {
            data.levels.forEach((lvl, idx) => {
              const opt = document.createElement('option');
              opt.value = idx;
              opt.textContent = lvl.height + 'p (' + Math.round(lvl.bitrate/1000) + ' kbps)';
              qualitySelector.appendChild(opt);
            });
            video.play().catch(() => {});
          });
          hlsInstance.on(Hls.Events.LEVEL_SWITCHED, (e, data) => {
            const lvl = hlsInstance.levels[data.level];
            if (lvl) {
              statRes.textContent = lvl.width + 'x' + lvl.height;
              statBitrate.textContent = Math.round(lvl.bitrate/1000) + ' kbps';
            }
          });
        } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
          video.src = url;
        }
      } else if (type === 'dash') {
        dashInstance = dashjs.MediaPlayer().create();
        dashInstance.initialize(video, url, true);
      }
    }

    function changeQuality(val) {
      if (hlsInstance) {
        hlsInstance.currentLevel = parseInt(val);
      }
    }

    setInterval(() => {
      if (video.buffered.length > 0) {
        const bufferedEnd = video.buffered.end(video.buffered.length - 1);
        const remaining = Math.max(0, bufferedEnd - video.currentTime);
        statBuffer.textContent = remaining.toFixed(1) + 's';
      }
      if (video.videoWidth > 0 && !hlsInstance) {
        statRes.textContent = video.videoWidth + 'x' + video.videoHeight;
      }
    }, 1000);

    // Initial load
    loadStream('%s', '%s');
  </script>
</body>
</html>`, hlsActive, dashActive, initialStream, hasThumbnails, initialStream, initialType)
}
