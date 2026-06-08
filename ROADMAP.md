# Mosaic Roadmap

This roadmap tracks planned work at a high level. It is not a release promise.

## Done

- [x] HLS CMAF pipeline.
- [x] DASH CMAF pipeline.
- [x] Functional options for threads, GPU backend, logging, and orientation normalization.
- [x] Progress callback support from FFmpeg `-progress`.
- [x] Hardware encoder selection for NVENC, VAAPI, and VideoToolbox.
- [x] Orientation-aware probing and ladder selection.
- [x] Optional physical orientation normalization.
- [x] Output rotation metadata clearing.
- [x] Aspect-preserving output ladders for square, portrait, landscape, and non-standard dimensions.
- [x] Automatic HLS/DASH output directory creation.
- [x] Executor abstraction with mock-driven tests.
- [x] Human and AI documentation set.

## Next

- [ ] Improve progress model with computed percentage.
- [ ] Expand real-media integration fixtures under `testdata/`.
- [ ] Add modern codec options for HEVC and AV1.
- [ ] Improve hardware acceleration paths, especially VAAPI upload/filter requirements.
- [ ] Add thumbnail and sprite generation helpers.
- [ ] Add richer rendition configuration hooks.
- [ ] Add cloud output hooks for S3/GCS upload workflows.
- [ ] Add DRM integration surfaces for Widevine and FairPlay packaging workflows.

## Maintenance

- [ ] Keep Markdown docs synchronized with code changes in every PR.
- [ ] Keep smoke-test examples current for square, portrait, rotated portrait, and remote URL inputs.
- [ ] Keep FFmpeg command behavior covered by mock tests.
- [ ] Avoid adding fixture media files unless they are small and intentionally licensed for repository use.

## Documentation

Primary docs:

- `README.md`
- `docs/API.md`
- `docs/ARCHITECTURE.md`
- `docs/ENCODING.md`
- `docs/TESTING.md`
- `docs/TROUBLESHOOTING.md`
- `SUPPORT.md`
- `SECURITY.md`
- `AGENTS.md`
