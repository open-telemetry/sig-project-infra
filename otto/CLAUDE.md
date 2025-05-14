# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build: `make build` or `go build -o otto ./cmd/otto`
- Run: `make run` or `./otto`
- Test all: `make test` or `go test ./...`
- Test single file: `go test ./path/to/package -run TestName`
- Generate test coverage: `make test-coverage`
- Generate HTML coverage report: `make test-coverage-html`
- Check coverage threshold: `make test-coverage-check`
- Lint: `make lint` or `golangci-lint run`
- Auto-fix linting issues: `make lint-fix`
- Format code: `make fmt` (basic) or `make fmt-all` (comprehensive)
- Fix imports ordering: `make fix-imports`
- Fix line length: `make fix-lines`
- Fix comments: `make fix-comments`
- Run all linting and formatting: `make lint-all`
- Install required tools: `make install-tools`
- Docker build: `make docker-build` or `docker build -t otel-otto:latest .`
- Database migrations: `make migrate-up` or `make migrate-down`

## Architecture
- The codebase follows a clean architecture pattern with dependency injection
- Database operations use the Repository pattern with Go generics for type safety
- All major components define interfaces for testability and use constructor-based dependency injection
- Database migrations are handled via the golang-migrate library
- The code is organized into focused packages with clean separation of concerns

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
