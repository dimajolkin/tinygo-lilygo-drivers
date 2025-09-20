# Makefile for TinyGo LilyGo Drivers

.PHONY: help clean test test-integration test-coverage lint fmt vet build build-examples quality-check install-tools

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Clean build artifacts
clean: ## Clean build artifacts and temporary files
	@echo "🧹 Cleaning build artifacts..."
	@go clean -cache -testcache -modcache
	@rm -f coverage.out
	@find . -name "*.bin" -delete
	@find . -name "*.hex" -delete
	@echo "✅ Clean complete"

# Testing
test: ## Run unit tests
	@echo "🧪 Running unit tests..."
	@go test -v -race -timeout=10m ./...

test-integration: ## Run integration tests (requires hardware)
	@echo "🔌 Running integration tests..."
	@go test -v -tags=integration -timeout=30m ./...

test-coverage: ## Run tests with coverage report
	@echo "📊 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | grep total
	@echo "📄 Coverage report generated: coverage.html"

# Code quality
fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	@gofmt -w -s .
	@echo "✅ Code formatted"

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Go vet passed"

lint: ## Run golangci-lint
	@echo "🔧 Running golangci-lint..."
	@golangci-lint run --timeout=10m
	@echo "✅ Linting passed"

quality-check: fmt vet lint test-coverage build ## Run comprehensive quality checks
	@echo "🔍 Running comprehensive quality checks..."
	@echo "✅ All quality checks completed successfully!"

# Building
build: ## Build all examples with regular Go (syntax check)
	@echo "🏗️  Building examples with Go..."
	@for dir in examples/*/; do \
		if [ -d "$$dir" ]; then \
			echo "Building $$(basename $$dir)..."; \
			(cd "$$dir" && go build -o /tmp/$$(basename $$dir)-test . && rm -f /tmp/$$(basename $$dir)-test); \
		fi \
	done
	@echo "✅ All examples build successfully"

build-examples: ## Build examples with TinyGo for ESP32-S3
	@echo "🤖 Building examples with TinyGo..."
	@if ! command -v tinygo >/dev/null 2>&1; then \
		echo "❌ TinyGo not found. Please install TinyGo first."; \
		exit 1; \
	fi
	@for dir in examples/*/; do \
		if [ -d "$$dir" ]; then \
			example=$$(basename $$dir); \
			echo "Building $$example for ESP32-S3..."; \
			(cd "$$dir" && tinygo build -target=esp32s3 -o $$example.bin .); \
		fi \
	done
	@echo "✅ All examples built with TinyGo"

build-all-targets: ## Build examples for all supported targets
	@echo "🎯 Building examples for all targets..."
	@if ! command -v tinygo >/dev/null 2>&1; then \
		echo "❌ TinyGo not found. Please install TinyGo first."; \
		exit 1; \
	fi
	@targets="esp32s3 esp32c3 pico"; \
	for dir in examples/*/; do \
		if [ -d "$$dir" ]; then \
			example=$$(basename $$dir); \
			echo "Building $$example for all targets..."; \
			for target in $$targets; do \
				echo "  - Building for $$target..."; \
				(cd "$$dir" && tinygo build -target=$$target -o $$example-$$target.bin . 2>/dev/null) || echo "    ⚠️  $$target not supported"; \
			done \
		fi \
	done
	@echo "✅ Multi-target build complete"

# Development setup
install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "📦 Installing pre-commit (requires Python)..."
	@pip3 install pre-commit || echo "⚠️  Could not install pre-commit (pip3 not found)"
	@echo "✅ Development tools installed"

setup-hooks: ## Setup pre-commit hooks
	@echo "🪝 Setting up pre-commit hooks..."
	@pre-commit install || echo "❌ Could not setup pre-commit hooks (pre-commit not found)"
	@echo "✅ Pre-commit hooks installed"

# Module management
mod-tidy: ## Run go mod tidy
	@echo "📦 Running go mod tidy..."
	@go mod tidy
	@echo "✅ Go modules tidied"

mod-verify: ## Verify go modules
	@echo "🔐 Verifying go modules..."
	@go mod verify
	@echo "✅ Go modules verified"

# Release preparation
pre-release: quality-check build build-examples ## Run all checks before release
	@echo "🚀 Pre-release checks complete!"
	@echo ""
	@echo "Ready to create a release. Run:"
	@echo "  git tag -a v1.0.0 -m 'Release v1.0.0'"
	@echo "  git push origin v1.0.0"

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	@go doc -all ./st7789 > docs/st7789-api.txt || mkdir -p docs && go doc -all ./st7789 > docs/st7789-api.txt
	@echo "✅ Documentation generated in docs/"

# Quick development workflow
dev: fmt vet test ## Quick development workflow (format, vet, test)
	@echo "✅ Development workflow complete"

# Full CI simulation
ci: quality-check test-coverage build build-examples ## Simulate full CI pipeline locally
	@echo "🎉 Full CI simulation complete!"
