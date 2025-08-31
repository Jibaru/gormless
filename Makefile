# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Test targets
test: ## Run all tests
	go test ./...

test-verbose: ## Run all tests with verbose output
	go test -v ./...

coverage: ## Generate test coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage-html: ## Generate test coverage report with HTML output
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build targets
build: ## Build the application
	go build -o gormless.exe .

build-linux: ## Build for Linux
	GOOS=linux GOARCH=amd64 go build -o gormless .

build-mac: ## Build for macOS
	GOOS=darwin GOARCH=amd64 go build -o gormless .

build-all: build build-linux build-mac ## Build for all platforms
