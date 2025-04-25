# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Structure

This repository contains infrastructure projects for the OpenTelemetry SIG-Project-Infra.

The main project is Otto, a Go-based service located in the `./otto` directory.

## Commands for Otto

For Otto-specific commands, refer to `./otto/CLAUDE.md`.

```bash
# Run Otto tests
cd otto && make test

# Run a specific Otto test
cd otto && go test ./package/path -run TestName

# Build Otto
cd otto && make build

# Lint Otto code
cd otto && make lint
```

## General Guidelines

- Create code and configs following the standards in each project's directory
- Use descriptive commit messages that explain why changes were made
- Add appropriate tests for new functionality
- Maintain compatibility with GitHub Actions workflows