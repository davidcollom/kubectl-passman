# Makefile for kubectl-passman

.PHONY: build test clean fmt vet lint install

# Build variables
BINARY_NAME=kubectl-passman
CMD_DIR=./cmd/kubectl-passman
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build the binary
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Format code
fmt:
	$(GOFMT) -s -w .

# Run vet
vet:
	$(GOVET) ./...

# Run all linting
lint: fmt vet

# Install binary to $GOPATH/bin
install:
	$(GOCMD) install $(CMD_DIR)

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)

build-darwin:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)

build-windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

# Development helpers
dev-build: lint test build

# Show help
help:
	@echo "Available targets:"
	@echo "  build       Build the binary"
	@echo "  test        Run tests"
	@echo "  clean       Clean build artifacts"
	@echo "  fmt         Format code"
	@echo "  vet         Run go vet"
	@echo "  lint        Run all linting (fmt + vet)"
	@echo "  install     Install binary to \$$GOPATH/bin"
	@echo "  build-all   Build for multiple platforms"
	@echo "  dev-build   Run lint, test, and build"
	@echo "  help        Show this help message"
