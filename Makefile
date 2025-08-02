# Excel Schema Generator - Makefile

.PHONY: help build test test-coverage lint clean install dev docs fmt vet security check-deps

# Variables
BINARY_NAME=excel-schema-generator
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WIN=$(BINARY_NAME).exe
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -s -w"

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .

build-macos: ## Build universal macOS binary
	@echo "Building universal macOS binary..."
	@chmod +x ./scripts/build_macos.sh
	./scripts/build_macos.sh

build-windows: ## Build Windows binary
	@echo "Building Windows binary..."
	@chmod +x ./scripts/build_windows.bat
	./scripts/build_windows.bat

# Test targets
test: ## Run tests
	@echo "Running tests..."
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

test-coverage-ci: ## Run tests with coverage for CI
	@echo "Running tests with coverage for CI..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < 80.0" | bc -l) -eq 1 ]; then \
		echo "Coverage $$COVERAGE% is below required 80%"; \
		exit 1; \
	fi

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Quality targets
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run --timeout=5m

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

security: ## Run security scan
	@echo "Running security scan..."
	gosec ./...

check-deps: ## Check for outdated dependencies
	@echo "Checking dependencies..."
	go list -u -m all
	go mod verify

# Development targets
dev: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	go mod verify
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/github-action-gosec@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development environment ready!"

install: build ## Build and install the application
	@echo "Installing $(BINARY_NAME)..."
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache
	go clean -testcache

docs: ## Generate documentation
	@echo "Generating documentation..."
	godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./bin/$(BINARY_NAME)

run-gui: build ## Build and run GUI mode
	@echo "Running $(BINARY_NAME) in GUI mode..."
	./bin/$(BINARY_NAME)

run-cli-help: build ## Build and show CLI help
	@echo "Running $(BINARY_NAME) CLI help..."
	./bin/$(BINARY_NAME) --help

# Release targets
release-prep: ## Prepare for release
	@echo "Preparing for release..."
	@echo "Current version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	$(MAKE) clean
	$(MAKE) test-coverage-ci
	$(MAKE) lint
	$(MAKE) security
	$(MAKE) build-all
	@echo "Release preparation complete!"

# Docker targets (if needed in future)
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -it $(BINARY_NAME):$(VERSION)

# Git helpers
git-tag: ## Create a git tag (usage: make git-tag VERSION=v1.0.0)
	@echo "Creating git tag $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# Quality gate
quality-gate: ## Run all quality checks
	@echo "Running quality gate..."
	$(MAKE) fmt
	$(MAKE) vet
	$(MAKE) lint
	$(MAKE) security
	$(MAKE) test-coverage-ci
	@echo "✅ All quality gates passed!"

# CI targets
ci: ## Run CI pipeline locally
	@echo "Running CI pipeline..."
	$(MAKE) quality-gate
	$(MAKE) build-all
	@echo "✅ CI pipeline completed successfully!"

# Show project info
info: ## Show project information
	@echo "Project: $(BINARY_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"
	@echo ""
	@echo "Available Make targets:"
	@$(MAKE) help