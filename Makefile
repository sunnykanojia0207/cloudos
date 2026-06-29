# CloudOS — Makefile
# Targets: build, lint, test, clean, dev, tidy

GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: help build lint test test-race test-coverage clean dev fmt tidy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build all Go packages
	$(GO) build ./...

lint: ## Run golangci-lint
	$(GOLANGCI_LINT) run ./...

fmt: ## Format Go code
	$(GO) fmt ./...
	$(GO) mod tidy

test: ## Run all tests
	$(GO) test ./...

test-race: ## Run tests with race detector
	$(GO) test -race ./...

test-coverage: ## Run tests with coverage report
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

tidy: ## Tidy Go module dependencies
	$(GO) mod tidy

clean: ## Clean build artifacts
	$(GO) clean ./...
	rm -f coverage.out coverage.html
	rm -rf bin/

dev: ## Run CloudOS kernel in development mode
	@echo "Starting CloudOS kernel..."
	$(GO) run ./tools/cloudos
