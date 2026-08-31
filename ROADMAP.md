# Mosaic Roadmap

This roadmap tracks completed capabilities and planned future work at a high level.

## Done

- [x] HLS CMAF (`fMP4`) and MPEG-TS packaging pipelines.
- [x] DASH CMAF (`manifest.mpd`, `init.m4s`, `chunk.m4s`) pipeline.
- [x] Next-Gen Codecs (**AV1** with `libsvtav1`, **HEVC / H.265** with `libx265`, **H.264** with `libx264`).
- [x] Capped-CRF Content-Aware Bitrate optimization (`WithCRF()`).
- [x] Storyboard Sprite Sheets and WebVTT cue generation (`WithThumbnails()`).
- [x] Local Web Preview DevTools HTTP Player (`mosaic preview`).
- [x] Subtitles Management (automatic SRT to WebVTT conversion + HLS/DASH manifest injection).
- [x] EBU R128 Broadcast Audio Volume Normalization (`WithNormalizeAudio()`).
- [x] Dynamic Logo / Watermark Overlays with custom positioning, scaling, and opacity (`WithWatermark()`).
- [x] HLS AES-128 Segment Encryption (`WithAES128Encryption()`).
- [x] Zero-Dependency Direct Cloud Upload to S3 / MinIO / Cloudflare R2 (`WithS3Upload()`).
- [x] I-Frames Only Trick-Play Playlists (`#EXT-X-I-FRAMES-ONLY`).
- [x] Real-time computed progress model (`Percentage`, `CurrentTime`, `Bitrate`, `Speed`).
- [x] Functional options for threads, GPU backend, logging, and orientation normalization.
- [x] Hardware encoder selection for NVENC, VAAPI, and VideoToolbox.
- [x] Orientation-aware probing and physical orientation normalization with metadata clearing.
- [x] Aspect-preserving output ladders for square, portrait, landscape, and non-standard dimensions.
- [x] Automatic HLS/DASH output directory creation.
- [x] Executor abstraction with 100% mock-driven test suite.

## Next

- [ ] Dual-codec simultaneous ladder generation (H.264 + AV1 in a single pass).
- [ ] VMAF / SSIM perceptual quality scoring pipeline.
- [ ] DRM packaging hooks for Widevine and FairPlay (CENC / CBCS).
- [ ] Webhook lifecycle callbacks (onSegmentComplete, onJobSuccess, onJobError).

## Maintenance

- [x] Keep Markdown docs synchronized with code changes in every PR.
- [x] Zero lint errors contract (`golangci-lint run`).
- [x] 100% test pass rate with Go race detector (`go test -v -race ./...`).

## Documentation

Primary docs:

- `README.md`
- `docs/API.md`
- `docs/options.md`
- `docs/quickstart.md`
- `docs/ARCHITECTURE.md`
- `docs/ENCODING.md`
- `docs/EXAMPLES.md`
- `docs/ROADMAP_PROGRESS.md`
- `docs/TESTING.md`
- `docs/TROUBLESHOOTING.md`
