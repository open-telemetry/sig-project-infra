# Makefile for Otto
BINARY := otto
CMD_DIR := ./cmd/otto
VERSION ?= $(shell git describe --tags --always --dirty --match "v*" 2> /dev/null || echo "dev")

.PHONY: all build clean run test lint format import-check license-check sec-check release snapshot

all: build

build:
	go build -o $(BINARY) $(CMD_DIR)

clean:
	rm -f $(BINARY)

run: build
	./$(BINARY)

test:
	go test ./...

lint:
	golangci-lint run

format:
	gofmt -s -w .
	go mod tidy

import-check:
	gci write -s standard -s default -s "prefix(github.com/open-telemetry)" .

license-check:
	go run github.com/google/addlicense -c "The OpenTelemetry Authors" -l apache -check .

license-add:
	go run github.com/google/addlicense -c "The OpenTelemetry Authors" -l apache .

sec-check:
	go run github.com/google/osv-scanner:v1 .

# GoReleaser commands
release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean

docker-build:
	docker build -t ghcr.io/open-telemetry/sig-project-infra/otto:latest .

goreleaser-check:
	goreleaser check