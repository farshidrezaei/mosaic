# Mosaic Package Structure

This is a quick repository map. For deeper architecture notes, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Top-Level Layout

```text
mosaic/
├── .github/workflows/go.yml      # CI, when present
├── AGENTS.md                     # AI agent instructions
├── CHANGELOG.md                  # release notes
├── CONTRIBUTING.md               # human contribution workflow
├── README.md                     # user-facing entry point
├── ROADMAP.md                    # planned work
├── SECURITY.md                   # security reporting policy
├── STRUCTURE.md                  # quick package map
├── SUPPORT.md                    # support and bug-report guidance
├── encode.go                     # public encode orchestration API
├── job.go                        # public Job/Profile/Progress types
├── orientation.go                # orientation normalization helpers
├── config/
│   ├── profiles.go               # profiles and GPU constants
│   └── profiles_test.go
├── docs/
│   ├── API.md
│   ├── ARCHITECTURE.md
│   ├── ENCODING.md
│   ├── TESTING.md
│   └── TROUBLESHOOTING.md
├── probe/
│   ├── probe.go                  # FFprobe integration and metadata parsing
│   ├── probe_test.go
│   └── probe_integration_test.go
├── ladder/
│   ├── types.go                  # Rendition type
│   ├── ladder.go                 # aspect-preserving ladder generation
│   └── ladder_test.go
├── optimize/
│   ├── cost.go                   # bitrate cap rules
│   ├── optimize.go               # bitrate optimization and rung trimming
│   └── optimize_test.go
├── encoder/
│   ├── common.go                 # shared encoder helpers
│   ├── hls_cmaf.go               # HLS CMAF command generation
│   ├── dash_cmaf.go              # DASH CMAF command generation
│   └── *_test.go
├── internal/executor/
│   ├── executor.go               # real command executor
│   ├── mock.go                   # mock executor
│   └── executor_test.go
└── examples/
    ├── simple_hls/
    ├── advanced_dash/
    └── multi_gpu/
```

## Runtime Flow

```text
Job
 └─ encode.go
    ├─ prepareInputForEncoding
    │  └─ optional NormalizeVideoOrientation
    ├─ probe.InputWithExecutor
    │  └─ ffprobe metadata and audio detection
    ├─ ladder.Build
    │  └─ aspect-preserving quality rungs
    ├─ optimize.Apply
    │  └─ bitrate cap and rung trimming
    └─ encoder.Encode{HLS|DASH}CMAFWithExecutor
       └─ output directory creation + ffmpeg execution
```

## Package Responsibilities

- `probe`: source introspection via FFprobe.
- `ladder`: aspect-preserving initial rendition ladder generation.
- `optimize`: bitrate post-processing and redundant rung trimming.
- `encoder`: FFmpeg command assembly for HLS/DASH CMAF.
- `internal/executor`: command execution abstraction and mocks.
- `config`: profile and GPU backend constants.
- root package `mosaic`: public API, option wiring, and orchestration.

## Documentation Map

- `README.md`: primary user documentation.
- `docs/API.md`: public API details.
- `docs/ARCHITECTURE.md`: runtime and package design.
- `docs/ENCODING.md`: ladder, orientation, and FFmpeg behavior.
- `docs/TESTING.md`: test and smoke-test guidance.
- `docs/TROUBLESHOOTING.md`: operational debugging.
- `SUPPORT.md`: support and bug-report guidance.
- `SECURITY.md`: security reporting policy.
- `AGENTS.md`: AI coding-agent instructions.
