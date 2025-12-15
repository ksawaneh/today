# Release Guide

This document describes how to create a new release of the `today` application.

## Prerequisites

Before creating a release, ensure:

1. All tests pass: `go test ./...`
2. Code is properly formatted: `go fmt ./...`
3. CHANGELOG is updated (or prepare release notes)
4. Version number follows [semantic versioning](https://semver.org/)

## Release Process

The release process is fully automated using GitHub Actions and GoReleaser. Here's how to create a new release:

### 1. Prepare the Release

Update any necessary documentation:

```bash
# Update CHANGELOG.md or prepare release notes
# Review README.md for any updates needed
# Ensure all PRs are merged
```

### 2. Create and Push a Version Tag

```bash
# Create an annotated tag (replace X.Y.Z with your version)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to GitHub
git push origin v1.0.0
```

### 3. Automated Build Process

Once the tag is pushed, GitHub Actions will automatically:

1. Run all tests
2. Build binaries for:
   - Linux (amd64, arm64)
   - macOS (amd64/Intel, arm64/Apple Silicon)
   - Windows (amd64)
3. Create release archives (.tar.gz for Unix, .zip for Windows)
4. Generate checksums
5. Create a GitHub release draft
6. Upload all artifacts

### 4. Review and Publish

1. Go to [GitHub Releases](https://github.com/yourusername/today/releases)
2. Find the draft release
3. Review the generated changelog
4. Edit the release notes if needed
5. Click "Publish release"

## Version Numbering

Follow semantic versioning:

- **Major version** (X.0.0): Breaking changes
- **Minor version** (0.X.0): New features, backward compatible
- **Patch version** (0.0.X): Bug fixes, backward compatible

Examples:
- `v1.0.0` - First stable release
- `v1.1.0` - Added new feature (e.g., new pane type)
- `v1.1.1` - Bug fix
- `v2.0.0` - Breaking change (e.g., changed config format)

## Pre-releases

For alpha, beta, or release candidate versions:

```bash
# Alpha release
git tag -a v1.1.0-alpha.1 -m "Release v1.1.0-alpha.1"

# Beta release
git tag -a v1.1.0-beta.1 -m "Release v1.1.0-beta.1"

# Release candidate
git tag -a v1.1.0-rc.1 -m "Release v1.1.0-rc.1"

# Push the tag
git push origin v1.1.0-alpha.1
```

GoReleaser will automatically mark these as pre-releases on GitHub.

## Homebrew Tap (Optional)

If you've set up a Homebrew tap:

1. Create a separate repository: `homebrew-tap`
2. Add `HOMEBREW_TAP_GITHUB_TOKEN` secret to GitHub Actions
3. Uncomment the `brews` section in `.goreleaser.yaml`
4. The Homebrew formula will be automatically updated on release

## Manual Release (Fallback)

If GitHub Actions is unavailable, you can build locally with GoReleaser:

```bash
# Install GoReleaser
brew install goreleaser

# Create a release (requires GITHUB_TOKEN)
export GITHUB_TOKEN="your_github_token"
goreleaser release --clean

# Or build without releasing
goreleaser build --snapshot --clean
```

## Troubleshooting

### Build Fails

1. Check the GitHub Actions logs
2. Verify all tests pass locally: `go test ./...`
3. Ensure the tag follows the `v*` pattern
4. Check for syntax errors in `.goreleaser.yaml`

### Missing Artifacts

1. Check the GoReleaser configuration
2. Ensure all required files exist (README.md, LICENSE, docs/today.1)
3. Review the GitHub Actions workflow logs

### Homebrew Tap Issues

1. Verify `HOMEBREW_TAP_GITHUB_TOKEN` is set
2. Ensure the tap repository exists
3. Check the tap repository has write permissions

## Post-Release

After publishing a release:

1. Update the main branch if needed
2. Close related GitHub issues/milestones
3. Announce the release (Twitter, blog, etc.)
4. Monitor for bug reports

## Rollback

To rollback a release:

1. Delete the GitHub release
2. Delete the tag locally and remotely:

```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin :refs/tags/v1.0.0
```

3. Create a new patch release with fixes if needed

## Resources

- [GoReleaser Documentation](https://goreleaser.com)
- [Semantic Versioning](https://semver.org/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
