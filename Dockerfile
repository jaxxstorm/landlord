# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download and cache dependencies before copying source
# This layer is cached unless go.mod or go.sum changes
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimized flags for production
# CGO_ENABLED=0 ensures static linking with no C dependencies
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o landlord \
    ./cmd/landlord

# Runtime stage
FROM debian:bookworm-slim AS runtime

# Install minimal runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -r -u 65532 -s /bin/false nonroot

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/landlord .

# Expose API port (default: 8080)
# Can be configured via HTTP_PORT environment variable
EXPOSE 8080

# Health check - use curl to check the /health endpoint
# Configure via HTTP_PORT, DB_PROVIDER, DB_SQLITE_PATH, DB_HOST, DB_PORT, etc.
# See README for configuration examples
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["curl", "-f", "http://localhost:8080/health", "||", "exit", "1"]

# Run as nonroot user (uid 65532 in distroless) for security
USER nonroot

# Set entrypoint to run the binary
ENTRYPOINT ["/app/landlord"]

# Default command (can be overridden)
CMD []
