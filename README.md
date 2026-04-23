# Pong Terminal

Terminal Pong built with Go, Bubble Tea, Lip Gloss, and Bubbles.

## Run locally

```bash
go run .
```

## Build locally

```bash
go build -o pong-terminal .
```

On Windows the output will be `pong-terminal.exe`.

## Cross-platform releases

This repo is configured for release builds on:

- Windows
- macOS
- Linux

for these architectures:

- `amd64`
- `arm64`

Release packaging is handled by [GoReleaser](https://goreleaser.com/).

### Local release test

Install GoReleaser and run:

```bash
goreleaser release --snapshot --clean
```

Artifacts will be generated in `dist/`.

### Publish a real release

1. Commit your changes.
2. Create a git tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

3. GitHub Actions will build and publish archives for Windows, macOS, and Linux automatically.

### Produced artifacts

- `pong-terminal_<version>_windows_amd64.zip`
- `pong-terminal_<version>_linux_amd64.tar.gz`
- `pong-terminal_<version>_linux_arm64.tar.gz`
- `pong-terminal_<version>_darwin_amd64.tar.gz`
- `pong-terminal_<version>_darwin_arm64.tar.gz`

Each release also includes `checksums.txt`.

## Notes

- The binary name is `pong-terminal`.
- If you later publish this as `go install ...`, update `go.mod` to your final repository module path.
