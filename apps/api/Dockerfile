# Multi-stage Dockerfile for KubeChat API Backend
# Container-first development with hot-reload support using Air

# Base stage with Go 1.23 Alpine
FROM golang:1.23.2-alpine AS base

# Install system dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Dependencies stage
FROM base AS deps

# Copy go mod and sum files
COPY go.mod go.sum ./

# Set working directory for Go module
WORKDIR /app

# Download dependencies
RUN go mod download && go mod verify

# Builder stage
FROM base AS builder

# Copy dependencies from deps stage
COPY --from=deps /go/pkg/mod /go/pkg/mod

# Copy source code
COPY . ./

# Set working directory
WORKDIR /app

# Build flags for optimized binary
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Build the application
RUN go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/kubechat-api \
    ./cmd/server

# Production stage
FROM alpine:3.18 AS production

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 -S kubechat
RUN adduser -S kubechat -u 1001 -G kubechat

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder --chown=kubechat:kubechat /app/kubechat-api .

# Switch to non-root user
USER kubechat

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Start the application
CMD ["./kubechat-api"]

# Development stage with hot-reload
FROM base AS development

# Install Air for hot-reload
RUN go install github.com/cosmtrek/air@latest

# Create non-root user for development
RUN addgroup -g 1001 -S kubechat
RUN adduser -S kubechat -u 1001 -G kubechat

# Create directories and set permissions
RUN mkdir -p /app && chown -R kubechat:kubechat /app
RUN mkdir -p /go && chown -R kubechat:kubechat /go

# Switch to development user
USER kubechat

# Set working directory
WORKDIR /app

# Copy Air configuration
COPY --chown=kubechat:kubechat apps/api/.air.toml ./apps/api/

# Copy source code
COPY --chown=kubechat:kubechat apps/api ./apps/api

# Set working directory for Go module
WORKDIR /app/apps/api

# Download dependencies as kubechat user
RUN go mod download

# Environment variables for development
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GIN_MODE=debug

# Expose port
EXPOSE 8080

# Health check for development
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Start development server with hot-reload
CMD ["air", "-c", ".air.toml"]

# Default target is production
FROM production