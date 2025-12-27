.PHONY: all build install test clean uninstall help fmt lint vet check

# Binary name
BINARY_NAME=booklife

# Build directory
BUILD_DIR=build

# Platform-specific install directories
ifeq ($(OS),Windows_NT)
	# Windows
	INSTALL_DIR=$(USERPROFILE)\AppData\Local\Programs
	BIN_PATH=$(INSTALL_DIR)\$(BINARY_NAME).exe
else
	# Unix-like (Linux, macOS, BSD)
	INSTALL_DIR=$(HOME)/.local/bin
	BIN_PATH=$(INSTALL_DIR)/$(BINARY_NAME)
	# Go bin directory
	GOBIN?=$(shell go env GOPATH)/bin
	ifeq ($(GOBIN),/bin)
		GOBIN=$(HOME)/go/bin
	endif
endif

# Go build flags
GO=go
GOFLAGS=-trimpath
LDFLAGS=-s -w

# Version info (can be overridden)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S' 2>/dev/null || echo "unknown")
LDFLAGS+=-X 'github.com/user/booklife-mcp/cmd/booklife.Version=$(VERSION)' -X 'github.com/user/booklife-mcp/cmd/booklife.BuildTime=$(BUILD_TIME)'

# Default target
all: build

## build: Build the binary for the current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build binaries for all supported platforms
build-all:
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)
	@echo "Binaries built in $(BUILD_DIR)/"

## install: Install the binary to ~/.local/bin (or platform equivalent)
install: build
	@echo "Installing $(BINARY_NAME) to $(BIN_PATH)..."
	@mkdir -p $(INSTALL_DIR)
ifdef COPY_ON_WRITE
	@echo "Installing with copy-on-write (reflink)..."
	@if command -v reflink >/dev/null 2>&1; then \
		reflink $(BUILD_DIR)/$(BINARY_NAME) $(BIN_PATH); \
	elif command -v cp >/dev/null 2>&1 && cp --help 2>&1 | grep -q reflink; then \
		cp --reflink=auto $(BUILD_DIR)/$(BINARY_NAME) $(BIN_PATH); \
	else \
		echo "Warning: reflink not available, falling back to regular copy"; \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(BIN_PATH); \
	fi
else
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(BIN_PATH)
endif
	@chmod +x $(BIN_PATH)
	@echo "Installed to: $(BIN_PATH)"

## install-go: Install the binary to $(GOPATH)/bin (Go binary folder)
install-go: build
	@echo "Installing $(BINARY_NAME) to $(GOBIN)..."
	@mkdir -p $(GOBIN)
ifdef COPY_ON_WRITE
	@echo "Installing with copy-on-write (reflink)..."
	@if command -v reflink >/dev/null 2>&1; then \
		reflink $(BUILD_DIR)/$(BINARY_NAME) $(GOBIN)/$(BINARY_NAME); \
	elif command -v cp >/dev/null 2>&1 && cp --help 2>&1 | grep -q reflink; then \
		cp --reflink=auto $(BUILD_DIR)/$(BINARY_NAME) $(GOBIN)/$(BINARY_NAME); \
	else \
		echo "Warning: reflink not available, falling back to regular copy"; \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(GOBIN)/$(BINARY_NAME); \
	fi
else
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOBIN)/$(BINARY_NAME)
endif
	@chmod +x $(GOBIN)/$(BINARY_NAME)
	@echo "Installed to: $(GOBIN)/$(BINARY_NAME)"

## uninstall: Remove the installed binary
uninstall:
	@echo "Removing $(BIN_PATH)..."
	@rm -f $(BIN_PATH)
	@echo "Uninstalled"

## uninstall-go: Remove the binary from Go bin folder
uninstall-go:
	@echo "Removing $(GOBIN)/$(BINARY_NAME)..."
	@rm -f $(GOBIN)/$(BINARY_NAME)
	@echo "Uninstalled from $(GOBIN)/"

## test: Run all tests
test:
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## test-short: Run tests without race detection (faster)
test-short:
	$(GO) test -short ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned"

## fmt: Format Go code
fmt:
	$(GO) fmt ./...

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/"; \
	fi

## vet: Run go vet
vet:
	$(GO) vet ./...

## check: Run fmt, vet, and tests
check: fmt vet test-short

## deps: Download and verify dependencies
deps:
	$(GO) mod download
	$(GO) mod verify

## tidy: Clean up go.mod and go.sum
tidy:
	$(GO) mod tidy

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"

## help: Show this help message
help:
	@echo "BookLife MCP Server - Makefile targets"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
