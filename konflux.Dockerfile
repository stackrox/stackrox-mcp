# Multi-stage Dockerfile for ACS MCP Server build on Konflux

# Stage 1: Builder - Build the Go binary
FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_golang_1.25@sha256:bd531796aacb86e4f97443797262680fbf36ca048717c00b6f4248465e1a7c0c AS builder

# Build arguments for application version and branding
ARG VERSION=dev
ARG SERVER_NAME="acs-mcp-server"
ARG PRODUCT_DISPLAY_NAME="Red Hat Advanced Cluster Security (ACS)"

# Set working directory
WORKDIR /workspace

# Copy go module files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (cached layer)
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
# Output to "/tmp" directory, because user can not copy built binary to "/workspace"
# Go build uses "venodr" mode and that fails, that's why explicit "-mod=mod" is set.
RUN RACE=0 CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) \
    go build \
    -mod=mod \
    -ldflags="-w -s \
      -X 'github.com/stackrox/stackrox-mcp/internal/config.version=${VERSION}' \
      -X 'github.com/stackrox/stackrox-mcp/internal/config.serverName=${SERVER_NAME}' \
      -X 'github.com/stackrox/stackrox-mcp/internal/config.productDisplayName=${PRODUCT_DISPLAY_NAME}'" \
    -trimpath \
    -o /tmp/stackrox-mcp \
    ./cmd/stackrox-mcp

# Stage 2: Runtime - Minimal runtime image
FROM registry.access.redhat.com/ubi9/ubi-micro@sha256:093a704be0eaef9bb52d9bc0219c67ee9db13c2e797da400ddb5d5ae6849fa10

# Set default environment variables
ENV LOG_LEVEL=INFO

# Set working directory
WORKDIR /app

# Copy trusted certificates from builder
COPY --from=builder /etc/pki/ca-trust/extracted/ /etc/pki/ca-trust/extracted/
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /tmp/stackrox-mcp /app/stackrox-mcp

# Set ownership for OpenShift arbitrary UID support
# Files owned by 4000, group 0 (root), with group permissions matching user
RUN chown -R 4000:0 /app && \
    chmod -R g=u /app

# Switch to non-root user (can be overridden by OpenShift SCC)
USER 4000

# Expose port for MCP server
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/stackrox-mcp"]
