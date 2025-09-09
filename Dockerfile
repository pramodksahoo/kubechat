# Build stage for Go backend
FROM golang:1.21-alpine AS go-builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kubechat ./cmd/server

# Build stage for React frontend
FROM node:18-alpine AS web-builder

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./
RUN npm ci

# Copy web source code
COPY web/ ./

# Build the React app
RUN npm run build

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

# Copy Go binary
COPY --from=go-builder /app/kubechat .

# Copy React build
COPY --from=web-builder /app/web/build ./web/build

# Create non-root user
RUN adduser -D -s /bin/sh kubechat
RUN chown -R kubechat:kubechat /root
USER kubechat

EXPOSE 8080

CMD ["./kubechat"]