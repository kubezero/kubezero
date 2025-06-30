# KubeZero CLI Makefile

.PHONY: build clean test deps run-bootstrap help

# Default target
help:
	@echo "KubeZero CLI - Available targets:"
	@echo "  build         - Build the kubezero binary"
	@echo "  clean         - Remove built binaries"
	@echo "  test          - Run tests"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  run-bootstrap - Run bootstrap command directly"
	@echo "  run-dev       - Run bootstrap with default config"
	@echo "  status        - Check cluster status"
	@echo "  cleanup       - Force cleanup cluster"
	@echo "  install       - Install binary to /usr/local/bin"
	@echo "  help          - Show this help message"

# Build binary
build: deps
	@echo "Building kubezero CLI..."
	go build -o kubezero .
	@echo "✅ Build complete! Run './kubezero --help' to get started."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f kubezero
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run bootstrap command directly (for development)
run-bootstrap:
	@echo "Running bootstrap command..."
	go run . bootstrap

# Install to system
install: build
	@echo "Installing kubezero to /usr/local/bin..."
	sudo cp kubezero /usr/local/bin/
	@echo "✅ kubezero installed! Run 'kubezero --help' from anywhere."

# Development: run with custom config
run-dev:
	go run . bootstrap --config ../bootstrap/k3d-bootstrap-cluster.yaml

# Quick status check
status:
	./kubezero status

# Quick cleanup
cleanup:
	./kubezero cleanup --force
