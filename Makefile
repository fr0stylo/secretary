# Makefile for Secretary project

# Variables
BINARY_NAME=secretary
GITHUB_REPO=ghcr.io/fr0stylo/secretory
VERSION?=latest
GOFLAGS=-ldflags="-s -w"
GOOS?=linux
GOARCH?=amd64
CGO_ENABLED?=0

# Go commands
GO=go
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOTEST=$(GO) test
GOGET=$(GO) get
GOINSTALL=$(GO) install

# Paths
SRC_DIR=.
BIN_DIR=./bin

.PHONY: all build clean test lint install docker docker-push help

all: clean build test

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) $(GOFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(SRC_DIR)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BIN_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Install the application
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOINSTALL) $(GOFLAGS) .

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t $(GITHUB_REPO):$(VERSION) -f Containerfile .

# Push Docker image to registry
docker-push: docker
	@echo "Pushing Docker image to registry..."
	docker push $(GITHUB_REPO):$(VERSION)

# Show help
help:
	@echo "Available targets:"
	@echo "  all         - Clean, build, and test the application"
	@echo "  build       - Build the application"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter"
	@echo "  install     - Install the application"
	@echo "  docker      - Build Docker image"
	@echo "  docker-push - Build and push Docker image to registry"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION     - Docker image version tag (default: latest)"
	@echo "  GOOS        - Target operating system (default: linux)"
	@echo "  GOARCH      - Target architecture (default: amd64)"
	@echo "  CGO_ENABLED - Enable CGO (default: 0)"
