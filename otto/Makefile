# Makefile for Otto
BINARY := otto
CMD_DIR := ./cmd/otto

.PHONY: all build clean run test lint

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

docker-build:
	docker build -t otel-otto:latest .
