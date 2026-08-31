# Mosaic Package Structure

This is a quick repository map. For deeper architecture notes, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Top-Level Layout

```text
mosaic/
├── .github/workflows/go.yml      # CI workflows
├── AGENTS.md                     # AI agent instructions and quality contracts
├── CHANGELOG.md                  # Release notes and version history
├── CONTRIBUTING.md               # Human contribution workflow
├── README.md                     # User-facing entry point and highlights
├── ROADMAP.md                    # Roadmap and future direction
├── SECURITY.md                   # Security reporting policy
├── STRUCTURE.md                  # Quick package map
├── SUPPORT.md                    # Support and bug-report guidance
├── encode.go                     # Public encode orchestration API & Options
├── job.go                        # Public Job/Profile/Progress types
├── orientation.go                # Orientation normalization helpers
├── cmd/mosaic/
│   └── main.go                   # CLI entrypoint (encoding + preview subcommand)
├── config/
│   ├── profiles.go               # Profiles, VideoCodec, and GPU constants
│   └── profiles_test.go
├── docs/                         # GitHub Pages documentation portal (Docsify)
│   ├── API.md
│   ├── ARCHITECTURE.md
│   ├── ENCODING.md
│   ├── EXAMPLES.md
│   ├── options.md
│   ├── orientation.md
│   ├── quickstart.md
│   ├── installation.md
│   ├── TESTING.md
│   └── TROUBLESHOOTING.md
├── probe/
│   ├── probe.go                  # FFprobe integration and metadata parsing
│   └── probe_test.go
├── ladder/
│   ├── types.go                  # Rendition type
│   ├── ladder.go                 # Aspect-preserving ladder generation
│   └── ladder_test.go
├── optimize/
│   ├── cost.go                   # Bitrate cap rules
│   ├── optimize.go               # Bitrate optimization and rung trimming
│   └── optimize_test.go
├── encoder/
│   ├── common.go                 # Shared encoder helpers
│   ├── codec.go                  # Video encoder resolution (x264, x265, SVT-AV1, GPU)
│   ├── hls_cmaf.go               # HLS CMAF/TS command generation & AES-128
│   ├── dash_cmaf.go              # DASH CMAF command generation
│   └── *_test.go
├── thumbnail/
│   ├── thumbnail.go              # Sprite sheet generation and WebVTT cue builder
│   └── thumbnail_test.go
├── preview/
│   ├── server.go                 # Local dark-mode web player (HLS.js / Dash.js)
│   └── server_test.go
├── subtitles/
│   ├── subtitles.go              # SRT-to-VTT converter and HLS/DASH manifest injector
│   └── subtitles_test.go
├── watermark/
│   ├── watermark.go              # Dynamic overlay coordinate builder and opacity
│   └── watermark_test.go
├── encryption/
│   ├── encryption.go             # AES-128 key generation and enc.keyinfo setup
│   └── encryption_test.go
├── storage/
│   ├── storage.go                # Zero-dependency S3 / MinIO / R2 uploader (AWS SigV4)
│   └── storage_test.go
├── internal/executor/
│   ├── executor.go               # Real command executor with usage stats
│   ├── mock.go                   # Mock executor for fast isolated testing
│   └── executor_test.go
└── examples/
    ├── simple_hls/
    ├── advanced_dash/
    ├── thumbnails_and_preview/
    ├── watermark_and_subtitles/
    ├── encryption_aes128/
    ├── nextgen_av1_hevc/
    ├── s3_cloud_upload/
    └── orientation_normalization/
```

## Runtime Flow

```text
Job
 └─ encode.go
    ├─ prepareInputForEncoding (optional NormalizeVideoOrientation)
    ├─ probe.InputWithExecutor (FFprobe metadata and audio streams)
    ├─ ladder.Build (Aspect-preserving quality rungs)
    ├─ optimize.Apply (Bitrate caps and redundant rung trimming)
    ├─ encryption.SetupKeyInfo (Optional AES-128 key generation)
    ├─ encoder.Encode{HLS|DASH}CMAFWithExecutor (FFmpeg single-pass packaging)
    ├─ thumbnail.GenerateWithExecutor (Optional storyboard sprite + VTT)
    ├─ subtitles.ProcessTracks (Optional SRT-to-VTT + manifest injection)
    └─ storage.UploadDirectory (Optional S3 / MinIO direct sync)
```

## Package Responsibilities

- `probe`: Source introspection via FFprobe.
- `ladder`: Aspect-preserving initial rendition ladder generation.
- `optimize`: Bitrate post-processing and redundant rung trimming.
- `encoder`: FFmpeg command assembly for HLS/DASH CMAF, Codecs, and Encryption.
- `thumbnail`: Automatic storyboard sprite sheet and WebVTT cues generation.
- `preview`: Local devtools HTTP preview server with web player.
- `subtitles`: SRT-to-VTT subtitle converter and manifest injector.
- `watermark`: Responsive logo/watermark overlay filter graph generation.
- `encryption`: HLS AES-128 key generation and keyinfo configuration.
- `storage`: Pure Go AWS SigV4 direct upload to S3, MinIO, or Cloudflare R2.
- `internal/executor`: Command execution abstraction and test mocks.
- `config`: Profile, Codec, and GPU backend constants.
- root package `mosaic`: Public API, option wiring, and orchestration.
