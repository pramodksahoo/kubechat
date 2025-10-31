# Kubechat Monorepo Structure Integration

We are building Kubechat directly within the Kubechat monorepo. While you should have this already, this document outlines the integrations and modules.

## Module structure

- **cmd/kubechat** (existing main entry point)
- **pkg/** - Shared utilities
- **internal/** - Domain specific modules
  - **internal/ai** - AI provider adapter layer
  - **internal/audit** - Audit logging
  - **internal/k8s** - Kubernetes adapters
  - **internal/api** - HTTP APIs
  - **internal/web** - Web serving and assets

# Go Modules

Single Go module with tidy use of go.mod. Use go1.23.

# Tests

Project uses Go tests.
