# Security Policy

## Supported Versions

Mosaic is currently pre-1.0. Security fixes target the default branch unless a release branch is explicitly maintained.

## Reporting a Vulnerability

Please do not open a public issue for sensitive security reports.

Send a private security report to:
- **Maintainer Email**: `farshid.net1@gmail.com`
- **GitHub Security Advisory**: [Report a vulnerability](https://github.com/farshidrezaei/mosaic/security/advisories/new)

Include in your report:

- a clear description of the vulnerability
- affected versions or commits, if known
- reproduction steps
- impact assessment
- any suggested mitigation

## Scope

Relevant security issues may include:

- command argument injection
- unsafe handling of untrusted paths or URLs
- unexpected file writes outside `OutputDir`
- unsafe temporary file handling
- denial-of-service behavior from crafted media metadata
- sensitive data leakage in errors or logs

Out of scope:

- vulnerabilities in FFmpeg, FFprobe, drivers, or operating system packages
- playback vulnerabilities in downstream media players
- issues requiring unsupported FFmpeg builds

## Operational Guidance

Mosaic invokes external FFmpeg and FFprobe binaries. Applications using Mosaic should:

- run encodes with least-privilege filesystem permissions
- isolate untrusted media processing where possible
- use writable output directories with controlled ownership
- validate or restrict remote input URLs when accepting user input
- keep FFmpeg and FFprobe updated
- impose application-level job timeouts and file-size limits

## Dependency Updates

The Go module currently has no third-party Go dependencies. Keep the Go toolchain and FFmpeg installation current in deployment environments.
