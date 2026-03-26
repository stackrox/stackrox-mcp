# Multi-stage Dockerfile for ACS MCP Server build on Konflux

# Stage 1: Builder - Build the Go binary
FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_golang_1.25@sha256:bd531796aacb86e4f97443797262680fbf36ca048717c00b6f4248465e1a7c0c AS builder

# Build arguments for application version and branding
ARG VERSION=dev
ARG SERVER_NAME="acs-mcp-server"
ARG PRODUCT_DISPLAY_NAME="Red Hat Advanced Cluster Security (ACS)"

# Set working directory
WORKDIR /workspace

# Copy source code
COPY . .

# Build the binary with optimizations
# Output to "/tmp" directory, because user can not copy built binary to "/workspace"
# Go build uses "venodr" mode and that fails, that's why explicit "-mod=mod" is set.
RUN RACE=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) \
    go build \
    -mod=mod \
    -ldflags="-w -s \
      -X 'github.com/stackrox/stackrox-mcp/internal/config.version=${VERSION}' \
      -X 'github.com/stackrox/stackrox-mcp/internal/config.serverName=${SERVER_NAME}' \
      -X 'github.com/stackrox/stackrox-mcp/internal/config.productDisplayName=${PRODUCT_DISPLAY_NAME}'" \
    -trimpath \
    -o /tmp/stackrox-mcp \
    ./cmd/stackrox-mcp


# Stage 2: Runtime base - used to preserve rpmdb when installing packages
FROM registry.access.redhat.com/ubi9/ubi-micro:latest@sha256:2173487b3b72b1a7b11edc908e9bbf1726f9df46a4f78fd6d19a2bab0a701f38 AS ubi-micro-base


# Stage 3: Package installer - installs ca-certificates and openssl into /ubi-micro-base-root/
FROM registry.access.redhat.com/ubi9/ubi:latest@sha256:05fa0100593c08b5e9dde684cd3eaa94b4d5d7b3cc09944f1f73924e49fde036 AS package_installer

# Copy ubi-micro base to /ubi-micro-base-root/ to preserve its rpmdb
COPY --from=ubi-micro-base / /ubi-micro-base-root/

# Install packages directly to /ubi-micro-base-root/ using --installroot
# Note: --setopt=reposdir=/etc/yum.repos.d instructs dnf to use repo configurations pointing to RPMs
# prefetched by Hermeto/Cachi2, instead of installroot's default UBI repos.
# hadolint ignore=DL3041 # We are installing ca-certificates and openssl only to include trusted certs.
RUN dnf install -y \
    --installroot=/ubi-micro-base-root/ \
    --releasever=9 \
    --setopt=install_weak_deps=False \
    --setopt=reposdir=/etc/yum.repos.d \
    --nodocs \
    ca-certificates \
    openssl && \
    dnf clean all --installroot=/ubi-micro-base-root/ && \
    rm -rf /ubi-micro-base-root/var/cache/*


# Stage 4: Runtime - Minimal runtime image
FROM ubi-micro-base

# Set default environment variables
ENV LOG_LEVEL=INFO

# Set working directory
WORKDIR /app

COPY --from=package_installer /ubi-micro-base-root/ /

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
