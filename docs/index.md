---
layout: home

hero:
  name: "Mosaic"
  text: "Adaptive Bitrate Video Packaging for Go"
  tagline: "Standardized HLS fMP4 & DASH CMAF with Next-Gen AV1/HEVC, Storyboards, AES-128 Encryption, and Cloud S3 Upload."
  image:
    src: /logo.jpg
    alt: Mosaic Logo
  actions:
    - theme: brand
      text: 🚀 Get Started
      link: /quickstart
    - theme: alt
      text: 📖 API Reference
      link: /API
    - theme: alt
      text: ⭐ Star on GitHub
      link: https://github.com/farshidrezaei/mosaic

features:
  - icon: 🚀
    title: Next-Gen Codecs & CRF
    details: First-class AV1 (libsvtav1), HEVC (libx265), and H.264 encoding with Capped-CRF Content-Aware Bitrate optimization saving 30–50% bandwidth.
  - icon: 🎞️
    title: Storyboard Thumbnails
    details: Automatic timeline scrubber JPEG sprite sheets and WebVTT cue files (thumbnails.vtt) plus HLS I-frame trick-play playlists.
  - icon: 📺
    title: Web Preview DevTools
    details: Built-in dark-mode preview player (mosaic preview) with Hls.js & Dash.js, live quality switching, and telemetry.
  - icon: 🔒
    title: HLS AES-128 Encryption
    details: Automated cryptographic 16-byte key generation and secure segment envelope encryption with standard #EXT-X-KEY tagging.
  - icon: ☁️
    title: Direct S3 Cloud Upload
    details: Zero-dependency pure Go AWS SigV4 signer for parallel streaming asset syncing directly to S3, MinIO, and Cloudflare R2.
  - icon: 🎚️
    title: Subtitles & Loudnorm
    details: Automatic SRT-to-WebVTT conversion, master playlist injection, and broadcast-standard EBU R128 audio volume leveling.
  - icon: 📐
    title: Aspect Preservation
    details: Dynamic rendition ladders preserving 16:9, 1:1, 9:16 portrait, and non-standard ratios without letterboxing or black bars.
  - icon: 🔄
    title: Orientation Normalization
    details: Auto-probe rotation tags and physical frame transposition with output metadata clearing for seamless mobile playback.
---

<div class="vp-doc">

## ⚡ Interactive Bitrate & Ladder Simulator

<AbrCalculator />

---

## 📺 Live Stream Player Preview

<StreamPlayer />

</div>
