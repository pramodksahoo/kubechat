# KubeChat Frontend Dockerfile
# Multi-stage production build for minimal Next.js application

# Build stage
FROM node:20.10-alpine AS builder

# Install system dependencies for build
RUN apk add --no-cache libc6-compat

# Install PNPM globally
RUN npm install -g pnpm@10

# Set working directory
WORKDIR /app

# Copy package configuration files
COPY pnpm-workspace.yaml ./
COPY pnpm-lock.yaml ./
COPY package.json ./
COPY apps/web/package.json ./apps/web/
COPY packages/shared/package.json ./packages/shared/
COPY packages/ui/package.json ./packages/ui/
COPY packages/config/package.json ./packages/config/

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy only necessary source files
COPY apps/web ./apps/web/
COPY packages ./packages/

# Build the application
RUN pnpm --filter web build

# Production stage
FROM node:20.10-alpine AS runner

# Security labels for container metadata
LABEL org.opencontainers.image.title="KubeChat Web Frontend" \
      org.opencontainers.image.description="Secure Next.js frontend for KubeChat" \
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
    dumb-init \
    curl \
    ca-certificates \
    tzdata && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Set working directory
WORKDIR /app

# Set environment variables
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Copy built application from builder stage
COPY --from=builder --chown=nextjs:nodejs /app/apps/web/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/apps/web/.next/static ./apps/web/.next/static
COPY --from=builder --chown=nextjs:nodejs /app/apps/web/public ./apps/web/public

# Security: Set proper file permissions and remove shell access
RUN chmod 755 /app && \
    echo 'nextjs:!:0:0:nextjs:/app:/sbin/nologin' > /etc/passwd && \
    echo 'nodejs:!:1001:' > /etc/group

# Create necessary directories with proper permissions
RUN mkdir -p /app/logs && \
    chown -R nextjs:nodejs /app/logs && \
    chmod 755 /app/logs

# Switch to non-root user
USER nextjs

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/api/health || exit 1

# Use dumb-init to properly handle signals
ENTRYPOINT ["dumb-init", "--"]

# Start the application
CMD ["node", "apps/web/server.js"]