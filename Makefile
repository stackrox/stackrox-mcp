# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME=stackrox-mcp

# Version (can be overridden with VERSION=x.y.z make build)
VERSION?=0.1.0

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOCLEAN=$(GOCMD) clean

# Build flags
LDFLAGS=-ldflags "-X github.com/stackrox/stackrox-mcp/internal/server.version=$(VERSION)"

# Coverage files
COVERAGE_OUT=coverage.out

# JUnit files
JUNIT_OUT=junit.xml

# Lint files
LINT_OUT=report.xml

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/stackrox-mcp

.PHONY: test
test: ## Run unit tests with coverage
	go install github.com/jstemmer/go-junit-report/v2@v2.1.0
	$(GOTEST) -v -cover -coverprofile=$(COVERAGE_OUT) ./... -json 2>&1 | go-junit-report -parser gojson > tests.xml

.PHONY: coverage-html
coverage-html: test ## Generate and open HTML coverage report
	$(GOCMD) tool cover -html=$(COVERAGE_OUT)

.PHONY: fmt
fmt: ## Format Go code
	$(GOFMT) ./...

.PHONY: fmt-check
fmt-check: ## Check if Go code is formatted (fails if not)
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted:"; \
		gofmt -l .; \
		exit 1; \
	fi

.PHONY: lint
lint: ## Run golangci-lint
	go install -v "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6"
	golangci-lint run

.PHONY: clean
clean: ## Clean build artifacts and coverage files
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_OUT)
	rm -f $(LINT_OUT)
