# 🚀 Mosaic Roadmap to Stardom

To make `mosaic` the go-to Go package for video encoding, focus on **Developer Experience (DX)**, **Performance**, and **Modern Features**.

## 1. Developer Experience (DX) ⭐⭐⭐⭐⭐

*Crucial for adoption. If it's easy to use, people will use it.*

- [x] **Context Support**: Add `context.Context` to all long-running functions. Users need to be able to cancel encodings or set timeouts.
  - *Change*: `EncodeHls(ctx context.Context, job Job)`
- [ ] **Progress Reporting**: Video encoding takes time. Users need to know the status.
  - *Feature*: Add a callback or channel to report percentage/current segment.
- [ ] **Functional Options Pattern**: Make configuration flexible without breaking changes.
  - *Example*: `EncodeHls(job, WithGPU(), WithThreads(4))`
- [ ] **Structured Logging**: Allow users to plug in their own logger (slog, zap, logrus) instead of printing to stdout/stderr.

## 2. Killer Features 🚀

*What makes your library better than writing a shell script?*

- [ ] **Hardware Acceleration**: Support NVENC (NVIDIA), VAAPI (Intel/AMD), and VideoToolbox (macOS). This is a massive performance booster.
- [ ] **Modern Codecs**: Add H.265 (HEVC) and AV1 support.
- [ ] **Thumbnail Generation**: Auto-generate a sprite sheet or VTT thumbnails for the player seek bar.
- [ ] **Cloud Hooks**: Interfaces to upload segments directly to S3/GCS as they are created, rather than writing to disk first.

## 3. Reliability & CI/CD 🛡️

*Builds trust.*

- [x] **GitHub Actions**: Set up a CI pipeline to run tests on every push.
- [x] **Go Report Card**: Add the badge to README (aim for A+).
- [x] **Benchmarks**: Prove it's fast. Compare against other tools or raw FFmpeg.
- [x] **Linting Enforcement**: Add `golangci-lint` to CI to catch common issues (like unkeyed fields).

## 4. Documentation & Community 📚

- [x] **Examples Directory**: Real-world examples (e.g., "Simple CLI", "Web Server Worker").
- [x] **Godoc**: Ensure all exported functions have comments that show up nicely on pkg.go.dev.
- [x] **Contributing Guide**: `CONTRIBUTING.md` to help others help you.

## Recommended Next Step
**Implement Context Support**. It's the standard way to handle long-running processes in Go and is expected in any serious library.
