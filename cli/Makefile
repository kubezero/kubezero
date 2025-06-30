# KubeZero CLI Makefile

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build flags
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Binary name
BINARY_NAME = kubezero

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) $(VERSION)"
	go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Built binaries:"
	@ls -la dist/

# Test the application
.PHONY: test
test:
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linting
.PHONY: lint
lint:
	go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed, skipping..."; \
		echo "Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Tidy dependencies
.PHONY: tidy
tidy:
	go mod tidy

# Install the binary to $GOPATH/bin or $HOME/go/bin
.PHONY: install
install:
	go install -ldflags="$(LDFLAGS)" .

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	rm -f kubezero-*
	rm -rf dist/
	rm -f coverage.out coverage.html

# Development commands
.PHONY: dev
dev: fmt lint test build

# Quick test with local mode
.PHONY: test-local
test-local: build
	./$(BINARY_NAME) bootstrap --local --cloud aws --region eu-west-1

# Show version information
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

# Help
.PHONY: help
help:
	@echo "KubeZero CLI Build Commands:"
	@echo ""
	@echo "  make build          - Build the binary for current platform"
	@echo "  make build-all      - Build binaries for all platforms"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linting tools"
	@echo "  make fmt            - Format code"
	@echo "  make install        - Install binary to GOPATH/bin"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make dev            - Run full development cycle (fmt, lint, test, build)"
	@echo "  make test-local     - Build and test local bootstrap command"
	@echo "  make version        - Show version information"
	@echo "  make help           - Show this help message"
