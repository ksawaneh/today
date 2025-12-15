# Distribution Setup

This document provides an overview of the distribution infrastructure for the `today` application.

## Overview

The `today` app uses a modern Go distribution stack:

- **GoReleaser** - Cross-platform binary builds
- **GitHub Actions** - CI/CD automation
- **GitHub Releases** - Binary distribution
- **Homebrew** (optional) - macOS/Linux package manager

## Files and Structure

### Build Configuration

**`.goreleaser.yaml`**
- Defines build targets (Linux, macOS, Windows)
- Configures archive formats
- Sets up checksums and changelog generation
- Includes man page in releases

### CI/CD Workflows

**`.github/workflows/release.yml`**
- Triggered on version tags (`v*`)
- Runs tests before building
- Executes GoReleaser
- Creates GitHub release with artifacts

**`.github/workflows/ci.yml`**
- Runs on PRs and commits to main
- Tests on Linux, macOS, Windows
- Runs linters
- Generates coverage reports

### Documentation

**`docs/today.1`**
- Man page in troff format
- Full command reference
- Keybindings documentation
- Examples and tips

**`docs/RELEASE_GUIDE.md`**
- Step-by-step release process
- Version numbering guidelines
- Troubleshooting tips

**`LICENSE`**
- MIT License

## Supported Platforms

### Binary Releases

The following platforms are built automatically:

| OS      | Architecture | Format  | Notes                |
|---------|-------------|---------|----------------------|
| Linux   | amd64       | tar.gz  | Most Linux distros   |
| Linux   | arm64       | tar.gz  | ARM-based servers    |
| macOS   | amd64       | tar.gz  | Intel Macs           |
| macOS   | arm64       | tar.gz  | Apple Silicon Macs   |
| Windows | amd64       | zip     | Windows 10/11        |

All binaries are statically linked (CGO_ENABLED=0) for maximum portability.

## Installation Methods

### 1. Homebrew (Recommended for macOS/Linux)

Once set up:

```bash
brew tap yourusername/tap
brew install today
```

To enable Homebrew releases:
1. Create a `homebrew-tap` repository
2. Add `HOMEBREW_TAP_GITHUB_TOKEN` to GitHub secrets
3. Uncomment the `brews` section in `.goreleaser.yaml`

### 2. Binary Download

Users can download pre-built binaries from [GitHub Releases](https://github.com/yourusername/today/releases).

**Linux/macOS:**
```bash
curl -LO https://github.com/yourusername/today/releases/latest/download/today_linux_amd64.tar.gz
tar -xzf today_linux_amd64.tar.gz
sudo mv today /usr/local/bin/
```

**Windows:**
Download the `.zip` file and extract it to a directory in your PATH.

### 3. Go Install

For users with Go installed:

```bash
go install github.com/yourusername/today/cmd/today@latest
```

### 4. Build from Source

```bash
git clone https://github.com/yourusername/today.git
cd today
go build -o today ./cmd/today
```

## Version Information

Version information is embedded at build time via ldflags:

```bash
go build -ldflags="-X main.version=1.0.0 -X main.commit=abc123 -X main.date=2025-01-15"
```

GoReleaser automatically sets these values during release builds.

Users can check the version:

```bash
today --version
```

Output:
```
today version 1.0.0
  commit: abc123
  built:  2025-01-15T10:30:00Z
```

## Release Artifacts

Each release includes:

1. **Binaries** - For all supported platforms
2. **Archives** - tar.gz (Unix) or zip (Windows)
3. **Checksums** - SHA256 sums in `checksums.txt`
4. **Changelog** - Auto-generated from commit messages
5. **Documentation** - README, LICENSE, and man page

### Artifact Naming

Archives follow the pattern:
```
today_{version}_{os}_{arch}.{ext}
```

Examples:
- `today_1.0.0_linux_amd64.tar.gz`
- `today_1.0.0_darwin_arm64.tar.gz`
- `today_1.0.0_windows_amd64.zip`

## Man Page

The man page (`docs/today.1`) is:
- Written in troff format
- Included in all release archives
- Installed by Homebrew formula

Users can view it with:
```bash
man today
```

Or directly from the file:
```bash
man docs/today.1
```

## Security

### Checksums

All releases include SHA256 checksums for verification:

```bash
# Download binary and checksums
curl -LO https://github.com/yourusername/today/releases/latest/download/today_linux_amd64.tar.gz
curl -LO https://github.com/yourusername/today/releases/latest/download/checksums.txt

# Verify
sha256sum -c checksums.txt --ignore-missing
```

### Static Linking

All binaries are statically linked (`CGO_ENABLED=0`), which:
- Eliminates runtime dependencies
- Improves portability
- Reduces attack surface
- Simplifies deployment

## Customization

### Adding New Platforms

Edit `.goreleaser.yaml`:

```yaml
builds:
  - goos:
      - linux
      - darwin
      - windows
      - freebsd  # Add this
    goarch:
      - amd64
      - arm64
      - 386      # Add this
```

### Custom Archives

Add additional files to archives:

```yaml
archives:
  - files:
      - README.md
      - LICENSE
      - docs/today.1
      - CHANGELOG.md  # Add this
```

### Docker Images

Uncomment the `dockers` section in `.goreleaser.yaml` to build Docker images.

### Snap Packages

Uncomment the `snapcrafts` section for Ubuntu Snap packages.

## Maintenance

### Update Dependencies

```bash
go get -u ./...
go mod tidy
```

### Test Builds Locally

```bash
# Install GoReleaser
brew install goreleaser

# Build without releasing
goreleaser build --snapshot --clean

# Check output
ls -lh dist/
```

### Update Documentation

When adding features:
1. Update `README.md`
2. Update `docs/today.1` (man page)
3. Update help text in `cmd/today/main.go`

## Monitoring

### GitHub Actions

View build status:
- [Actions tab](https://github.com/yourusername/today/actions)
- Badge in README (optional)

### Download Statistics

View release downloads:
- GitHub Insights > Traffic > Releases

## Troubleshooting

### Build Failures

1. Check GitHub Actions logs
2. Verify all tests pass: `go test ./...`
3. Ensure tag format is `v*` (e.g., `v1.0.0`)
4. Validate `.goreleaser.yaml`: `goreleaser check`

### Missing Files in Archives

Ensure files exist before tagging:
```bash
ls README.md LICENSE docs/today.1
```

### Homebrew Formula Issues

1. Verify tap repository exists
2. Check `HOMEBREW_TAP_GITHUB_TOKEN` secret
3. Review GoReleaser Homebrew docs

## Resources

- [GoReleaser Documentation](https://goreleaser.com/intro/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Semantic Versioning](https://semver.org/)
- [Man Page Format (troff)](https://man7.org/linux/man-pages/man7/groff_man.7.html)
