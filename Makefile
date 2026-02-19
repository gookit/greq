# Makefile for greq project
# 
# Usage:
#   make build          - Build all binaries for current platform
#   make build-all      - Build binaries for all platforms
#   make build-linux    - Build binaries for Linux
#   make build-darwin   - Build binaries for macOS
#   make build-windows  - Build binaries for Windows
#   make test           - Run all tests
#   make clean          - Clean build artifacts
#   make install        - Install binaries to GOPATH/bin

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Binary names
BINS := greq gbench

# Build directory
BUILD_DIR := bin

# Version info (can be overridden by env vars or git tags)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Ldflags for version injection
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Go files (note: find command may have issues on Windows)
# GO_FILES := $(shell find . -name '*.go' -type f)

# Platform-specific variables
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Binary extension for Windows
ifeq ($(GOOS),windows)
    BINARY_EXT := .exe
else
    BINARY_EXT :=
endif

# Race detector support (requires CGO on Windows)
ifeq ($(GOOS),windows)
    TEST_FLAGS := -v -cover
else
    TEST_FLAGS := -v -race -cover
endif

# Default target
.PHONY: all
all: build

# ============================================================================
# Build targets
# ============================================================================

.PHONY: build
build: ## Build all binaries for current platform
	@echo "Building binaries for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	@for bin in $(BINS); do \
		echo "  Building $$bin..."; \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin$(BINARY_EXT) ./cmd/$$bin; \
	done
	@echo "Build complete. Binaries in $(BUILD_DIR)/"

.PHONY: build-greq
build-greq: ## Build greq binary only
	@echo "Building greq..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/greq$(BINARY_EXT) ./cmd/greq

.PHONY: build-gbench
build-gbench: ## Build gbench binary only
	@echo "Building gbench..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/gbench$(BINARY_EXT) ./cmd/gbench

# ============================================================================
# Cross-platform build targets
# ============================================================================

# Linux builds
.PHONY: build-linux
build-linux: ## Build binaries for Linux (amd64, arm64)
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@for arch in amd64 arm64; do \
		echo "  Building for linux/$$arch..."; \
		for bin in $(BINS); do \
			GOOS=linux GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin-linux-$$arch ./cmd/$$bin; \
		done; \
	done
	@echo "Linux build complete."

.PHONY: build-linux-amd64
build-linux-amd64: ## Build binaries for Linux amd64
	@mkdir -p $(BUILD_DIR)
	@for bin in $(BINS); do \
		echo "Building $$bin for linux/amd64..."; \
		GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin-linux-amd64 ./cmd/$$bin; \
	done

.PHONY: build-linux-arm64
build-linux-arm64: ## Build binaries for Linux arm64
	@mkdir -p $(BUILD_DIR)
	@for bin in $(BINS); do \
		echo "Building $$bin for linux/arm64..."; \
		GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin-linux-arm64 ./cmd/$$bin; \
	done

# macOS builds
.PHONY: build-darwin
build-darwin: ## Build binaries for macOS (amd64, arm64)
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@for arch in amd64 arm64; do \
		echo "  Building for darwin/$$arch..."; \
		for bin in $(BINS); do \
			GOOS=darwin GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin-darwin-$$arch ./cmd/$$bin; \
		done; \
	done
	@echo "macOS build complete."

.PHONY: build-darwin-amd64
build-darwin-amd64: ## Build binaries for macOS amd64 (Intel)
	@mkdir -p $(BUILD_DIR)
	@for bin in $(BINS); do \
		echo "Building $$bin for darwin/amd64..."; \
		GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin-darwin-amd64 ./cmd/$$bin; \
	done

.PHONY: build-darwin-arm64
build-darwin-arm64: ## Build binaries for macOS arm64 (Apple Silicon)
	@mkdir -p $(BUILD_DIR)
	@for bin in $(BINS); do \
		echo "Building $$bin for darwin/arm64..."; \
		GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin-darwin-arm64 ./cmd/$$bin; \
	done

# Windows builds
.PHONY: build-windows
build-windows: ## Build binaries for Windows (amd64, arm64)
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@for arch in amd64 arm64; do \
		echo "  Building for windows/$$arch..."; \
		for bin in $(BINS); do \
			GOOS=windows GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin-windows-$$arch.exe ./cmd/$$bin; \
		done; \
	done
	@echo "Windows build complete."

.PHONY: build-windows-amd64
build-windows-amd64: ## Build binaries for Windows amd64
	@mkdir -p $(BUILD_DIR)
	@for bin in $(BINS); do \
		echo "Building $$bin for windows/amd64..."; \
		GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin-windows-amd64.exe ./cmd/$$bin; \
	done

.PHONY: build-windows-arm64
build-windows-arm64: ## Build binaries for Windows arm64
	@mkdir -p $(BUILD_DIR)
	@for bin in $(BINS); do \
		echo "Building $$bin for windows/arm64..."; \
		GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$bin-windows-arm64.exe ./cmd/$$bin; \
	done

# Build all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows ## Build binaries for all platforms
	@echo "All platform builds complete."
	@ls -la $(BUILD_DIR)/

# ============================================================================
# Development targets
# ============================================================================

.PHONY: test
test: ## Run all tests
	$(GOTEST) $(TEST_FLAGS) ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	$(GOTEST) $(TEST_FLAGS) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: lint
lint: ## Run linters
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	$(GOCMD) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GOCMD) vet ./...

# ============================================================================
# Install targets
# ============================================================================

.PHONY: install
install: ## Install binaries to GOPATH/bin
	@echo "Installing binaries..."
	@for bin in $(BINS); do \
		echo "  Installing $$bin..."; \
		$(GOCMD) install ./cmd/$$bin; \
	done
	@echo "Installation complete."

.PHONY: install-greq
install-greq: ## Install greq to GOPATH/bin
	$(GOCMD) install ./cmd/greq

.PHONY: install-gbench
install-gbench: ## Install gbench to GOPATH/bin
	$(GOCMD) install ./cmd/gbench

# ============================================================================
# Clean targets
# ============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	$(GOCLEAN)
	@echo "Clean complete."

# ============================================================================
# Utility targets
# ============================================================================

.PHONY: deps
deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) verify

.PHONY: deps-update
deps-update: ## Update dependencies
	$(GOMOD) tidy
	$(GOGET) -u ./...

.PHONY: version
version: ## Show version info
	@echo "Version:    $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# Compressed release archives
# ============================================================================

.PHONY: release
release: build-all ## Create release archives for all platforms
	@echo "Creating release archives..."
	@mkdir -p release
	@cd $(BUILD_DIR) && \
	for bin in $(BINS); do \
		zip ../release/$$bin-$(VERSION)-linux-amd64.zip $$bin-linux-amd64; \
		zip ../release/$$bin-$(VERSION)-linux-arm64.zip $$bin-linux-arm64; \
		zip ../release/$$bin-$(VERSION)-darwin-amd64.zip $$bin-darwin-amd64; \
		zip ../release/$$bin-$(VERSION)-darwin-arm64.zip $$bin-darwin-arm64; \
		zip ../release/$$bin-$(VERSION)-windows-amd64.zip $$bin-windows-amd64.exe; \
		zip ../release/$$bin-$(VERSION)-windows-arm64.zip $$bin-windows-arm64.exe; \
	done
	@echo "Release archives created in release/"
