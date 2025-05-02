# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build: `go build -o otto ./cmd/otto`
- Run: `./otto`
- Test all: `go test ./...`
- Test single file: `go test ./path/to/package -run TestName`
- Lint: `golangci-lint run`
- Docker build: `docker build -t otel-otto:latest .`

## Release Process
- Otto uses automated semantic versioning through release-please-action
- Version is managed in `.release-please-manifest.json` at the repository root
- Binary and Docker builds are automated with GoReleaser
- Docker images are published to GitHub Container Registry (ghcr.io)

## Code Style Guidelines
- License: Include SPDX license header (`// SPDX-License-Identifier: Apache-2.0`) in all Go files
- Imports: Use standard Go import grouping (stdlib, then external, then internal)
- Error handling: Use structured error handling with AppError type from internal/error.go
- Testing: Follow test-driven development (TDD) - write tests first, then implement code to pass tests
- Testing: Write table-driven tests with clear test cases and failure messages
- Variables: Use descriptive variable names in camelCase (e.g., errType, appErr)
- Documentation: Add comments for exported functions, constants, and types
- Logging: Use slog package for structured logging with appropriate levels
- Dependencies: This is an OpenTelemetry project; follow OTel conventions
- File formatting: Always include a newline at the end of every file

## Commit Message Guidelines
Follow [Conventional Commits](https://www.conventionalcommits.org/) format for all commits:
- `fix:` for bug fixes (triggers patch version bump)
- `feat:` for new features (triggers minor version bump)
- `feat!:` or `fix!:` for breaking changes (triggers major version bump)
- `docs:`, `chore:`, `test:`, etc. for non-release changes
