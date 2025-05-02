# Otto Release Process

This document outlines the release process for Otto.

## PR-Based Release Process

Otto uses a PR-based release process. New releases are created automatically when a PR with the `release` label is merged into the main branch.

### Release Labels

When you want to trigger a release, add one of the following labels to your PR:

- `release` - Creates a patch release (e.g., v1.0.0 → v1.0.1)
- `release:minor` - Creates a minor release (e.g., v1.0.0 → v1.1.0)
- `release:major` - Creates a major release (e.g., v1.0.0 → v2.0.0)

### Release Workflow

When a labeled PR is merged:

1. A GitHub Action will automatically:
   - Determine the next version number based on the latest tag
   - Create and push a new tag
   - Build binaries and Docker images using GoReleaser
   - Push Docker images to GitHub Container Registry (ghcr.io)
   - Create a GitHub release with release notes

### Available Artifacts

Each release produces:

- Binary releases for multiple platforms (Linux, macOS, AMD64, ARM64)
- Docker images published to `ghcr.io/open-telemetry/sig-project-infra/otto`
- Multi-architecture Docker manifests for convenient image pulling

### Docker Images

Docker images can be pulled without specifying architecture:

```sh
# Pull by version
docker pull ghcr.io/open-telemetry/sig-project-infra/otto:v1.0.0

# Pull latest
docker pull ghcr.io/open-telemetry/sig-project-infra/otto:latest
```

### Verifying Signatures

Otto releases are signed using Cosign. You can verify the signatures using the public key.

To verify binary checksums:

```sh
# Download the public key
curl -O https://raw.githubusercontent.com/open-telemetry/sig-project-infra/main/otto/cosign.pub

# Verify checksums
cosign verify-blob --key cosign.pub --signature checksums.txt.sig checksums.txt
```

To verify Docker images:

```sh
# Verify a specific version
cosign verify --key cosign.pub ghcr.io/open-telemetry/sig-project-infra/otto:v1.0.0

# Verify latest
cosign verify --key cosign.pub ghcr.io/open-telemetry/sig-project-infra/otto:latest
```

### Future Improvements

Planned improvements to the release process:

- Further enhance security with SBOM generation and attestations
- Explore OpenTelemetry Collector's GoReleaser Pro features for improved build times
- Implement automatic changelog generation based on PR content
- Add user-friendly release verification and installation instructions