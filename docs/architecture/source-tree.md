# KubeChat Source Tree & Project Structure

## Overview

This document defines the definitive source code organization for KubeChat, ensuring consistent development patterns across the monorepo.

**Last Updated:** 2025-01-10  
**Repository:** https://github.com/pramodksahoo/kubechat  
**Branch:** develop

---

## Monorepo Structure

```plaintext
kubechat/
├── apps/                          # Application services
│   ├── web/                       # React TypeScript frontend
│   │   ├── src/
│   │   │   ├── components/        # UI components
│   │   │   ├── pages/            # Next.js pages
│   │   │   ├── hooks/            # Custom React hooks
│   │   │   ├── stores/           # Zustand state stores
│   │   │   ├── services/         # API communication layer
│   │   │   ├── types/            # TypeScript type definitions
│   │   │   └── utils/            # Utility functions
│   │   ├── public/               # Static assets
│   │   ├── tests/                # Frontend tests
│   │   ├── package.json
│   │   ├── next.config.js
│   │   ├── tailwind.config.js
│   │   └── tsconfig.json
│   │
│   └── api/                       # Go backend services
│       ├── cmd/
│       │   ├── server/           # Main API server
│       │   ├── ai-processor/     # AI request processor
│       │   └── audit-writer/     # Audit log service
│       ├── internal/
│       │   ├── handlers/         # HTTP request handlers
│       │   ├── services/         # Business logic
│       │   ├── models/           # Data models
│       │   ├── k8s/              # Kubernetes client utilities
│       │   ├── ai/               # AI service integrations
│       │   ├── audit/            # Audit logging
│       │   ├── auth/             # Authentication/authorization
│       │   └── config/           # Configuration management
│       ├── pkg/                  # Public Go packages
│       ├── tests/                # Backend tests
│       ├── go.mod
│       └── go.sum
│
├── packages/                      # Shared packages
│   ├── shared/                    # Shared TypeScript types
│   │   ├── src/
│   │   │   ├── types/            # Common type definitions
│   │   │   ├── constants/        # Shared constants
│   │   │   └── utils/            # Cross-platform utilities
│   │   ├── package.json
│   │   └── tsconfig.json
│   │
│   ├── ui/                        # Component library
│   │   ├── src/
│   │   │   ├── components/       # Reusable UI components
│   │   │   ├── styles/           # Shared styles
│   │   │   └── icons/            # Icon components
│   │   ├── storybook/            # Component documentation
│   │   ├── package.json
│   │   └── tsconfig.json
│   │
│   └── config/                    # Shared configuration
│       ├── eslint/               # ESLint configurations
│       ├── typescript/           # TypeScript configurations
│       ├── tailwind/             # Tailwind configurations
│       └── vite/                 # Vite configurations
│
├── infrastructure/                # DevOps and deployment
│   ├── helm/                     # Kubernetes deployment charts
│   │   ├── kubechat/             # Main application chart
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   ├── values-dev.yaml
│   │   │   ├── values-prod.yaml
│   │   │   └── templates/
│   │   │       ├── deployment.yaml
│   │   │       ├── service.yaml
│   │   │       ├── ingress.yaml
│   │   │       ├── configmap.yaml
│   │   │       ├── secret.yaml
│   │   │       └── rbac.yaml
│   │   │
│   │   ├── ollama/               # AI inference chart
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   └── templates/
│   │   │
│   │   ├── postgresql/           # Database chart
│   │   └── redis/                # Cache chart
│   │
│   ├── docker/                   # Container configurations
│   │   ├── web/
│   │   │   ├── Dockerfile
│   │   │   └── .dockerignore
│   │   ├── api/
│   │   │   ├── Dockerfile
│   │   │   └── .dockerignore
│   │   └── ollama/
│   │       └── Dockerfile
│   │
│   ├── scripts/                  # Build automation
│   │   ├── build.sh              # Container build scripts
│   │   ├── deploy.sh             # Deployment scripts
│   │   ├── dev-setup.sh          # Development environment
│   │   └── ci-cd.sh              # CI/CD pipeline scripts
│   │
│   └── k8s/                      # Raw Kubernetes manifests
│       ├── namespace.yaml
│       ├── rbac.yaml
│       └── network-policies.yaml
│
├── docs/                         # Project documentation
│   ├── architecture/             # Technical architecture
│   │   ├── architecture.md       # Main architecture document
│   │   ├── tech-stack.md         # Technology stack
│   │   ├── coding-standards.md   # Coding standards
│   │   ├── source-tree.md        # This document
│   │   ├── deployment.md         # Deployment guide
│   │   └── air-gapped.md         # Air-gapped setup
│   │
│   ├── api/                      # API documentation
│   │   ├── openapi.yaml          # OpenAPI specification
│   │   ├── endpoints.md          # Endpoint documentation
│   │   └── authentication.md    # Auth documentation
│   │
│   ├── development/              # Development guides
│   │   ├── getting-started.md    # Quick start guide
│   │   ├── testing.md            # Testing strategies
│   │   ├── debugging.md          # Debugging guide
│   │   └── troubleshooting.md    # Common issues
│   │
│   └── user/                     # User documentation
│       ├── installation.md       # Installation guide
│       ├── user-guide.md         # User manual
│       └── faq.md                # Frequently asked questions
│
├── tests/                        # Cross-service integration tests
│   ├── e2e/                      # End-to-end tests (Playwright)
│   ├── integration/              # Service integration tests
│   └── performance/              # Performance benchmarks
│
├── .github/                      # GitHub configuration
│   ├── workflows/                # CI/CD workflows
│   │   ├── ci.yml                # Continuous integration
│   │   ├── cd.yml                # Continuous deployment
│   │   └── security.yml          # Security scanning
│   ├── ISSUE_TEMPLATE/           # Issue templates
│   └── pull_request_template.md  # PR template
│
├── .bmad-core/                   # BMAD architecture tools
├── poc/                          # Proof of concept files (reference)
├── .gitignore                    # Git ignore rules
├── .env.example                  # Environment template
├── pnpm-workspace.yaml           # PNPM workspace config
├── Makefile                      # Development commands
├── README.md                     # Project overview
└── CONTRIBUTING.md               # Contribution guidelines
```

---

## Application Layer Details

### Frontend (`apps/web/`)

**Key Directories:**

```plaintext
apps/web/src/
├── components/                   # UI Components
│   ├── common/                   # Reusable components
│   │   ├── Button/
│   │   ├── Modal/
│   │   ├── Spinner/
│   │   └── Tooltip/
│   ├── layout/                   # Layout components
│   │   ├── Header/
│   │   ├── Sidebar/
│   │   ├── Footer/
│   │   └── Navigation/
│   ├── forms/                    # Form components
│   │   ├── LoginForm/
│   │   ├── QueryForm/
│   │   └── ConfigForm/
│   └── k8s/                      # Kubernetes-specific components
│       ├── ResourceList/
│       ├── PodViewer/
│       ├── LogViewer/
│       └── MetricsChart/
│
├── pages/                        # Next.js pages
│   ├── api/                      # API routes (Next.js)
│   ├── auth/                     # Authentication pages
│   │   ├── login.tsx
│   │   └── callback.tsx
│   ├── dashboard/                # Main dashboard
│   │   ├── index.tsx
│   │   └── [cluster].tsx
│   ├── clusters/                 # Cluster management
│   │   ├── index.tsx
│   │   ├── [id].tsx
│   │   └── add.tsx
│   ├── query/                    # Natural language query
│   │   ├── index.tsx
│   │   └── history.tsx
│   ├── monitoring/               # Monitoring views
│   │   ├── index.tsx
│   │   ├── alerts.tsx
│   │   └── metrics.tsx
│   ├── settings/                 # Configuration
│   │   ├── index.tsx
│   │   ├── security.tsx
│   │   └── integrations.tsx
│   ├── _app.tsx                  # App wrapper
│   ├── _document.tsx             # Document wrapper
│   └── index.tsx                 # Landing page
│
├── hooks/                        # Custom React hooks
│   ├── useAuth.ts                # Authentication hook
│   ├── useWebSocket.ts           # WebSocket connection
│   ├── useK8sResources.ts        # Kubernetes data fetching
│   ├── useAI.ts                  # AI query processing
│   └── useLocalStorage.ts        # Local storage management
│
├── stores/                       # Zustand state stores
│   ├── authStore.ts              # User authentication state
│   ├── clusterStore.ts           # Cluster connection state
│   ├── queryStore.ts             # Query history and state
│   ├── uiStore.ts                # UI state (modals, themes)
│   └── settingsStore.ts          # User preferences
│
├── services/                     # API communication layer
│   ├── api.ts                    # Base API client
│   ├── auth.ts                   # Authentication API
│   ├── clusters.ts               # Cluster management API
│   ├── query.ts                  # AI query API
│   ├── monitoring.ts             # Monitoring data API
│   └── websocket.ts              # WebSocket client
│
├── types/                        # TypeScript definitions
│   ├── api.ts                    # API response types
│   ├── auth.ts                   # Authentication types
│   ├── cluster.ts                # Cluster types
│   ├── query.ts                  # Query and AI types
│   └── ui.ts                     # UI component types
│
└── utils/                        # Utility functions
    ├── formatting.ts             # Data formatting
    ├── validation.ts             # Input validation
    ├── constants.ts              # Application constants
    └── helpers.ts                # General helpers
```

### Backend (`apps/api/`)

**Key Directories:**

```plaintext
apps/api/
├── cmd/                          # Application entry points
│   ├── server/                   # Main API server
│   │   ├── main.go
│   │   ├── routes.go
│   │   └── middleware.go
│   ├── ai-processor/             # AI processing service
│   │   ├── main.go
│   │   └── processor.go
│   └── audit-writer/             # Audit logging service
│       ├── main.go
│       └── writer.go
│
├── internal/                     # Private application code
│   ├── handlers/                 # HTTP request handlers
│   │   ├── auth.go               # Authentication endpoints
│   │   ├── clusters.go           # Cluster management
│   │   ├── query.go              # AI query processing
│   │   ├── monitoring.go         # Monitoring endpoints
│   │   ├── websocket.go          # WebSocket handlers
│   │   └── health.go             # Health checks
│   │
│   ├── services/                 # Business logic services
│   │   ├── auth_service.go       # Authentication service
│   │   ├── cluster_service.go    # Cluster management
│   │   ├── query_service.go      # Query processing
│   │   ├── monitoring_service.go # Monitoring service
│   │   └── audit_service.go      # Audit logging
│   │
│   ├── models/                   # Data models
│   │   ├── user.go               # User model
│   │   ├── cluster.go            # Cluster model
│   │   ├── query.go              # Query model
│   │   ├── audit.go              # Audit model
│   │   └── monitoring.go         # Monitoring models
│   │
│   ├── k8s/                      # Kubernetes client utilities
│   │   ├── client.go             # Kubernetes client wrapper
│   │   ├── resources.go          # Resource operations
│   │   ├── auth.go               # K8s authentication
│   │   └── discovery.go          # API discovery
│   │
│   ├── ai/                       # AI service integrations
│   │   ├── ollama.go             # Ollama client
│   │   ├── openai.go             # OpenAI client
│   │   ├── processor.go          # Query processing
│   │   └── models.go             # AI model management
│   │
│   ├── audit/                    # Audit logging
│   │   ├── logger.go             # Audit logger
│   │   ├── storage.go            # Audit storage
│   │   └── integrity.go          # Cryptographic integrity
│   │
│   ├── auth/                     # Authentication/authorization
│   │   ├── jwt.go                # JWT handling
│   │   ├── rbac.go               # Role-based access control
│   │   ├── session.go            # Session management
│   │   └── middleware.go         # Auth middleware
│   │
│   └── config/                   # Configuration management
│       ├── config.go             # Configuration loader
│       ├── env.go                # Environment variables
│       ├── database.go           # Database configuration
│       └── kubernetes.go         # Kubernetes configuration
│
├── pkg/                          # Public Go packages
│   ├── logger/                   # Structured logging
│   ├── database/                 # Database utilities
│   ├── cache/                    # Redis cache utilities
│   └── validation/               # Input validation
│
└── tests/                        # Backend tests
    ├── unit/                     # Unit tests
    ├── integration/              # Integration tests
    └── fixtures/                 # Test fixtures
```

---

## Package Layer Details

### Shared Package (`packages/shared/`)

```plaintext
packages/shared/src/
├── types/                        # Common TypeScript types
│   ├── api/                      # API-related types
│   │   ├── requests.ts           # Request types
│   │   ├── responses.ts          # Response types
│   │   └── errors.ts             # Error types
│   ├── auth/                     # Authentication types
│   │   ├── user.ts               # User types
│   │   ├── session.ts            # Session types
│   │   └── permissions.ts        # Permission types
│   ├── k8s/                      # Kubernetes types
│   │   ├── resources.ts          # Resource types
│   │   ├── clusters.ts           # Cluster types
│   │   └── events.ts             # Event types
│   └── ui/                       # UI component types
│       ├── components.ts         # Component props
│       ├── themes.ts             # Theme types
│       └── layouts.ts            # Layout types
│
├── constants/                    # Shared constants
│   ├── api.ts                    # API constants
│   ├── ui.ts                     # UI constants
│   ├── k8s.ts                    # Kubernetes constants
│   └── errors.ts                 # Error codes
│
└── utils/                        # Cross-platform utilities
    ├── validation.ts             # Validation functions
    ├── formatting.ts             # Data formatting
    ├── dates.ts                  # Date utilities
    └── crypto.ts                 # Cryptographic utilities
```

### UI Component Library (`packages/ui/`)

```plaintext
packages/ui/src/
├── components/                   # Reusable UI components
│   ├── atoms/                    # Atomic design - atoms
│   │   ├── Button/
│   │   ├── Input/
│   │   ├── Label/
│   │   ├── Icon/
│   │   └── Badge/
│   ├── molecules/                # Atomic design - molecules
│   │   ├── SearchBox/
│   │   ├── Dropdown/
│   │   ├── FormField/
│   │   └── Card/
│   ├── organisms/                # Atomic design - organisms
│   │   ├── Header/
│   │   ├── Sidebar/
│   │   ├── DataTable/
│   │   └── Modal/
│   └── templates/                # Atomic design - templates
│       ├── DashboardTemplate/
│       ├── FormTemplate/
│       └── ListTemplate/
│
├── styles/                       # Shared styles
│   ├── globals.css               # Global styles
│   ├── tokens.css                # Design tokens
│   ├── themes/                   # Theme definitions
│   │   ├── light.css
│   │   ├── dark.css
│   │   └── high-contrast.css
│   └── utilities.css             # Utility classes
│
└── icons/                        # Icon components
    ├── system/                   # System icons
    ├── k8s/                      # Kubernetes-specific icons
    └── brands/                   # Brand icons
```

---

## Infrastructure Layer Details

### Helm Charts (`infrastructure/helm/`)

```plaintext
infrastructure/helm/
├── kubechat/                     # Main application chart
│   ├── Chart.yaml                # Chart metadata
│   ├── Chart.lock                # Dependency lock
│   ├── values.yaml               # Default values
│   ├── values-dev.yaml           # Development values
│   ├── values-staging.yaml       # Staging values
│   ├── values-prod.yaml          # Production values
│   ├── values-air-gapped.yaml    # Air-gapped deployment
│   │
│   └── templates/                # Kubernetes manifests
│       ├── _helpers.tpl          # Template helpers
│       ├── namespace.yaml        # Namespace definition
│       ├── configmap.yaml        # Configuration
│       ├── secret.yaml           # Secrets
│       ├── serviceaccount.yaml   # Service account
│       ├── rbac.yaml             # RBAC rules
│       ├── deployment-api.yaml   # API deployment
│       ├── deployment-web.yaml   # Frontend deployment
│       ├── service-api.yaml      # API service
│       ├── service-web.yaml      # Frontend service
│       ├── ingress.yaml          # Ingress rules
│       ├── hpa.yaml              # Horizontal pod autoscaler
│       ├── pdb.yaml              # Pod disruption budget
│       └── monitoring.yaml       # Monitoring configuration
│
├── dependencies/                 # Dependency charts
│   ├── postgresql/               # Database chart
│   ├── redis/                    # Cache chart
│   ├── ollama/                   # AI inference chart
│   └── monitoring/               # Monitoring stack
│
└── scripts/                      # Helm automation
    ├── install.sh                # Installation script
    ├── upgrade.sh                # Upgrade script
    └── uninstall.sh              # Uninstallation script
```

### Docker Configurations (`infrastructure/docker/`)

```plaintext
infrastructure/docker/
├── web/                          # Frontend container
│   ├── Dockerfile                # Multi-stage build
│   ├── .dockerignore             # Build exclusions
│   └── nginx.conf                # Nginx configuration
│
├── api/                          # Backend container
│   ├── Dockerfile                # Multi-stage build
│   ├── .dockerignore             # Build exclusions
│   └── entrypoint.sh             # Container entrypoint
│
├── ollama/                       # AI inference container
│   ├── Dockerfile                # Ollama with models
│   ├── models/                   # Pre-downloaded models
│   └── init.sh                   # Model initialization
│
└── base/                         # Base images
    ├── node.Dockerfile           # Node.js base
    ├── go.Dockerfile             # Go base
    └── alpine.Dockerfile         # Alpine base
```

---

## File Naming Conventions

### TypeScript/JavaScript Files

```plaintext
# Components (PascalCase)
LoginForm.tsx
UserProfile.tsx
ClusterDashboard.tsx

# Hooks (camelCase with 'use' prefix)
useAuth.ts
useWebSocket.ts
useK8sResources.ts

# Stores (camelCase with 'Store' suffix)
authStore.ts
clusterStore.ts
uiStore.ts

# Services (camelCase)
apiClient.ts
authService.ts
queryProcessor.ts

# Types (camelCase)
apiTypes.ts
userTypes.ts
clusterTypes.ts

# Utils (camelCase)
dateUtils.ts
validationUtils.ts
formatUtils.ts
```

### Go Files

```plaintext
# Handlers (snake_case)
auth_handler.go
cluster_handler.go
query_handler.go

# Services (snake_case with 'service' suffix)
auth_service.go
cluster_service.go
monitoring_service.go

# Models (snake_case)
user.go
cluster.go
audit_log.go

# Tests (snake_case with '_test' suffix)
auth_service_test.go
cluster_handler_test.go
query_processor_test.go
```

### Configuration Files

```plaintext
# Kubernetes manifests (kebab-case)
deployment-api.yaml
service-web.yaml
configmap-app.yaml

# Docker files (PascalCase or specific)
Dockerfile
Dockerfile.dev
.dockerignore

# Configuration files (lowercase)
package.json
tsconfig.json
go.mod
Makefile
```

---

## Import/Export Patterns

### TypeScript Import Organization

```typescript
// 1. Node modules
import React from 'react'
import { NextPage } from 'next'
import { useRouter } from 'next/router'

// 2. Internal packages (packages/)
import { ApiResponse, UserType } from '@kubechat/shared'
import { Button, Modal } from '@kubechat/ui'

// 3. Internal application modules (src/)
import { useAuth } from '@/hooks/useAuth'
import { authStore } from '@/stores/authStore'
import { apiClient } from '@/services/api'

// 4. Relative imports
import { LoginFormProps } from './types'
import { validateCredentials } from '../utils/validation'
```

### Go Import Organization

```go
package main

import (
    // 1. Standard library
    "context"
    "fmt"
    "net/http"
    
    // 2. External packages
    "github.com/gin-gonic/gin"
    "k8s.io/client-go/kubernetes"
    
    // 3. Internal packages
    "github.com/pramodksahoo/kubechat/internal/handlers"
    "github.com/pramodksahoo/kubechat/internal/services"
    "github.com/pramodksahoo/kubechat/pkg/logger"
)
```

---

## Development Workflow Integration

### Container-First Development

```bash
# All development happens in containers
make dev-build        # Build all containers
make dev-deploy       # Deploy to local Kubernetes
make dev-logs         # View application logs
make dev-shell-api    # Shell into API container
make dev-shell-web    # Shell into web container
```

### Testing Structure

```plaintext
tests/
├── unit/                         # Component/function level
├── integration/                  # Service integration
├── e2e/                          # End-to-end scenarios
│   ├── auth/                     # Authentication flows
│   ├── clusters/                 # Cluster management
│   ├── query/                    # AI query processing
│   └── monitoring/               # Monitoring features
└── performance/                  # Load and performance
```

---

## Next Steps

1. **Validate Structure:** Review monorepo organization with team
2. **Create Templates:** Generate code templates for consistent patterns
3. **Setup Tooling:** Configure linting, formatting, and validation
4. **Document Patterns:** Create detailed coding standards documentation

For questions about project structure or file organization, consult the architecture team or create an issue in the repository.

---

*📚 Generated with [Claude Code](https://claude.ai/code)*

*Co-Authored-By: Winston (Architect) <architect@kubechat.dev>*