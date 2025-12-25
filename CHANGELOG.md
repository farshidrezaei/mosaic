# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Progress Reporting**: Real-time encoding updates via `ProgressHandler` and `ProgressInfo` struct.
- **Functional Options Pattern**: Flexible configuration for `EncodeHls` and `EncodeDash` using `WithThreads`, `WithGPU`, `WithLogLevel`, and `WithLogger`.
- **Multi-GPU Hardware Acceleration**: Support for NVIDIA NVENC, Intel/AMD VAAPI, and Apple VideoToolbox backends.
- **Structured Logging**: Integrated `log/slog` for internal library logging, customizable via `WithLogger`.
- **New Examples**:
    - `examples/advanced_dash`: Demonstrates hardware acceleration and custom logging.
    - `examples/multi_gpu`: Showcases different hardware acceleration backends.
- **Improved Testing**: Achieved 100% test coverage on core logic with a new mocked command executor and comprehensive unit tests.

### Changed
- **API Refactor**: `EncodeHls` and `EncodeDash` now accept variadic `Option` arguments.
- **Documentation**: Major updates to `README.md`, `ROADMAP.md`, and `STRUCTURE.md` to reflect new features and project layout.
- **Godocs**: Enhanced documentation for all public APIs across all packages.
- **Internal Logic**: Refined GOP calculation and FFmpeg command construction for better stability and performance.

### Fixed
- **Race Condition**: Fixed a race condition in the executor's progress reading logic.
- **Linting**: Resolved multiple `fieldalignment`, `shadow`, and `staticcheck` issues across the codebase.
- **Error Handling**: Improved error reporting in `RealCommandExecutor` by capturing `stderr`.
