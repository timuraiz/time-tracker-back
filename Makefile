.PHONY: run build clean test deps hot-reload help

# Default target
all: deps build

# Run the server
run:
	@echo "Starting time tracker server..."
	go run main.go

# Build the application
build:
	@echo "Building time tracker..."
	go build -o bin/time-tracker .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod tidy
	go mod download

# Run with hot reload (requires air)
hot-reload:
	@echo "Starting server with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to normal run..."; \
		make run; \
	fi

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi

# Start development server
dev: dev-setup hot-reload

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  run         - Run the server"
	@echo "  build       - Build the application"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  deps        - Download dependencies"
	@echo "  dev         - Start development server with hot reload"
	@echo "  hot-reload  - Run with hot reload (requires air)"
	@echo "  dev-setup   - Setup development environment"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code (requires golangci-lint)"
	@echo "  help        - Show this help message"