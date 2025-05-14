# Makefile for Otto
BINARY := otto
CMD_DIR := ./cmd/otto
COVERAGE_DIR := ./coverage

.PHONY: all build clean run test lint lint-fix format fix-comments fmt fix-imports fix-lines fmt-all lint-all test-coverage test-coverage-html test-coverage-badge test-coverage-check migrate-up migrate-down docker-build install-tools

all: build

build:
	go build -o $(BINARY) $(CMD_DIR)

clean:
	rm -f $(BINARY)
	rm -rf $(COVERAGE_DIR)

run: build
	./$(BINARY)

test:
	go test ./...

# Run tests with coverage and output to console
test-coverage:
	mkdir -p $(COVERAGE_DIR)
	go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -func=$(COVERAGE_DIR)/coverage.out

# Run tests with coverage and generate HTML report
test-coverage-html: test-coverage
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated at $(COVERAGE_DIR)/coverage.html"

# Generate a coverage badge for use in documentation
test-coverage-badge: test-coverage
	@go run scripts/cmd/badge/main.go -coverage=$(COVERAGE_DIR)/coverage.out -output=$(COVERAGE_DIR)/coverage-badge.svg
	@echo "Coverage badge generated at $(COVERAGE_DIR)/coverage-badge.svg"

# Check if coverage meets threshold
test-coverage-check: test-coverage
	@go run scripts/cmd/checker/main.go -coverage=$(COVERAGE_DIR)/coverage.out -threshold=70
	@echo "Coverage check passed!"

# Run database migrations up
migrate-up:
	@echo "Running migrations up..."
	@go run $(CMD_DIR) migrate up

# Run database migrations down
migrate-down:
	@echo "Running migrations down..."
	@go run $(CMD_DIR) migrate down

# Run linting checks
lint:
	golangci-lint run

# Fix linting issues automatically when possible
lint-fix:
	golangci-lint run --fix

# Format code with gofumpt (stricter than gofmt)
format:
	gofumpt -l -w .

# Fix missing periods in Go comments with godot
fix-comments:
	godot -w .

# Fix imports ordering and grouping
fix-imports:
	goimports -local go.opentelemetry.io -w .

# Fix line length with golines
fix-lines:
	golines -w --max-len=120 .

# Run all code formatters
fmt: format fix-comments

# Run all code formatters and fixers
fmt-all: fix-lines fix-imports format fix-comments

# Run all linting and formatting checks and auto-fix where possible
lint-all: lint-fix fmt-all

# Install required formatting and linting tools
install-tools:
	go install mvdan.cc/gofumpt@latest
	go install github.com/segmentio/golines@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/tetafro/godot/cmd/godot@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

docker-build:
	docker build -t otel-otto:latest .