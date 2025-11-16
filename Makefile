.PHONY: help build run test test-verbose test-coverage clean docker-build docker-run fmt lint deps

# Default target
help:
	@echo "Gaxy - Google Analytics Proxy"
	@echo ""
	@echo "Available targets:"
	@echo "  make build         - Build the binary"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run tests"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make bench         - Run benchmarks"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linters"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make deps          - Download dependencies"

# Build the application
build:
	@echo "Building gaxy..."
	@go build -o bin/gaxy -ldflags="-s -w" .

# Run the application
run:
	@echo "Running gaxy..."
	@go run *.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linters
lint:
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Installing..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	@golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f bin/gaxy
	@rm -f coverage.out coverage.html
	@go clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t gaxy:latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -it --rm -p 3000:3000 gaxy:latest

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
