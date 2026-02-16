# Mosaic Package Structure

This document reflects the current repository/package layout and runtime flow.

## Top-Level Layout

```text
mosaic/
├── .github/workflows/go.yml      # CI (build/test/lint)
├── .golangci.yml                 # linter config
├── encode.go                     # public orchestration API
├── job.go                        # public Job/Profile/Progress types
├── config/
│   ├── profiles.go
│   └── profiles_test.go
├── probe/
│   ├── probe.go
│   ├── probe_test.go
│   └── probe_integration_test.go
├── ladder/
│   ├── types.go
│   ├── ladder.go
│   └── ladder_test.go
├── optimize/
│   ├── cost.go
│   ├── optimize.go
│   └── optimize_test.go
├── encoder/
│   ├── common.go
│   ├── hls_cmaf.go
│   ├── dash_cmaf.go
│   └── *_test.go
├── internal/executor/
│   ├── executor.go
│   ├── mock.go
│   └── executor_test.go
├── examples/
│   ├── simple_hls/
│   ├── advanced_dash/
│   └── multi_gpu/
├── README.md
├── CONTRIBUTING.md
├── ROADMAP.md
└── CHANGELOG.md
```

## Runtime Flow

```text
Job
 └─ encode.go
    ├─ probe.InputWithExecutor
    │  └─ ffprobe (video stream + audio stream)
    │     └─ width/height/fps/audio + orientation metadata
    ├─ ladder.Build
    │  └─ base ladder from effective display dimensions
    ├─ optimize.Apply
    │  └─ bitrate cap + rung trimming
    └─ encoder.Encode{HLS|DASH}CMAFWithExecutor
       └─ ffmpeg command construction + execution
```

## Package Responsibilities

- `probe`: source introspection via FFprobe.
- `ladder`: initial rendition ladder generation.
- `optimize`: post-processing of ladder bitrates/rungs.
- `encoder`: FFmpeg command assembly for HLS/DASH CMAF.
- `internal/executor`: command execution abstraction and mocks.
- `config`: profile and GPU backend constants.
- root package (`mosaic`): user-facing API and option wiring.

## Notes

- Dependency direction is intentionally simple: public API orchestrates lower-level packages.
- Test files are colocated with production code.
- Orientation support is handled during probing and ladder selection.
