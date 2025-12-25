# Contributing to Mosaic

First off, thanks for taking the time to contribute! 🎉

The following is a set of guidelines for contributing to Mosaic. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## 🛠️ Development Setup

1. **Fork the repo** and clone it locally.
2. **Install Go** (1.20+).
3. **Install FFmpeg** (required for integration tests).
   ```bash
   # macOS
   brew install ffmpeg

   # Ubuntu/Debian
   sudo apt install ffmpeg
   ```

## 🧪 Running Tests

We strive for **100% test coverage**. Please ensure all tests pass before submitting a PR.

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run linter
golangci-lint run
```

## 📝 Code Style

- Follow standard Go conventions (use `gofmt`).
- **Use keyed fields for struct literals** (e.g., `User{Name: "Alice"}` instead of `User{"Alice"}`).
- Exported functions and types **must** have comments (for Godoc).
- Keep packages small and focused.
- **Use structured logging** (`log/slog`) instead of `fmt.Printf`.
- **Update the Changelog**: Add a brief description of your changes to the `[Unreleased]` section of `CHANGELOG.md`.

## 📝 Changelog Management

We follow the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format. Please ensure your PR includes an update to the `CHANGELOG.md` file under the appropriate category:
- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Deprecated` for soon-to-be removed features.
- `Removed` for now removed features.
- `Fixed` for any bug fixes.
- `Security` in case of vulnerabilities.

## 🚀 Submitting a Pull Request

1. Create a new branch: `git checkout -b my-feature-branch`
2. Make your changes and write tests.
3. Ensure `go test ./...` passes.
4. Push to your fork and submit a Pull Request.
5. Provide a clear description of what you changed and why.

## 🐛 Reporting Bugs

Open an issue with:
- A clear title.
- Steps to reproduce.
- Expected vs. actual behavior.
- Your environment (OS, Go version, FFmpeg version).

Thanks for helping make Mosaic better! ❤️
