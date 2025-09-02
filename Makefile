# GoTsunami - Load Testing Tool
# Makefile for building, testing and releasing

.PHONY: help build test clean lint fmt vet coverage benchmark install run-examples docker release

# Variables
BINARY_NAME=gotsunami
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Default target
help: ## Show this help message
	@echo "GoTsunami - Load Testing Tool"
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the binary
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/gotsunami

build-linux: ## Build for Linux
	@echo "Building $(BINARY_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux ./cmd/gotsunami

build-windows: ## Build for Windows
	@echo "Building $(BINARY_NAME) for Windows..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows.exe ./cmd/gotsunami

build-darwin: ## Build for macOS
	@echo "Building $(BINARY_NAME) for macOS..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin ./cmd/gotsunami

build-all: build-linux build-windows build-darwin ## Build for all platforms

# Development targets
install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) ./cmd/gotsunami

# Code quality targets
fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@golangci-lint run

# Test targets
test: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage: test-coverage ## Alias for test-coverage

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Integration tests
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

# Example scenarios
run-examples: ## Run example scenarios
	@echo "Running example scenarios..."
	@./bin/$(BINARY_NAME) run examples/scenarios/basic_get.json --vus 5 --duration 10s --live

# Clean targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t gotsunami:$(VERSION) .
	@docker build -t gotsunami:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -it gotsunami:latest

# Release targets
release: clean build-all ## Create release builds
	@echo "Creating release $(VERSION)..."
	@mkdir -p releases
	@cp bin/$(BINARY_NAME)-linux releases/$(BINARY_NAME)-$(VERSION)-linux-amd64
	@cp bin/$(BINARY_NAME)-windows.exe releases/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe
	@cp bin/$(BINARY_NAME)-darwin releases/$(BINARY_NAME)-$(VERSION)-darwin-amd64
	@echo "Release files created in releases/"

# Development workflow
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@go mod download
	@go mod tidy
	@cp .env.example .env

# Quick development cycle
dev: fmt vet test build ## Quick development cycle (format, vet, test, build)

# CI/CD targets
ci: fmt vet lint test ## Run CI pipeline (format, vet, lint, test)

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all ./... > docs/API.md

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Version info
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"
