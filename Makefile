# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME=stackrox-mcp

# Version can be overridden with VERSION=x.y.z make build (default: extracted from git tags or use dev)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOCLEAN=$(GOCMD) clean

# Set the container runtime command - prefer podman, fallback to docker
DOCKER_CMD = $(shell command -v podman >/dev/null 2>&1 && echo podman || echo docker)

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

.PHONY: image
image: ## Build the docker image
	$(DOCKER_CMD) build \
		--build-arg VERSION=$(VERSION) \
		-t quay.io/stackrox-io/mcp:$(VERSION) \
		.

.PHONY: dockerfile-lint
dockerfile-lint: ## Run hadolint for Dockerfile
	$(DOCKER_CMD) run --rm -i --env HADOLINT_FAILURE_THRESHOLD=info ghcr.io/hadolint/hadolint < Dockerfile

.PHONY: helm-lint
helm-lint: ## Run helm lint for Helm chart
	helm lint charts/stackrox-mcp

.PHONY: test
test: ## Run unit tests
	$(GOTEST) -v ./...

.PHONY: e2e-smoke-test
e2e-smoke-test: ## Run E2E smoke test (build and verify mcpchecker)
	@cd e2e-tests && ./scripts/smoke-test.sh

.PHONY: e2e-test mock-start
e2e-test: ## Run E2E tests
	@cd e2e-tests && ./scripts/run-tests.sh --mock

.PHONY: test-coverage-and-junit
test-coverage-and-junit: ## Run unit tests with coverage and junit output
	go install github.com/jstemmer/go-junit-report/v2@v2.1.0
	$(GOTEST) -v -cover -race -coverprofile=$(COVERAGE_OUT) ./... 2>&1 | go-junit-report -set-exit-code -iocopy -out $(JUNIT_OUT)

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

.PHONY: proto-setup
proto-setup: ## Setup proto files from go mod cache
	@./scripts/setup-proto-files.sh

.PHONY: proto-generate
proto-generate: ## Generate proto descriptors for WireMock
	@./scripts/generate-proto-descriptors.sh

.PHONY: proto-clean
proto-clean: ## Clean generated proto files
	@rm -rf wiremock/proto/ wiremock/grpc/

.PHONY: proto-check
proto-check: ## Verify proto setup is correct
	@if [ ! -f wiremock/proto/descriptors/stackrox.pb ]; then \
		echo "❌ Proto descriptors not found"; \
		echo "Run: make proto-generate"; \
		exit 1; \
	fi
	@echo "✓ Proto descriptors present"

.PHONY: mock-download
mock-download: ## Download WireMock JARs
	@./scripts/download-wiremock.sh

.PHONY: mock-start
mock-start: proto-check ## Start WireMock mock Central locally
	@./scripts/start-mock-central.sh

.PHONY: mock-stop
mock-stop: ## Stop WireMock mock Central
	@./scripts/stop-mock-central.sh

.PHONY: mock-logs
mock-logs: ## View WireMock logs
	@tail -f wiremock/wiremock.log

.PHONY: mock-restart
mock-restart: mock-stop mock-start ## Restart WireMock

.PHONY: mock-status
mock-status: ## Check WireMock status
	@if [ -f wiremock/wiremock.pid ]; then \
		PID=$$(cat wiremock/wiremock.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "WireMock is running (PID: $$PID)"; \
		else \
			echo "WireMock PID file exists but process not running"; \
		fi \
	else \
		echo "WireMock is not running"; \
	fi

.PHONY: mock-test
mock-test: ## Run WireMock smoke tests
	@./scripts/smoke-test-wiremock.sh

.PHONY: clean
clean: ## Clean build artifacts and coverage files
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_OUT)
	rm -f $(LINT_OUT)
