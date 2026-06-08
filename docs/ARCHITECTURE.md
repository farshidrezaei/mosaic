# Mosaic Architecture

Mosaic is intentionally small. The root package provides orchestration while lower-level packages own focused behavior.

## Runtime Flow

```text
Job
 └─ encode.go
    ├─ prepareInputForEncoding
    │  └─ optional orientation normalization
    ├─ probe.InputWithExecutor
    │  └─ ffprobe video stream + audio stream
    ├─ ladder.Build
    │  └─ aspect-preserving base ladder
    ├─ optimize.Apply
    │  └─ bitrate capping and rung trimming
    └─ encoder.Encode{HLS|DASH}CMAFWithExecutor
       └─ ffmpeg command assembly and execution
```

## Packages

### Root Package

Files:

```text
encode.go
job.go
orientation.go
```

Responsibilities:

- Public API.
- Functional option parsing.
- Profile mapping from public `mosaic.Profile` to internal `config.Profile`.
- End-to-end encode orchestration.
- Optional orientation normalization.

### config

Files:

```text
config/profiles.go
```

Responsibilities:

- Internal HLS/DASH profile configuration.
- GPU backend constants.

### probe

Files:

```text
probe/probe.go
```

Responsibilities:

- Run FFprobe.
- Parse video width, height, average FPS, audio presence, and rotation metadata.
- Expose display dimension helpers:
  - `DisplayWidth`
  - `DisplayHeight`
  - `IsPortrait`

Rotation metadata is read from FFprobe side data first, then `tags.rotate`.

### ladder

Files:

```text
ladder/types.go
ladder/ladder.go
```

Responsibilities:

- Generate initial ABR rungs from source display dimensions.
- Preserve source display aspect ratio.
- Keep dimensions even for H.264 and `yuv420p` compatibility.
- Avoid upscaling sources below 360p.

### optimize

Files:

```text
optimize/cost.go
optimize/optimize.go
```

Responsibilities:

- Cap bitrates based on rung height.
- Recompute VBV buffer size from max bitrate.
- Trim redundant rungs that are too close in height.

### encoder

Files:

```text
encoder/common.go
encoder/hls_cmaf.go
encoder/dash_cmaf.go
```

Responsibilities:

- Build FFmpeg arguments.
- Create output directory.
- Select video encoder from CPU/GPU options.
- Map audio streams when present.
- Generate HLS and DASH CMAF output.
- Parse progress messages.
- Clear video rotation metadata on output streams.

### internal/executor

Files:

```text
internal/executor/executor.go
internal/executor/mock.go
```

Responsibilities:

- Wrap command execution.
- Provide usage statistics.
- Capture command stderr in structured errors.
- Support progress streaming from FFmpeg stdout.
- Provide mocks for tests.

## Dependency Direction

```text
mosaic
 ├─ config
 ├─ encoder
 ├─ internal/executor
 ├─ ladder
 ├─ optimize
 └─ probe
```

Lower-level packages do not import the root package.

This keeps public API orchestration separate from command generation and makes package tests straightforward.

## External Processes

Mosaic depends on external binaries:

- `ffprobe`
- `ffmpeg`

All invocations go through `executor.CommandExecutor`.

This makes command behavior observable in tests without actually running FFmpeg.

## State and Side Effects

Side effects are intentionally narrow:

- Output directory creation.
- Temporary normalized input files in the OS temp directory when orientation normalization is enabled.
- FFmpeg output files under `Job.OutputDir`.
- External process execution.

Mosaic does not keep global encode state.

## Error Boundaries

Important error boundaries:

- Probe errors abort before ladder generation.
- Orientation normalization errors abort before final encoding.
- Output directory creation errors abort before FFmpeg encode.
- FFmpeg errors are wrapped as `ffmpeg HLS failed` or `ffmpeg DASH failed`.
- Command stderr is included through `executor.CommandError`.

## Testing Strategy

Most behavior is tested with mocks:

- Probe parsing.
- Ladder generation.
- Optimizer behavior.
- FFmpeg argument construction.
- Progress parsing.
- Orientation normalization command construction.

Integration tests and smoke tests can use real FFmpeg and FFprobe where available.
