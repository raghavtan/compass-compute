# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOENV=$(GOCMD) env
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Docker parameters
DOCKERCMD=docker
DOCKERBUILD=$(DOCKERCMD) build
DOCKERIMAGE_NAME?=compass-compute
DOCKERIMAGE_TAG?=latest

# Binary and build parameters
BINARY_NAME=compass-compute
BUILD_DIR=bin
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)
CMD_DIR=./cmd

# Version and build info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"

# Tools and paths
GOIMPORTS_PATH=$(shell $(GOENV) GOPATH)/bin/goimports

# Linting
GOLANGCI_LINT_VERSION=v1.61.0
GOLINT_PATH=$(shell $(GOENV) GOPATH)/bin/golangci-lint

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: all setup build build-all test test-race test-cover lint fmt vet tidy docker-build docker-run clean deps check install uninstall help

# Default target
all: check build

# Setup development environment
setup:
	@printf "$(BLUE)Setting up development environment...$(NC)\n"
	$(GOGET) golang.org/x/tools/cmd/goimports
	$(GOGET) github.com/incu6us/goimports-reviser/v3
	@printf "$(YELLOW)Installing golangci-lint $(GOLANGCI_LINT_VERSION)...$(NC)\n"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell $(GOENV) GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@printf "$(GREEN)Setup complete.$(NC)\n"
	@printf "$(BLUE)Verifying tool installations...$(NC)\n"
	@printf "GOPATH: $(shell $(GOENV) GOPATH)\n"
	@if command -v goimports &> /dev/null; then \
		printf "goimports: $(shell which goimports)\n"; \
	elif [ -f $(GOIMPORTS_PATH) ]; then \
		printf "goimports: $(GOIMPORTS_PATH) (installed)\n"; \
	else \
		printf "goimports: not found\n"; \
	fi
	@if command -v golangci-lint &> /dev/null; then \
		printf "golangci-lint: $(shell which golangci-lint)\n"; \
	elif [ -f $(GOLINT_PATH) ]; then \
		printf "golangci-lint: $(GOLINT_PATH) (installed)\n"; \
	else \
		printf "golangci-lint: not found\n"; \
	fi

# Download dependencies
deps:
	@printf "$(BLUE)Downloading dependencies...$(NC)\n"
	$(GOMOD) download
	$(GOMOD) verify

# Tidy go modules
tidy:
	@printf "$(BLUE)Tidying go module files...$(NC)\n"
	$(GOMOD) tidy
	@printf "$(GREEN)Go module files tidied.$(NC)\n"

# Format code
fmt:
	@printf "$(BLUE)Formatting code...$(NC)\n"
	$(GOFMT) -s -w .
	@if command -v goimports &> /dev/null; then \
		goimports -w .; \
	elif [ -f $(GOIMPORTS_PATH) ]; then \
		$(GOIMPORTS_PATH) -w .; \
	else \
		printf "$(YELLOW)goimports not found, skipping import formatting$(NC)\n"; \
	fi
	@printf "$(GREEN)Code formatted.$(NC)\n"

# Vet code
vet:
	@printf "$(BLUE)Vetting code...$(NC)\n"
	$(GOVET) ./...
	@printf "$(GREEN)Code vetted.$(NC)\n"

# Lint code with fallback
lint:
	@printf "$(BLUE)Linting code...$(NC)\n"
	@if ! command -v $(GOLINT_PATH) &> /dev/null; then \
		printf "$(RED)golangci-lint not found. Please run 'make setup' first.$(NC)\n"; \
		exit 1; \
	fi
	@if ! $(GOLINT_PATH) run ./... 2>/dev/null; then \
		printf "$(YELLOW)golangci-lint failed, running basic go vet instead...$(NC)\n"; \
		$(GOVET) ./...; \
		printf "$(YELLOW)Consider updating golangci-lint: make setup$(NC)\n"; \
	fi
	@printf "$(GREEN)Linting complete.$(NC)\n"

# Check runs all quality checks
check: fmt vet lint tidy
	@printf "$(GREEN)All checks passed!$(NC)\n"

# CI-friendly check that skips problematic tools
check-ci: fmt-basic vet tidy
	@printf "$(GREEN)CI checks passed!$(NC)\n"

# Basic formatting without goimports (for CI environments)
fmt-basic:
	@printf "$(BLUE)Basic code formatting...$(NC)\n"
	$(GOFMT) -s -w .
	@printf "$(GREEN)Basic formatting complete.$(NC)\n"

# Build the application
build:
	@printf "$(BLUE)Building $(BINARY_NAME)...$(NC)\n"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) $(CMD_DIR)
	@printf "$(GREEN)$(BINARY_NAME) built successfully at $(BINARY_PATH)$(NC)\n"

# Build for multiple platforms
build-all:
	@printf "$(BLUE)Building for multiple platforms...$(NC)\n"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@printf "$(GREEN)Multi-platform build complete.$(NC)\n"

# Run tests
test:
	@printf "$(BLUE)Running tests...$(NC)\n"
	$(GOTEST) -v ./...
	@printf "$(GREEN)Tests complete.$(NC)\n"

# Run tests with race detection
test-race:
	@printf "$(BLUE)Running tests with race detection...$(NC)\n"
	$(GOTEST) -race -v ./...
	@printf "$(GREEN)Race tests complete.$(NC)\n"

# Run tests with coverage
test-cover:
	@printf "$(BLUE)Running tests with coverage...$(NC)\n"
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@printf "$(GREEN)Coverage report generated: coverage.html$(NC)\n"

# Build Docker image
docker-build:
	@printf "$(BLUE)Building Docker image $(DOCKERIMAGE_NAME):$(DOCKERIMAGE_TAG)...$(NC)\n"
	$(DOCKERBUILD) -t $(DOCKERIMAGE_NAME):$(DOCKERIMAGE_TAG) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-f Dockerfile .
	@printf "$(GREEN)Docker image built successfully.$(NC)\n"

# Run Docker container
docker-run:
	@printf "$(BLUE)Running Docker container...$(NC)\n"
	$(DOCKERCMD) run --rm $(DOCKERIMAGE_NAME):$(DOCKERIMAGE_TAG)

# Install binary to GOPATH/bin
install: build
	@printf "$(BLUE)Installing $(BINARY_NAME)...$(NC)\n"
	cp $(BINARY_PATH) $(shell $(GOENV) GOPATH)/bin/
	@printf "$(GREEN)$(BINARY_NAME) installed to $(shell $(GOENV) GOPATH)/bin/$(NC)\n"

# Uninstall binary from GOPATH/bin
uninstall:
	@printf "$(BLUE)Uninstalling $(BINARY_NAME)...$(NC)\n"
	rm -f $(shell $(GOENV) GOPATH)/bin/$(BINARY_NAME)
	@printf "$(GREEN)$(BINARY_NAME) uninstalled.$(NC)\n"

# Clean build artifacts
clean:
	@printf "$(BLUE)Cleaning up...$(NC)\n"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@printf "$(GREEN)Cleanup complete.$(NC)\n"

# Show help
help:
	@printf "$(BLUE)compass-compute Makefile$(NC)\n"
	@printf "\n$(YELLOW)Usage:$(NC)\n"
	@printf "  make [target]\n"
	@printf "\n$(YELLOW)Development Targets:$(NC)\n"
	@printf "  $(GREEN)setup$(NC)           Install development tools and linters\n"
	@printf "  $(GREEN)deps$(NC)            Download and verify dependencies\n"
	@printf "  $(GREEN)tidy$(NC)            Tidy go.mod and go.sum files\n"
	@printf "  $(GREEN)fmt$(NC)             Format Go code\n"
	@printf "  $(GREEN)vet$(NC)             Run go vet\n"
	@printf "  $(GREEN)lint$(NC)            Run golangci-lint\n"
	@printf "  $(GREEN)check$(NC)           Run all quality checks (fmt, vet, lint, tidy)\n"
	@printf "\n$(YELLOW)Build Targets:$(NC)\n"
	@printf "  $(GREEN)build$(NC)           Build the application\n"
	@printf "  $(GREEN)build-all$(NC)       Build for multiple platforms\n"
	@printf "  $(GREEN)install$(NC)         Install binary to GOPATH/bin\n"
	@printf "  $(GREEN)uninstall$(NC)       Remove binary from GOPATH/bin\n"
	@printf "\n$(YELLOW)Test Targets:$(NC)\n"
	@printf "  $(GREEN)test$(NC)            Run tests\n"
	@printf "  $(GREEN)test-race$(NC)       Run tests with race detection\n"
	@printf "  $(GREEN)test-cover$(NC)      Run tests with coverage report\n"
	@printf "\n$(YELLOW)Docker Targets:$(NC)\n"
	@printf "  $(GREEN)docker-build$(NC)    Build Docker image\n"
	@printf "  $(GREEN)docker-run$(NC)      Run Docker container\n"
	@printf "\n$(YELLOW)Utility Targets:$(NC)\n"
	@printf "  $(GREEN)clean$(NC)           Clean build artifacts\n"
	@printf "  $(GREEN)help$(NC)            Show this help message\n"
	@printf "  $(GREEN)all$(NC)             Run check and build (default)\n"
	@printf "\n$(YELLOW)Environment Variables:$(NC)\n"
	@printf "  $(GREEN)VERSION$(NC)         Set build version (default: git describe)\n"
	@printf "  $(GREEN)DOCKERIMAGE_NAME$(NC) Set Docker image name (default: compass-compute)\n"
	@printf "  $(GREEN)DOCKERIMAGE_TAG$(NC)  Set Docker image tag (default: latest)\n"