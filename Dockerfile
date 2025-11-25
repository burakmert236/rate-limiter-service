# Multi-stage build for smaller final image

# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
# git: for go mod download from git repositories
# ca-certificates: for HTTPS requests
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first (for better caching)
# Docker caches layers, so if go.mod/go.sum don't change, dependencies won't re-download
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0: Build a statically linked binary (no C dependencies)
# -ldflags="-w -s": Strip debug information to reduce binary size
#   -w: disable DWARF generation
#   -s: disable symbol table
# -trimpath: Remove file system paths from binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /build/ratelimiter-server \
    ./cmd/server

# Stage 2: Runtime stage
FROM alpine:3.19

# Install runtime dependencies
# ca-certificates: for HTTPS/TLS
# tzdata: for timezone support
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
# -D: Don't assign a password
# -H: Don't create home directory
# -u 1000: Set UID to 1000
RUN adduser -D -H -u 1000 appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/ratelimiter-server /app/ratelimiter-server

# Copy any additional files if needed (config templates, etc.)
# COPY --from=builder /build/configs /app/configs

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose gRPC port
EXPOSE 50051

# Health check
# grpc_health_probe is optional - for now we'll skip it
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ["/app/ratelimiter-server", "health"]

# Set environment variables with defaults
ENV REDIS_ADDR=redis:6379
ENV REDIS_PASSWORD=
ENV SERVER_PORT=50051
ENV LOG_LEVEL=info

# Run the binary
ENTRYPOINT ["/app/ratelimiter-server"]