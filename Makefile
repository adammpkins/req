.PHONY: build test lint package golden clean help

# Variables
BINARY_NAME=req
MAIN_PATH=./cmd/req
VERSION?=dev
BUILD_DIR=./bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the req binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run tests with race detector
	@echo "Running tests..."
	@go test -race -cover -v ./...

test-golden: ## Run golden file tests
	@echo "Running golden tests..."
	@go test -v ./tests -run TestGolden

golden: ## Regenerate golden test files
	@echo "Regenerating golden files..."
	@go test ./tests -run TestGolden -update

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@golangci-lint run

lint-fix: ## Run golangci-lint with auto-fix
	@golangci-lint run --fix

vulncheck: ## Run govulncheck
	@echo "Running govulncheck..."
	@govulncheck ./...

package: ## Build release artifacts locally (requires goreleaser)
	@echo "Building release artifacts..."
	@goreleaser build --snapshot --clean

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean -cache

install: build ## Install binary to GOPATH/bin
	@go install $(MAIN_PATH)

