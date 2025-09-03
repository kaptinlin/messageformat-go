# MessageFormat 2.0 Go Implementation Makefile
# Set up GOBIN so that our binaries are installed to ./bin instead of $GOPATH/bin.
PROJECT_ROOT = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
export GOBIN = $(PROJECT_ROOT)/bin

GOLANGCI_LINT_VERSION := $(shell $(GOBIN)/golangci-lint version --format short 2>/dev/null || $(GOBIN)/golangci-lint version --short 2>/dev/null)
REQUIRED_GOLANGCI_LINT_VERSION := $(shell cat .golangci.version)

# Directories containing independent Go modules.
MODULE_DIRS = .

.PHONY: all
all: submodules lint test

.PHONY: help
help: ## Show this help message
	@echo "MessageFormat 2.0 Go Implementation"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean build artifacts
	@rm -rf $(GOBIN)
	@go clean -cache -testcache

.PHONY: submodules
submodules: ## Initialize and update git submodules (required for official tests)
	@echo "[setup] Initializing git submodules..."
	@git submodule update --init --recursive

.PHONY: deps
deps: ## Download Go module dependencies
	@echo "[deps] Downloading dependencies..."
	@go mod download
	@go mod tidy

.PHONY: test
test: submodules ## Run all tests including official test suite
	@echo "[test] Running all tests..."
	@$(foreach mod,$(MODULE_DIRS),(cd $(mod) && go test -race ./...) &&) true

.PHONY: test-unit
test-unit: ## Run unit tests only (excluding official test suite)
	@echo "[test] Running unit tests..."
	@go test -race ./pkg/... ./internal/... .

.PHONY: test-official
test-official: submodules ## Run official MessageFormat 2.0 test suite only
	@echo "[test] Running official test suite..."
	@go test -race ./tests/

.PHONY: test-coverage
test-coverage: submodules ## Run tests with coverage report
	@echo "[test] Running tests with coverage..."
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "[test] Coverage report generated: coverage.html"

.PHONY: test-verbose
test-verbose: submodules ## Run tests with verbose output
	@echo "[test] Running tests with verbose output..."
	@go test -race -v ./...

.PHONY: bench
bench: ## Run benchmarks
	@echo "[bench] Running benchmarks..."
	@go test -bench=. -benchmem ./...

.PHONY: lint
lint: golangci-lint tidy-lint ## Run all linters

# Install golangci-lint with the required version in GOBIN if it is not already installed.
.PHONY: install-golangci-lint
install-golangci-lint:
	@# Ensure $(GOBIN) exists
	@mkdir -p $(GOBIN)
	@# Install only when version mismatch to avoid unnecessary downloads
	@if [ "$(GOLANGCI_LINT_VERSION)" != "$(REQUIRED_GOLANGCI_LINT_VERSION)" ]; then \
			echo "[lint] installing golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) (current: $(GOLANGCI_LINT_VERSION))"; \
			curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v$(REQUIRED_GOLANGCI_LINT_VERSION); \
	else \
			echo "[lint] golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) already installed"; \
		fi

.PHONY: golangci-lint
golangci-lint: install-golangci-lint ## Run golangci-lint
	@echo "[lint] $(shell $(GOBIN)/golangci-lint version)"
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] golangci-lint: $(mod)" && \
		$(GOBIN)/golangci-lint run --timeout=10m --path-prefix $(mod)) &&) true

.PHONY: tidy-lint
tidy-lint: ## Check if go.mod and go.sum are tidy
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] mod tidy: $(mod)" && \
		go mod tidy && \
		git diff --exit-code -- go.mod go.sum) &&) true

.PHONY: fmt
fmt: ## Format Go code
	@echo "[fmt] Formatting Go code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "[vet] Running go vet..."
	@go vet ./...



.PHONY: examples
examples: ## Run all examples
	@echo "[examples] Running examples..."
	@cd examples/basic && go run main.go
	@echo ""
	@echo "[examples] All examples completed successfully"

.PHONY: verify
verify: submodules deps fmt vet lint test ## Run all verification steps (format, vet, lint, test)
	@echo "[verify] All verification steps completed successfully"

.PHONY: ci
ci: verify ## Run CI pipeline locally
	@echo "[ci] CI pipeline completed successfully"
