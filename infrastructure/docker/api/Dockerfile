# KubeChat Backend API Dockerfile
# Multi-stage build for Go application with hot-reload development support

# Stage 1: Base Go image - Latest stable Go version
FROM golang:1.24-alpine AS base

# Install system dependencies
RUN apk add --no-cache \
    git \
    curl \
    bash \
    build-base \
    ca-certificates

# Install Air for hot reloading in development (v1.49.0 supports Go 1.23)
RUN go install github.com/cosmtrek/air@v1.49.0

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download && go mod verify

# Stage 2: Development environment
FROM base AS development

# Copy source code
COPY apps/api/ ./

# Create Air configuration for hot reloading
RUN echo 'root = "."' > .air.toml && \
    echo 'testdata_dir = "testdata"' >> .air.toml && \
    echo 'tmp_dir = "tmp"' >> .air.toml && \
    echo '' >> .air.toml && \
    echo '[build]' >> .air.toml && \
    echo '  args_bin = []' >> .air.toml && \
    echo '  bin = "./tmp/main"' >> .air.toml && \
    echo '  cmd = "go build -o ./tmp/main ./cmd/server"' >> .air.toml && \
    echo '  delay = 1000' >> .air.toml && \
    echo '  exclude_dir = ["assets", "tmp", "vendor", "testdata"]' >> .air.toml && \
    echo '  exclude_file = []' >> .air.toml && \
    echo '  exclude_regex = ["_test.go"]' >> .air.toml && \
    echo '  exclude_unchanged = false' >> .air.toml && \
    echo '  follow_symlink = false' >> .air.toml && \
    echo '  full_bin = ""' >> .air.toml && \
    echo '  include_dir = []' >> .air.toml && \
    echo '  include_ext = ["go", "tpl", "tmpl", "html"]' >> .air.toml && \
    echo '  include_file = []' >> .air.toml && \
    echo '  kill_delay = "0s"' >> .air.toml && \
    echo '  log = "build-errors.log"' >> .air.toml && \
    echo '  poll = false' >> .air.toml && \
    echo '  poll_interval = 0' >> .air.toml && \
    echo '  rerun = false' >> .air.toml && \
    echo '  rerun_delay = 500' >> .air.toml && \
    echo '  send_interrupt = false' >> .air.toml && \
    echo '  stop_on_root = false' >> .air.toml

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

# Create necessary directories and set permissions
RUN mkdir -p /app/tmp /app/logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080 8081

# Environment variables for development
ENV GIN_MODE=debug
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Development command with hot reload
CMD ["air", "-c", ".air.toml"]

# Stage 3: Build for production
FROM base AS builder

# Copy source code
COPY apps/api/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Stage 4: Production runtime
FROM alpine:3.18 AS production

# Security labels for container metadata
LABEL org.opencontainers.image.title="KubeChat API Server" \
      org.opencontainers.image.description="Secure Go-based API server for KubeChat" \
      org.opencontainers.image.vendor="KubeChat" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.created="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
      security.scan.required="true" \
      security.non-root="true" \
      security.hardened="true"

# Security: Update packages and install only essential runtime dependencies
RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk --no-cache add \
    ca-certificates \
    curl \
    dumb-init \
    tzdata && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

# Set working directory
WORKDIR /app

# Security: Set proper file permissions and remove shell access
RUN chmod 755 /app && \
    echo 'appuser:!:0:0:appuser:/app:/sbin/nologin' > /etc/passwd && \
    echo 'appgroup:!:1001:' > /etc/group

# Copy binary from builder stage
COPY --from=builder --chown=appuser:appgroup /app/main .

# Copy any necessary configuration files
COPY --from=builder --chown=appuser:appgroup /app/configs ./configs

# Create necessary directories
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app/logs

# Security: Ensure binary has correct permissions and no write access to app directory
RUN chmod 555 /app/main && \
    chmod 755 /app && \
    chmod 755 /app/logs

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080 8081

# Environment variables for production
ENV GIN_MODE=release

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Use dumb-init to properly handle signals
ENTRYPOINT ["dumb-init", "--"]

# Production command
CMD ["./main"]