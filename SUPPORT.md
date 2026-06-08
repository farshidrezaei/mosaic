# Support

Use this document to decide where to ask questions or report problems.

## Before Asking

Check:

- [README.md](README.md)
- [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
- [docs/ENCODING.md](docs/ENCODING.md)
- [CHANGELOG.md](CHANGELOG.md)

## Bug Reports

Include:

- exact input file or URL, if shareable
- expected output
- actual output
- Go version
- FFmpeg version
- FFprobe version
- OS and architecture
- full error output
- source FFprobe JSON
- output FFprobe JSON, if output was generated

Source probe command:

```bash
ffprobe -v error -select_streams v:0 \
  -show_entries stream=width,height,avg_frame_rate:stream_tags=rotate:stream_side_data=rotation \
  -of json input.mp4
```

Output probe command:

```bash
ffprobe -v error -select_streams v:0 \
  -show_entries stream=width,height:stream_tags=rotate:stream_side_data=rotation \
  -of json /tmp/output/init_0.mp4
```

## Feature Requests

Describe:

- the use case
- why current options are insufficient
- desired API shape, if known
- FFmpeg behavior or command examples, if relevant

## Questions

For usage questions, include:

- a minimal code snippet
- input media metadata
- output format target: HLS or DASH
- whether orientation normalization is enabled
- whether hardware encoding is enabled

## Security Reports

See [SECURITY.md](SECURITY.md).
