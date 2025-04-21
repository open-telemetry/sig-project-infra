# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, Test, and Lint Commands

```bash
# Build the Otto binary
make build

# Run all tests
make test 

# Run a single test
go test ./package/path -run TestName
# Example: go test ./internal -run TestIsSlashCommand

# Run linter
make lint

# Build Docker image
make docker-build
```

## Development Workflow

Otto follows a test-driven development (TDD) approach:

1. Write tests first that describe the expected behavior
2. Run the tests to verify they fail (Red)
3. Implement the minimum code necessary to pass the tests (Green)
4. Refactor the code while keeping tests passing (Refactor)
5. Repeat for each new feature or bug fix

## Code Style Guidelines

- **Imports**: Group standard library, then third-party, then local imports
- **Error Handling**: Return errors rather than panic; use error wrapping
- **Testing**: 
  - Write table-driven tests with clear input/output expectations
  - Include both happy path and error cases
  - Aim for high test coverage, especially for critical paths
- **Formatting**: Follow standard Go formatting with `gofmt`
- **Naming**: Use CamelCase for exported names, camelCase for unexported
- **Logging**: Use OpenTelemetry logging interfaces
- **Documentation**: Document all exported functions, types, and methods