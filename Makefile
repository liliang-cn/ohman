VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -ldflags "\
	-X github.com/liliang-cn/ohman/pkg/version.Version=$(VERSION) \
	-X github.com/liliang-cn/ohman/pkg/version.BuildTime=$(BUILD_TIME) \
	-X github.com/liliang-cn/ohman/pkg/version.Commit=$(COMMIT)"

BINARY := ohman
BUILD_DIR := bin

.PHONY: all
all: build

# Build
.PHONY: build
build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/ohman

# Development build (with debug info)
.PHONY: build-dev
build-dev:
	@echo "Building $(BINARY) for development..."
	@mkdir -p $(BUILD_DIR)
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY) ./cmd/ohman

# Cross-compile for all platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/ohman
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/ohman
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/ohman
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/ohman
	@echo "Build complete!"

# Install to system
.PHONY: install
install: build
	@echo "Installing $(BINARY) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/
	@echo "Installed successfully!"

# Uninstall
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY)..."
	sudo rm -f /usr/local/bin/$(BINARY)
	@echo "Uninstalled successfully!"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Run tests with coverage report
.PHONY: coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Generate mocks
.PHONY: mock
mock:
	@echo "Generating mocks..."
	go generate ./...

# Run program
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY)

# Show help
.PHONY: help
help:
	@echo "Oh Man! - Makefile Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make build       - Build the binary"
	@echo "  make build-dev   - Build with debug symbols"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make install     - Install to /usr/local/bin"
	@echo "  make uninstall   - Remove from /usr/local/bin"
	@echo "  make test        - Run tests"
	@echo "  make coverage    - Run tests with coverage"
	@echo "  make fmt         - Format code"
	@echo "  make tidy        - Tidy dependencies"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make run         - Build and run"
	@echo "  make help        - Show this help"
