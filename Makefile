# Makefile for Fluid GitLab Probe

# Variables
BINARY_NAME=gitlab-probe
BUILD_DIR=build
CONFIG_DIR=config
STATE_DIR=state

GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

VERSION?=0.1.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

.PHONY: all build clean run test deps help

all: clean build

deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

build: deps
	@echo "Building GitLab probe..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "Probe built in $(BUILD_DIR)/$(BINARY_NAME)"

build-all: deps
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)

	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/main.go

	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/main.go
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 cmd/main.go

	# Windows
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/main.go

	@echo "Build completed for all platforms"

clean:
	@echo "Cleaning build files..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(STATE_DIR)

run: build
	@echo "Starting GitLab probe..."
	@cd $(BUILD_DIR) && ./$(BINARY_NAME)

run-config: build
	@echo "Starting GitLab probe with custom configuration..."
	@cd $(BUILD_DIR) && ./$(BINARY_NAME) -config ../$(CONFIG_DIR)/probe.yml

dev: deps
	@echo "Starting in development mode..."
	$(GO) run ./cmd

test: deps
	@echo "Running tests..."
	$(GO) test -v ./...

# Check code with golangci-lint
lint:
	@echo "Checking code with golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Check vulnerabilities
security:
	@echo "Checking vulnerabilities..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Installing..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# Create distribution package
package: build
	@echo "Creating distribution package..."
	@mkdir -p $(BUILD_DIR)/package
	@cp -r $(CONFIG_DIR) $(BUILD_DIR)/package/
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/package/
	@cp README.md $(BUILD_DIR)/package/
	@cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-$(VERSION).tar.gz package/
	@echo "Package created: $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION).tar.gz"

# Show help
help:
	@echo "Available commands:"
	@echo "  deps          - Install dependencies"
	@echo "  build         - Build the probe"
	@echo "  build-all     - Build for all platforms"
	@echo "  clean         - Clean build files"
	@echo "  run           - Run the probe"
	@echo "  run-config    - Run with custom configuration"
	@echo "  dev           - Development mode"
	@echo "  test          - Run tests"
	@echo "  test-structure- Test file structure and rotation"
	@echo "  test-coverage - Tests with coverage"
	@echo "  lint          - Check code"
	@echo "  fmt           - Format code"
	@echo "  security      - Check vulnerabilities"
	@echo "  package       - Create distribution package"
	@echo "  help          - Show this help"
