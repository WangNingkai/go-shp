# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Directories
BUILD_DIR=build
COVERAGE_DIR=coverage

.PHONY: all build clean test coverage benchmark lint fmt help

# Default target
all: fmt lint test build ## Run fmt, lint, test and build

# Build the library (create example binaries)
build: ## Build example binaries
	@echo "Building example binaries..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/example ./example

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)

# Run tests
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# Run benchmarks
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Run linter
lint: ## Run linter
	@echo "Running linter..."
	$(GOLINT) run

# Format code
fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	$(GOCMD) mod tidy

# Vet code
vet: ## Vet code
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Check for updates
check-updates: ## Check for module updates
	@echo "Checking for updates..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Generate documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@$(GOCMD) doc -all . > docs/API.md

# Install development dependencies
install-dev: ## Install development dependencies
	@echo "Installing development dependencies..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Pre-commit checks
pre-commit: fmt vet lint test ## Run pre-commit checks

# Release preparation
release-check: pre-commit coverage benchmark ## Run all checks for release

# Performance profiling
profile-cpu: ## Run CPU profiling
	@echo "Running CPU profiling..."
	$(GOTEST) -cpuprofile=cpu.prof -bench=. ./...
	$(GOCMD) tool pprof cpu.prof

profile-mem: ## Run memory profiling
	@echo "Running memory profiling..."
	$(GOTEST) -memprofile=mem.prof -bench=. ./...
	$(GOCMD) tool pprof mem.prof

# Generate test data
generate-testdata: ## Generate test data
	@echo "Generating test data..."
	$(GOBUILD) -o $(BUILD_DIR)/testdata-gen ./tools/testdata-gen
	./$(BUILD_DIR)/testdata-gen

# Show help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
