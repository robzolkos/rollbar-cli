.PHONY: build test test-int test-e2e test-cover lint clean install all

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/robzolkos/rollbar-cli/internal/version.Version=$(VERSION) \
                     -X github.com/robzolkos/rollbar-cli/internal/version.Commit=$(COMMIT) \
                     -X github.com/robzolkos/rollbar-cli/internal/version.BuildDate=$(BUILD_DATE)"

# Default target
all: lint test build

# Build the binary
build:
	go build $(LDFLAGS) -o bin/rollbar ./cmd/rollbar

# Install to GOBIN
install:
	go install $(LDFLAGS) ./cmd/rollbar

# Run unit tests
test:
	go test -race -v ./...

# Run integration tests (with mock server)
test-int:
	go test -race -v -tags=integration ./...

# Run E2E tests (requires ROLLBAR_E2E_TOKEN env var)
test-e2e:
	@if [ -z "$$ROLLBAR_E2E_TOKEN" ]; then \
		echo "Error: ROLLBAR_E2E_TOKEN environment variable not set"; \
		echo "See e2e/README.md for setup instructions"; \
		exit 1; \
	fi
	go test -race -v -tags=e2e ./e2e/...

# Run all tests including E2E
test-all: test test-int test-e2e

# Test coverage
test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint with golangci-lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Generate shell completions
completions: build
	mkdir -p scripts/completions
	./bin/rollbar completion bash > scripts/completions/rollbar.bash
	./bin/rollbar completion zsh > scripts/completions/_rollbar
	./bin/rollbar completion fish > scripts/completions/rollbar.fish

# Development: build and run
run: build
	./bin/rollbar $(ARGS)

# Quick validation before commit
check: fmt lint test build
	@echo "All checks passed!"
