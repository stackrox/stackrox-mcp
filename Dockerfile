# Multi-stage Dockerfile for StackRox MCP Server

# Used images
ARG GOLANG_BUILDER=registry.access.redhat.com/ubi10/go-toolset:1.25
ARG MCP_SERVER_BASE_IMAGE=registry.access.redhat.com/ubi10/ubi-micro:10.1

# Build arguments for multi-arch build support
ARG BUILDPLATFORM

# Stage 1: Builder - Build the Go binary
FROM --platform=$BUILDPLATFORM $GOLANG_BUILDER AS builder

# Build arguments for multi-arch target
ARG TARGETOS
ARG TARGETARCH

# Build arguments for application version
ARG VERSION=dev

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
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /tmp/stackrox-mcp \
    ./cmd/stackrox-mcp

# Stage 2: Runtime - Minimal runtime image
FROM $MCP_SERVER_BASE_IMAGE

# Set default environment variables
ENV LOG_LEVEL=INFO

# Set working directory
WORKDIR /app

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
